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

func TestChatPage_CompletionProviderStability(t *testing.T) {
	// This test ensures that SetProvider is not called unnecessarily on every update
	// which was causing the continuous scrolling bug
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

	// Set up with "/" to trigger command completion
	p.editor.SetValue("/")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.True(t, p.showCompletionDialog)
	assert.Equal(t, "slash-commands", p.completionDialog.GetId())

	// Simulate multiple updates (like what happens during continuous scrolling)
	// This should NOT cause the provider to be reset each time
	for i := 0; i < 10; i++ {
		p.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		// Verify that the provider hasn't changed
		assert.Equal(t, "slash-commands", p.completionDialog.GetId(), "Provider should not change on window resize")
		assert.True(t, p.showCompletionDialog, "Dialog should remain visible")
	}

	// Type more characters and ensure provider remains stable
	p.editor.SetValue("/list")
	for i := 0; i < 5; i++ {
		// Simulate various update events that shouldn't affect provider
		p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}) // Random update
		p.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		p.Update(tea.MouseMsg{Type: tea.MouseMotion})
		assert.Equal(t, "slash-commands", p.completionDialog.GetId(), "Provider should remain 'slash-commands'")
		assert.True(t, p.showCompletionDialog, "Dialog should remain visible")
	}

	// Now test switching providers - this SHOULD change the provider
	p.editor.SetValue("@")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'@'}})
	assert.Equal(t, "files", p.completionDialog.GetId(), "Provider should change to 'files' when @ is typed")

	// Multiple updates with @ should not reset the provider
	for i := 0; i < 5; i++ {
		p.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		assert.Equal(t, "files", p.completionDialog.GetId(), "Provider should remain 'files'")
	}
}

func TestChatPage_NoInfiniteLoopOnContinuousTyping(t *testing.T) {
	// This test simulates the reported bug where typing / followed by characters
	// causes continuous scrolling
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

	// Type "/" 
	p.editor.SetValue("/")
	p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	initialProvider := p.completionDialog.GetId()
	assert.Equal(t, "slash-commands", initialProvider)

	// Type "list" character by character, simulating real typing
	chars := []rune{'l', 'i', 's', 't'}
	fullText := "/"
	for _, char := range chars {
		fullText += string(char)
		p.editor.SetValue(fullText)
		
		// Multiple updates per character (simulating the continuous loop)
		for j := 0; j < 3; j++ {
			p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
			// Provider should remain the same, not be reset
			assert.Equal(t, "slash-commands", p.completionDialog.GetId(), 
				"Provider should not change while typing after /")
		}
	}

	// Verify dialog is still showing and provider hasn't changed
	assert.True(t, p.showCompletionDialog, "Dialog should remain visible")
	assert.Equal(t, "slash-commands", p.completionDialog.GetId(), "Provider should still be slash-commands")
}