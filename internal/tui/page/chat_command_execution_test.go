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

func TestChatPage_SlashCommandExecution(t *testing.T) {
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
	
	t.Run("execute session clear command", func(t *testing.T) {
		// Execute /session clear
		msg := dialog.SlashCommandExecuteMsg{
			Raw: "/session clear",
		}
		
		updated, cmd := page.Update(msg)
		page = updated.(*chatPage)
		
		// Should trigger session clear
		assert.NotNil(t, cmd)
		
		// Process the command
		result := cmd()
		
		// Should receive SessionClearRequestedMsg
		assert.NotNil(t, result)
	})
	
	t.Run("execute project init command", func(t *testing.T) {
		// Execute /project init
		msg := dialog.SlashCommandExecuteMsg{
			Raw: "/project init",
		}
		
		updated, cmd := page.Update(msg)
		page = updated.(*chatPage)
		
		// Should trigger command execution
		assert.NotNil(t, cmd)
		
		// Process the command
		result := cmd()
		assert.NotNil(t, result)
	})
	
	t.Run("execute help command", func(t *testing.T) {
		// Execute /help
		msg := dialog.SlashCommandExecuteMsg{
			Raw: "/help",
		}
		
		updated, cmd := page.Update(msg)
		page = updated.(*chatPage)
		
		// Should trigger help
		assert.NotNil(t, cmd)
	})
	
	t.Run("execute auth login with argument", func(t *testing.T) {
		// Execute /auth login gemini
		msg := dialog.SlashCommandExecuteMsg{
			Raw: "/auth login gemini",
		}
		
		updated, cmd := page.Update(msg)
		page = updated.(*chatPage)
		
		// Should trigger auth login
		assert.NotNil(t, cmd)
		
		// Process the command
		result := cmd()
		assert.NotNil(t, result)
	})
	
	t.Run("execute invalid command", func(t *testing.T) {
		// Execute invalid command
		msg := dialog.SlashCommandExecuteMsg{
			Raw: "/nonexistent command",
		}
		
		updated, cmd := page.Update(msg)
		page = updated.(*chatPage)
		
		// Should still return a command (for error reporting)
		assert.NotNil(t, cmd)
		
		// Process the command
		result := cmd()
		
		// Should be a warning/error message
		assert.NotNil(t, result)
	})
}