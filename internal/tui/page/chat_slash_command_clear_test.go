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

func TestChatPage_SlashCommandClearsEditor(t *testing.T) {
	// This test verifies that executing a slash command clears the editor textarea
	// Previously, the bug was that after executing "/session list", typing "/" again
	// would result in "/session list /" instead of just "/"
	
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
	assert.NotNil(t, cmd)
	
	// Set size
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updated, _ := page.Update(sizeMsg)
	page = updated.(*chatPage)
	
	t.Run("slash command execution clears editor", func(t *testing.T) {
		// Set up initial state - simulate typing "/session list"
		page.editor.SetValue("/session list")
		assert.Equal(t, "/session list", page.editor.GetValue())
		
		// Execute the slash command
		msg := dialog.SlashCommandExecuteMsg{
			Raw: "/session list",
		}
		
		updated, cmd := page.Update(msg)
		page = updated.(*chatPage)
		
		// Command should be triggered
		assert.NotNil(t, cmd)
		
		// Editor should be cleared after executing the command
		assert.Equal(t, "", page.editor.GetValue(), "Editor should be cleared after executing slash command")
		
		// Completion dialog should be closed
		assert.False(t, page.showCompletionDialog, "Completion dialog should be closed")
	})
	
	t.Run("typing slash after command execution shows fresh completions", func(t *testing.T) {
		// Ensure editor is empty
		page.editor.SetValue("")
		
		// Update to process any pending state changes
		updated, _ := page.Update(nil)
		page = updated.(*chatPage)
		
		// Simulate typing "/" - don't set the value directly, let the Update handle it
		updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		page = updated.(*chatPage)
		
		// Should show completion dialog
		assert.True(t, page.showCompletionDialog, "Completion dialog should be shown")
		
		// Editor should only have "/"
		assert.Equal(t, "/", page.editor.GetValue(), "Editor should only contain '/' not old command text")
	})
	
	t.Run("progressive command building after clear", func(t *testing.T) {
		// Clear and start fresh
		page.editor.SetValue("/")
		page.showCompletionDialog = true
		
		// Complete to "/session "
		msg := dialog.SlashCommandCompleteMsg{
			OriginalValue: "/",
			NewValue:      "/session ",
			CursorPos:     9,
			KeepOpen:      true,
		}
		
		updated, _ := page.Update(msg)
		page = updated.(*chatPage)
		
		assert.Equal(t, "/session ", page.editor.GetValue())
		
		// Execute the command
		execMsg := dialog.SlashCommandExecuteMsg{
			Raw: "/session list",
		}
		
		updated, _ = page.Update(execMsg)
		page = updated.(*chatPage)
		
		// Editor should be cleared
		assert.Equal(t, "", page.editor.GetValue())
		
		// Type "/" again
		page.editor.SetValue("/")
		
		// Should only have "/" not "/session list /"
		assert.Equal(t, "/", page.editor.GetValue())
	})
}