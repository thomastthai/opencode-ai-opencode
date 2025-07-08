package page

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opencode-ai/opencode/internal/app"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/tui/components/dialog"
	"github.com/stretchr/testify/assert"
)

func TestChatPage_TabCompletionRaceCondition(t *testing.T) {
	// This test verifies that the race condition between tab completion and enter key is handled
	// The bug was: /sys<tab> -> /system, then h<tab> quickly followed by enter would execute "/system h" instead of "/system help"
	
	tmpDir, err := os.MkdirTemp("", "opencode-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configContent := `{
		"agents": {
			"coder": { "model": "test-model" }
		},
		"mcpServers": {}
	}`
	configPath := filepath.Join(tmpDir, ".opencode.json")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	_, err = config.Load(tmpDir, false)
	assert.NoError(t, err)

	db, _ := sql.Open("sqlite3", ":memory:")
	testApp, err := app.New(context.Background(), db, true)
	assert.NoError(t, err)
	
	// Create chat page
	model := NewChatPage(testApp)
	page := model.(*chatPage)
	
	// Initialize
	cmd := page.Init()
	if cmd != nil {
		cmd()
	}
	
	// Set size
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updated, _ := page.Update(sizeMsg)
	page = updated.(*chatPage)
	
	t.Run("tab completion sets pending state", func(t *testing.T) {
		// Type /sys
		for _, r := range "/sys" {
			updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			page = updated.(*chatPage)
		}
		
		assert.Equal(t, "/sys", page.editor.GetValue())
		assert.True(t, page.showCompletionDialog)
		assert.Equal(t, CompletionStateIdle, page.completionState)
		
		// Simulate tab completion starting
		startMsg := dialog.TabCompletionStartedMsg{
			CurrentValue: "/sys",
			TargetValue:  "/system ",
		}
		updated, _ = page.Update(startMsg)
		page = updated.(*chatPage)
		
		// Should be in pending state
		assert.Equal(t, CompletionStatePending, page.completionState)
		assert.Equal(t, "/system ", page.pendingCompletion)
	})
	
	t.Run("enter key blocked during pending completion", func(t *testing.T) {
		// Reset state
		page.editor.SetValue("")
		page.completionState = CompletionStateIdle
		page.pendingCompletion = ""
		page.showCompletionDialog = false
		
		// Type /system h
		page.editor.SetValue("/system h")
		page.showCompletionDialog = true
		
		// Simulate tab completion starting
		startMsg := dialog.TabCompletionStartedMsg{
			CurrentValue: "/system h",
			TargetValue:  "/system help ",
		}
		updated, _ = page.Update(startMsg)
		page = updated.(*chatPage)
		
		// Now simulate enter key (which would normally execute the command)
		execMsg := dialog.SlashCommandExecuteMsg{
			Raw: "/system h",
		}
		
		// Update with the execute message
		updated, cmd := page.Update(execMsg)
		page = updated.(*chatPage)
		
		// Command should NOT be executed (no command returned)
		// The editor should NOT be cleared
		assert.Equal(t, "/system h", page.editor.GetValue(), "Editor should not be cleared during pending completion")
		assert.Equal(t, CompletionStatePending, page.completionState, "Should still be pending")
		
		// Now complete the tab completion
		completeMsg := dialog.SlashCommandCompleteMsg{
			OriginalValue: "/system h",
			NewValue:      "/system help ",
			CursorPos:     13,
			KeepOpen:      true,
		}
		
		updated, _ = page.Update(completeMsg)
		page = updated.(*chatPage)
		
		// Should be back to idle
		assert.Equal(t, CompletionStateIdle, page.completionState)
		assert.Equal(t, "", page.pendingCompletion)
		
		// Now enter should work
		execMsg2 := dialog.SlashCommandExecuteMsg{
			Raw: "/system help",
		}
		
		updated, cmd = page.Update(execMsg2)
		page = updated.(*chatPage)
		
		// Command should be executed this time
		assert.NotNil(t, cmd, "Command should be executed after completion")
		assert.Equal(t, "", page.editor.GetValue(), "Editor should be cleared after execution")
	})
	
	t.Run("typing different text cancels pending completion", func(t *testing.T) {
		// Reset state
		page.editor.SetValue("/system h")
		page.showCompletionDialog = true
		
		// Start tab completion
		startMsg := dialog.TabCompletionStartedMsg{
			CurrentValue: "/system h",
			TargetValue:  "/system help ",
		}
		updated, _ = page.Update(startMsg)
		page = updated.(*chatPage)
		
		assert.Equal(t, CompletionStatePending, page.completionState)
		
		// User types something different
		page.editor.SetValue("/session")
		
		// Update the page (this would normally happen through layout update)
		updated, _ = page.Update(nil)
		page = updated.(*chatPage)
		
		// Pending completion should be canceled
		assert.Equal(t, CompletionStateIdle, page.completionState)
		assert.Equal(t, "", page.pendingCompletion)
	})
	
	t.Run("completion completes successfully without race", func(t *testing.T) {
		// Reset state
		page.editor.SetValue("")
		page.completionState = CompletionStateIdle
		page.showCompletionDialog = false
		
		// Type /sys
		for _, r := range "/sys" {
			updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			page = updated.(*chatPage)
		}
		
		// Tab completion flow
		// 1. Start
		startMsg := dialog.TabCompletionStartedMsg{
			CurrentValue: "/sys",
			TargetValue:  "/system ",
		}
		updated, _ = page.Update(startMsg)
		page = updated.(*chatPage)
		
		// 2. Complete
		completeMsg := dialog.SlashCommandCompleteMsg{
			OriginalValue: "/sys",
			NewValue:      "/system ",
			CursorPos:     8,
			KeepOpen:      true,
		}
		updated, _ = page.Update(completeMsg)
		page = updated.(*chatPage)
		
		// Should have completed value
		assert.Equal(t, "/system ", page.editor.GetValue())
		assert.Equal(t, CompletionStateIdle, page.completionState)
		
		// Type h
		page.editor.SetValue("/system h")
		
		// Tab again
		startMsg2 := dialog.TabCompletionStartedMsg{
			CurrentValue: "/system h",
			TargetValue:  "/system help ",
		}
		updated, _ = page.Update(startMsg2)
		page = updated.(*chatPage)
		
		completeMsg2 := dialog.SlashCommandCompleteMsg{
			OriginalValue: "/system h",
			NewValue:      "/system help ",
			CursorPos:     13,
			KeepOpen:      false,
		}
		updated, _ = page.Update(completeMsg2)
		page = updated.(*chatPage)
		
		assert.Equal(t, "/system help ", page.editor.GetValue())
		assert.Equal(t, CompletionStateIdle, page.completionState)
	})
}

func TestChatPage_TabCompletionStateTransitions(t *testing.T) {
	// Test various state transitions to ensure robustness
	
	tmpDir, err := os.MkdirTemp("", "opencode-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configContent := `{
		"agents": {
			"coder": { "model": "test-model" }
		},
		"mcpServers": {}
	}`
	configPath := filepath.Join(tmpDir, ".opencode.json")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	_, err = config.Load(tmpDir, false)
	assert.NoError(t, err)

	db, _ := sql.Open("sqlite3", ":memory:")
	testApp, err := app.New(context.Background(), db, true)
	assert.NoError(t, err)
	
	model := NewChatPage(testApp)
	page := model.(*chatPage)
	page.Init()
	
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updated, _ := page.Update(sizeMsg)
	page = updated.(*chatPage)
	
	t.Run("dialog close resets completion state", func(t *testing.T) {
		// Set pending state
		page.completionState = CompletionStatePending
		page.pendingCompletion = "/system help "
		page.showCompletionDialog = true
		
		// Close dialog
		closeMsg := dialog.CompletionDialogCloseMsg{}
		updated, _ := page.Update(closeMsg)
		page = updated.(*chatPage)
		
		// State should be reset
		assert.Equal(t, CompletionStateIdle, page.completionState)
		assert.Equal(t, "", page.pendingCompletion)
		assert.False(t, page.showCompletionDialog)
	})
	
	t.Run("rapid tab presses handled correctly", func(t *testing.T) {
		// Simulate rapid tab presses
		page.editor.SetValue("/ses")
		
		// First tab
		startMsg1 := dialog.TabCompletionStartedMsg{
			CurrentValue: "/ses",
			TargetValue:  "/session ",
		}
		updated, _ = page.Update(startMsg1)
		page = updated.(*chatPage)
		
		// Second tab before first completes
		startMsg2 := dialog.TabCompletionStartedMsg{
			CurrentValue: "/session n",
			TargetValue:  "/session new ",
		}
		updated, _ = page.Update(startMsg2)
		page = updated.(*chatPage)
		
		// Should track the latest
		assert.Equal(t, "/session new ", page.pendingCompletion)
		
		// Complete the second one
		completeMsg := dialog.SlashCommandCompleteMsg{
			OriginalValue: "/session n",
			NewValue:      "/session new ",
			CursorPos:     13,
			KeepOpen:      false,
		}
		updated, _ = page.Update(completeMsg)
		page = updated.(*chatPage)
		
		assert.Equal(t, CompletionStateIdle, page.completionState)
	})
}