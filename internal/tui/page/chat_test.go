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

func TestChatPage_Completion(t *testing.T) {
	// Create a temporary directory and a minimal config file to prevent panics.
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

	// Load the configuration from the test directory.
	_, err = config.Load(tmpDir, false)
	assert.NoError(t, err)

	// Now, create the app.
	db, _ := sql.Open("sqlite3", ":memory:")
	app, err := app.New(context.Background(), db, true)
	assert.NoError(t, err)

	chatPageModel := NewChatPage(app)
	p := chatPageModel.(*chatPage)

	// --- Test Case 1: Initial state ---
	assert.False(t, p.showCompletionDialog, "Dialog should not be visible initially")

	// --- Test Case 2: Typing '/' shows command completion ---
	p.editor.SetValue("/")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, p.showCompletionDialog, "Dialog should be visible after typing '/'")
	assert.Equal(t, "slash-commands", p.completionDialog.GetId(), "Provider should be 'slash-commands'")
	// Verify empty message is correct for commands (this is the key check)
	assert.Equal(t, "No command matches found", p.completionDialog.GetEmptyMessage(), "Empty message should be for commands")
	// Note: We may not have command items in test environment, but the message should be correct

	// --- Test Case 3: Typing '@' shows file completion ---
	p.editor.SetValue("@")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}})
	assert.True(t, p.showCompletionDialog, "Dialog should be visible after typing '@'")
	assert.Equal(t, "files", p.completionDialog.GetId(), "Provider should be 'files'")
	// Verify empty message changed to files
	assert.Equal(t, "No file matches found", p.completionDialog.GetEmptyMessage(), "Empty message should be for files")

	// --- Test Case 4: Typing a normal character hides the dialog ---
	p.editor.SetValue("a")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.False(t, p.showCompletionDialog, "Dialog should be hidden after typing 'a'")

	// --- Test Case 5: Backspace and re-typing '/' should show dialog ---
	// Type "/"
	p.editor.SetValue("/")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, p.showCompletionDialog, "Dialog should be visible after typing '/'")
	
	// Backspace to clear
	p.editor.SetValue("")
	p.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.False(t, p.showCompletionDialog, "Dialog should be hidden after backspace")
	
	// Type "/" again - this is the scenario from the bug report
	p.editor.SetValue("/")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, p.showCompletionDialog, "Dialog should be visible after re-typing '/'")
	assert.Equal(t, "slash-commands", p.completionDialog.GetId(), "Provider should be 'slash-commands'")
	// THIS CHECK WOULD HAVE CAUGHT THE BUG - verify the empty message is correct
	assert.Equal(t, "No command matches found", p.completionDialog.GetEmptyMessage(), "Empty message should be for commands after re-typing '/'")
	// The bug was that it showed "No file matches found" instead of "No commands found"

	// --- Test Case 6: Multiple backspaces and re-typing should still work ---
	// Type "/list"
	p.editor.SetValue("/list")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	assert.True(t, p.showCompletionDialog, "Dialog should remain visible while typing command")
	
	// Backspace everything
	p.editor.SetValue("")
	p.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.False(t, p.showCompletionDialog, "Dialog should be hidden after clearing all text")
	
	// Type "@" this time - testing provider switch
	p.editor.SetValue("@")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}})
	assert.True(t, p.showCompletionDialog, "Dialog should be visible after typing '@'")
	assert.Equal(t, "files", p.completionDialog.GetId(), "Provider should be 'files'")
	// Verify the empty message switched from commands to files
	assert.Equal(t, "No file matches found", p.completionDialog.GetEmptyMessage(), "Empty message should switch to files after typing '@'")
}

func TestChatPage_CompletionViewOutput(t *testing.T) {
	// Create a temporary directory and a minimal config file to prevent panics.
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

	// Load the configuration from the test directory.
	_, err = config.Load(tmpDir, false)
	assert.NoError(t, err)

	// Now, create the app.
	db, _ := sql.Open("sqlite3", ":memory:")
	app, err := app.New(context.Background(), db, true)
	assert.NoError(t, err)

	chatPageModel := NewChatPage(app)
	p := chatPageModel.(*chatPage)

	// Test View() output when no items match
	p.editor.SetValue("/nonexistentcommand")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	
	// Update the dialog to search for a non-existent command
	dialogModel, _ := p.completionDialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	p.completionDialog = dialogModel.(dialog.CompletionDialog)
	
	// The View() should contain the empty message
	view := p.completionDialog.View()
	assert.Contains(t, view, "No command matches found", "View should display 'No command matches found' when no commands match")
	
	// Switch to file completion with no matches
	p.editor.SetValue("@nonexistentfile")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}})
	
	dialogModel, _ = p.completionDialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	p.completionDialog = dialogModel.(dialog.CompletionDialog)
	
	view = p.completionDialog.View()
	assert.Contains(t, view, "No file matches found", "View should display 'No file matches found' when no files match")
}