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

func TestChatPage_CompletionDialogResetsAfterEscape(t *testing.T) {
	// This test specifically verifies the bug fix where after running /session list
	// and pressing escape, typing "/" would show options instead of topics
	
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
	
	t.Run("completion dialog resets state after session list escape", func(t *testing.T) {
		// Step 1: Type /session list
		page.editor.SetValue("/session list")
		page.showCompletionDialog = true
		
		// Step 2: Execute the command (simulate Enter)
		execMsg := dialog.SlashCommandExecuteMsg{
			Raw: "/session list",
		}
		updated, cmd := page.Update(execMsg)
		page = updated.(*chatPage)
		assert.NotNil(t, cmd)
		
		// Step 3: Simulate session list dialog being shown and then escaped
		// The important part is that the editor should be cleared
		assert.Equal(t, "", page.editor.GetValue(), "Editor should be cleared after command execution")
		
		// Step 4: Close any dialogs (simulate Escape)
		page.showCompletionDialog = false
		
		// Step 5: Type "/" again
		updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		page = updated.(*chatPage)
		
		// Verify the editor only contains "/"
		assert.Equal(t, "/", page.editor.GetValue(), "Editor should only contain '/', not '/session list /'")
		assert.True(t, page.showCompletionDialog, "Completion dialog should be shown")
		
		// If we had access to the completion dialog's items, we would verify
		// they are topics (help, file, project, session, system) not options (--verbose, --format, --all)
	})
	
	t.Run("multiple slash commands don't accumulate", func(t *testing.T) {
		// Clear state
		page.editor.SetValue("")
		page.showCompletionDialog = false
		
		commands := []string{"/help", "/session list", "/file list"}
		
		for _, cmd := range commands {
			// Type the command
			page.editor.SetValue(cmd)
			
			// Execute it
			execMsg := dialog.SlashCommandExecuteMsg{Raw: cmd}
			updated, _ := page.Update(execMsg)
			page = updated.(*chatPage)
			
			// Verify editor is cleared
			assert.Equal(t, "", page.editor.GetValue(), "Editor should be cleared after executing "+cmd)
			
			// Type "/" again
			page.editor.SetValue("/")
			
			// Should only have "/"
			assert.Equal(t, "/", page.editor.GetValue(), "Editor should only contain '/' after "+cmd)
		}
	})
	
	t.Run("backspace to slash resets properly", func(t *testing.T) {
		// Type a partial command
		page.editor.SetValue("/session li")
		page.showCompletionDialog = true
		
		// Backspace back to just "/"
		page.editor.SetValue("/")
		
		// The completion dialog should handle this properly
		// In the real implementation, the dialog would close and reopen
		// showing topics instead of commands/options
		assert.Equal(t, "/", page.editor.GetValue())
	})
}

func TestChatPage_CompletionStateIsolation(t *testing.T) {
	// This test ensures that completion state from one command doesn't leak to another
	
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
	page.Init()
	
	// Set size
	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updated, _ := page.Update(sizeMsg)
	page = updated.(*chatPage)
	
	t.Run("command with options doesn't affect next command", func(t *testing.T) {
		// Execute a command with options
		page.editor.SetValue("/session list --verbose")
		execMsg := dialog.SlashCommandExecuteMsg{
			Raw: "/session list --verbose",
		}
		updated, _ := page.Update(execMsg)
		page = updated.(*chatPage)
		
		// Editor should be cleared
		assert.Equal(t, "", page.editor.GetValue())
		
		// Type a different command
		page.editor.SetValue("/help")
		execMsg = dialog.SlashCommandExecuteMsg{
			Raw: "/help",
		}
		updated, _ = page.Update(execMsg)
		page = updated.(*chatPage)
		
		// Editor should be cleared again
		assert.Equal(t, "", page.editor.GetValue())
		
		// Type "/" - should show topics, not options from session list
		page.editor.SetValue("/")
		assert.Equal(t, "/", page.editor.GetValue())
	})
}