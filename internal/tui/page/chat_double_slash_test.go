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
	"github.com/stretchr/testify/assert"
)

func TestChatPage_PreventDoubleSlashAfterBackspace(t *testing.T) {
	// This test specifically covers the bug fixed in commit 8f2e81f
	// where typing /session<tab>, backspacing everything, then typing /se<tab>
	// would result in //se instead of /session
	
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
	
	t.Run("backspace to slash then type slash again should not create double slash", func(t *testing.T) {
		// Step 1: Type /session
		for _, r := range "/session" {
			updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			page = updated.(*chatPage)
		}
		
		assert.Equal(t, "/session", page.editor.GetValue())
		assert.True(t, page.showCompletionDialog)
		
		// Step 2: Backspace everything (8 times for "/session")
		for i := 0; i < 8; i++ {
			updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyBackspace})
			page = updated.(*chatPage)
		}
		
		// Editor should be empty
		assert.Equal(t, "", page.editor.GetValue())
		
		// Step 3: Type / again
		updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		page = updated.(*chatPage)
		
		// Should only have one slash, not //
		assert.Equal(t, "/", page.editor.GetValue(), "Should not create double slash")
		assert.True(t, page.showCompletionDialog)
	})
	
	t.Run("backspace to just slash closes and reopens dialog correctly", func(t *testing.T) {
		// Clear state
		page.editor.SetValue("")
		page.showCompletionDialog = false
		
		// Type /se
		for _, r := range "/se" {
			updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			page = updated.(*chatPage)
		}
		
		assert.Equal(t, "/se", page.editor.GetValue())
		assert.True(t, page.showCompletionDialog)
		
		// Backspace once to get /s
		updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		page = updated.(*chatPage)
		assert.Equal(t, "/s", page.editor.GetValue())
		assert.True(t, page.showCompletionDialog, "Dialog should remain open with /s")
		
		// Backspace again to get just /
		updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		page = updated.(*chatPage)
		assert.Equal(t, "/", page.editor.GetValue())
		
		// The key behavior: dialog should handle the single slash state correctly
		// When we type another character, it should not duplicate the slash
		updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
		page = updated.(*chatPage)
		assert.Equal(t, "/s", page.editor.GetValue(), "Should be /s not //s")
	})
	
	t.Run("tab completion after backspace to slash", func(t *testing.T) {
		// Clear state
		page.editor.SetValue("")
		page.showCompletionDialog = false
		
		// Type /session
		for _, r := range "/session" {
			updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			page = updated.(*chatPage)
		}
		
		// Backspace to just /
		for i := 0; i < 7; i++ { // Remove "session"
			updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyBackspace})
			page = updated.(*chatPage)
		}
		
		assert.Equal(t, "/", page.editor.GetValue())
		
		// Type se
		for _, r := range "se" {
			updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			page = updated.(*chatPage)
		}
		
		assert.Equal(t, "/se", page.editor.GetValue(), "Should be /se not //se")
		
		// Tab completion should work correctly
		updated, _ = page.Update(tea.KeyMsg{Type: tea.KeyTab})
		page = updated.(*chatPage)
		
		// The exact completion depends on what commands start with "se"
		// but it should not have a double slash
		editorValue := page.editor.GetValue()
		assert.NotContains(t, editorValue, "//", "Tab completion should not create double slash")
	})
}