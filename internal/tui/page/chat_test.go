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
	assert.Equal(t, "commands", p.completionDialog.GetId(), "Provider should be 'commands'")

	// --- Test Case 3: Typing '@' shows file completion ---
	p.editor.SetValue("@")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}})
	assert.True(t, p.showCompletionDialog, "Dialog should be visible after typing '@'")
	assert.Equal(t, "files", p.completionDialog.GetId(), "Provider should be 'files'")

	// --- Test Case 4: Typing a normal character hides the dialog ---
	p.editor.SetValue("a")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.False(t, p.showCompletionDialog, "Dialog should be hidden after typing 'a'")
}