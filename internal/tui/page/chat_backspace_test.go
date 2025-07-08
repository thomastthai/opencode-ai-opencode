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

func TestChatPage_BackspaceToSlashShowsCommands(t *testing.T) {
	// This test reproduces the bug where backspacing from /list to / shows "No commands found"
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
	app, err := app.New(context.Background(), db, true)
	assert.NoError(t, err)

	chatPageModel := NewChatPage(app)
	p := chatPageModel.(*chatPage)
	
	// Initialize the page
	initCmd := p.Init()
	if initCmd != nil {
		// Execute initialization
		initCmd()
	}

	// Step 1: Type "/list" - simulate each character separately
	// Type "/"
	updated, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	p = updated.(*chatPage)
	
	// Type "l"
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	p = updated.(*chatPage)
	
	// Type "i"
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	p = updated.(*chatPage)
	
	// Type "s"
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	p = updated.(*chatPage)
	
	// Type "t"
	updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	p = updated.(*chatPage)
	
	assert.True(t, p.showCompletionDialog, "Dialog should be visible with /list")
	assert.Equal(t, "slash-commands", p.completionDialog.GetId())

	// Step 2: Backspace to just "/" - simulate backspaces through the page update
	// The editor should currently have "/list"
	assert.Equal(t, "/list", p.editor.GetValue())
	
	// Backspace 4 times to remove "list"
	for i := 0; i < 4; i++ {
		updated, _ = p.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		p = updated.(*chatPage)
	}
	
	// Editor should now have just "/"
	assert.Equal(t, "/", p.editor.GetValue())

	// Step 3: Verify dialog is still showing and has commands
	assert.True(t, p.showCompletionDialog, "Dialog should still be visible with just /")
	assert.Equal(t, "slash-commands", p.completionDialog.GetId())
	
	// The key check: verify commands are shown, not "No commands found"
	items := p.completionDialog.GetListItems()
	assert.Greater(t, len(items), 0, "Should have command items when query is just /")
	
	// Also check the view doesn't show "No commands found"
	view := p.completionDialog.View()
	assert.NotContains(t, view, "No command matches found", "Should not show 'No command matches found' when / is typed")
}

func TestChatPage_BackspaceScenarios(t *testing.T) {
	// Test various backspace scenarios
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
	app, err := app.New(context.Background(), db, true)
	assert.NoError(t, err)

	t.Run("backspace from /l to /", func(t *testing.T) {
		chatPageModel := NewChatPage(app)
		p := chatPageModel.(*chatPage)
		
		// Initialize the page
		initCmd := p.Init()
		if initCmd != nil {
			initCmd()
		}

		// Type /l
		p.editor.SetValue("/l")
		p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})

		// Backspace to /
		p.editor.SetValue("/")
		p.Update(tea.KeyMsg{Type: tea.KeyBackspace})

		items := p.completionDialog.GetListItems()
		assert.Greater(t, len(items), 0, "Should show all commands after backspacing from /l to /")
	})

	t.Run("type / then immediately check", func(t *testing.T) {
		chatPageModel := NewChatPage(app)
		p := chatPageModel.(*chatPage)
		
		// Initialize the page
		initCmd := p.Init()
		if initCmd != nil {
			initCmd()
		}

		// Just type /
		p.editor.SetValue("/")
		p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

		items := p.completionDialog.GetListItems()
		assert.Greater(t, len(items), 0, "Should show commands immediately after typing /")
	})
}