package page

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opencode-ai/opencode/internal/app"
	"github.com/opencode-ai/opencode/internal/completions"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/tui/components/dialog"
	"github.com/stretchr/testify/assert"
)

func TestChatPage_SlashCommandCompletion(t *testing.T) {
	// Set up test environment
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
	
	t.Run("slash triggers completion dialog", func(t *testing.T) {
		// Type "/" in editor
		page.editor.SetValue("/")
		editorValue := page.editor.GetValue()
		assert.Equal(t, "/", editorValue)
		
		// Trigger update to process the change
		updated, _ := page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		page = updated.(*chatPage)
		
		// Update should show completion dialog
		assert.True(t, page.showCompletionDialog, "Completion dialog should be shown")
		assert.Equal(t, "slash-commands", page.completionDialog.GetId())
	})
	
	t.Run("progressive command building", func(t *testing.T) {
		// Start with "/"
		page.editor.SetValue("/")
		page.showCompletionDialog = true
		
		// Simulate selecting "session" from completions
		msg := dialog.SlashCommandCompleteMsg{
			OriginalValue: "/",
			NewValue:      "/session ",
			CursorPos:     9,
			KeepOpen:      true,
		}
		
		updated, _ := page.Update(msg)
		page = updated.(*chatPage)
		
		// Check editor was updated
		assert.Equal(t, "/session ", page.editor.GetValue())
		
		// Dialog should still be open
		assert.True(t, page.showCompletionDialog, "Dialog should remain open")
		
		// Simulate selecting "new" verb
		msg2 := dialog.SlashCommandCompleteMsg{
			OriginalValue: "/session ",
			NewValue:      "/session new ",
			CursorPos:     13,
			KeepOpen:      false, // Close after verb selection
		}
		
		updated, _ = page.Update(msg2)
		page = updated.(*chatPage)
		
		// Check final state
		assert.Equal(t, "/session new ", page.editor.GetValue())
		assert.False(t, page.showCompletionDialog, "Dialog should be closed")
	})
	
	t.Run("backspace handling", func(t *testing.T) {
		// Set up with partial command
		page.editor.SetValue("/sess")
		page.showCompletionDialog = true
		
		// Type more
		page.editor.SetValue("/session")
		
		// Still should show dialog
		assert.True(t, page.showCompletionDialog)
		
		// Backspace to "/"
		page.editor.SetValue("/")
		
		// Should still show dialog with slash
		assert.True(t, page.showCompletionDialog)
	})
}

func TestSlashCommandProvider_Integration(t *testing.T) {
	provider := completions.NewSlashCommandProvider()
	
	t.Run("provider registration", func(t *testing.T) {
		assert.Equal(t, "slash-commands", provider.GetId())
		
		entry := provider.GetEntry()
		assert.Equal(t, "/", entry.GetValue())
	})
	
	t.Run("progressive completions", func(t *testing.T) {
		// Test topic completions
		items, err := provider.GetChildEntries("/")
		assert.NoError(t, err)
		assert.Greater(t, len(items), 0)
		
		// Find session item
		var sessionItem dialog.CompletionItemI
		for _, item := range items {
			if item.GetValue() == "/session " {
				sessionItem = item
				break
			}
		}
		assert.NotNil(t, sessionItem)
		
		// Test verb completions
		items, err = provider.GetChildEntries("/session ")
		assert.NoError(t, err)
		assert.Greater(t, len(items), 0)
		
		// All verb completions should start with "/session "
		for _, item := range items {
			value := item.GetValue()
			assert.Contains(t, value, "/session ")
		}
	})
}