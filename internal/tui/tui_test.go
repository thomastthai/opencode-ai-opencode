package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opencode-ai/opencode/internal/app"
	"github.com/opencode-ai/opencode/internal/commands"
	"github.com/opencode-ai/opencode/internal/tui/command"
	"github.com/opencode-ai/opencode/internal/tui/components/dialog"
)

func TestTUIModelCreation(t *testing.T) {
	// Create a test app (simplified)
	testApp := &app.App{
		// Add minimal required fields for testing
	}
	
	// Create a new TUI model  
	model := New(testApp)
	if model == nil {
		t.Fatal("Expected model to be created, got nil")
	}
	
	// Test that the model implements tea.Model interface
	_, ok := model.(tea.Model)
	if !ok {
		t.Error("Expected model to implement tea.Model interface")
	}
}

func TestCommandSelectedMsgHandling(t *testing.T) {
	testApp := &app.App{}
	model := New(testApp)
	
	// Test built-in command with handler
	t.Run("built-in command with handler", func(t *testing.T) {
		cmdWithHandler := command.Command{
			ID:    "test-builtin",
			Title: "Test Built-in",
			Scope: command.BuiltinScope,
			Handler: func(cmd command.Command) tea.Cmd {
				return func() tea.Msg {
					return "handler_executed"
				}
			},
		}
		
		msg := dialog.CommandSelectedMsg{
			Command: dialog.Command(cmdWithHandler),
		}
		
		_, cmd := model.Update(msg)
		if cmd == nil {
			t.Error("Expected command to be returned for built-in command with handler")
		}
	})
	
	// Test custom command without placeholders
	t.Run("custom command without placeholders", func(t *testing.T) {
		customCmd := command.Command{
			ID:      "test-custom",
			Title:   "Test Custom",
			Content: "echo 'hello world'",
			Scope:   command.UserScope,
		}
		
		msg := dialog.CommandSelectedMsg{
			Command: dialog.Command(customCmd),
		}
		
		_, cmd := model.Update(msg)
		if cmd == nil {
			t.Error("Expected command to be returned for custom command")
		}
	})
	
	// Test dialog close
	t.Run("close command dialog", func(t *testing.T) {
		msg := dialog.CloseCommandDialogMsg{}
		updatedModel, _ := model.Update(msg)
		
		if updatedModel == nil {
			t.Error("Expected updated model to be returned")
		}
	})
}

func TestKeyboardShortcuts(t *testing.T) {
	testApp := &app.App{}
	model := New(testApp)
	
	// Test command dialog shortcut (Ctrl+K)
	t.Run("command dialog shortcut", func(t *testing.T) {
		msg := tea.KeyMsg{
			Type:  tea.KeyCtrlK,
			Runes: []rune{'k'},
		}
		
		updatedModel, _ := model.Update(msg)
		if updatedModel == nil {
			t.Error("Expected updated model to be returned")
		}
	})
	
	// Test help dialog shortcut (Ctrl+H)  
	t.Run("help dialog shortcut", func(t *testing.T) {
		msg := tea.KeyMsg{
			Type:  tea.KeyCtrlH,
			Runes: []rune{'h'},
		}
		
		updatedModel, _ := model.Update(msg)
		if updatedModel == nil {
			t.Error("Expected updated model to be returned")
		}
	})
	
	// Test quit shortcut (Ctrl+C)
	t.Run("quit shortcut", func(t *testing.T) {
		msg := tea.KeyMsg{
			Type:  tea.KeyCtrlC,
			Runes: []rune{'c'},
		}
		
		_, cmd := model.Update(msg)
		if cmd == nil {
			t.Error("Expected quit command after Ctrl+C")
		}
	})
}

func TestConvertRegistryCommand(t *testing.T) {
	// Create a test registry command
	regCmd := commands.NewCommand("test-registry", "Test Registry", "A test registry command").
		WithType(commands.BuiltinCommand).
		WithCategory("testing").
		WithAliases([]string{"tr", "test"}).
		Build()
	
	// Convert to TUI command
	tuiCmd := convertRegistryCommand(regCmd)
	
	// Test conversion
	t.Run("convert registry command", func(t *testing.T) {
		if tuiCmd.ID != "test-registry" {
			t.Errorf("Expected ID 'test-registry', got '%s'", tuiCmd.ID)
		}
		if tuiCmd.Title != "Test Registry" {
			t.Errorf("Expected Title 'Test Registry', got '%s'", tuiCmd.Title)
		}
		if tuiCmd.Description != "A test registry command" {
			t.Errorf("Expected Description 'A test registry command', got '%s'", tuiCmd.Description)
		}
		if tuiCmd.Scope != command.BuiltinScope {
			t.Errorf("Expected Scope to be BuiltinScope, got %v", tuiCmd.Scope)
		}
		if tuiCmd.Category != "testing" {
			t.Errorf("Expected Category 'testing', got '%s'", tuiCmd.Category)
		}
		if len(tuiCmd.Aliases) != 2 {
			t.Errorf("Expected 2 aliases, got %d", len(tuiCmd.Aliases))
		}
		if tuiCmd.Handler == nil {
			t.Error("Expected handler to be created for built-in command")
		}
	})
}

func TestDetermineCommandScope(t *testing.T) {
	testCases := []struct {
		commandID     string
		expectedScope command.CommandScope
	}{
		{"user:test", command.UserScope},
		{"project:deploy", command.ProjectScope},
		{"plain-command", command.UserScope}, // Default
		{"", command.UserScope},              // Default for empty
		{"no-prefix-command", command.UserScope}, // Default
	}
	
	for _, tc := range testCases {
		t.Run(tc.commandID, func(t *testing.T) {
			scope := determineCommandScope(tc.commandID)
			if scope != tc.expectedScope {
				t.Errorf("Expected scope %v for command ID '%s', got %v", tc.expectedScope, tc.commandID, scope)
			}
		})
	}
}


func TestWindowSizeMsg(t *testing.T) {
	testApp := &app.App{}
	model := New(testApp)
	
	// Test window size message handling
	t.Run("window size message", func(t *testing.T) {
		msg := tea.WindowSizeMsg{
			Width:  100,
			Height: 50,
		}
		
		updatedModel, cmd := model.Update(msg)
		if updatedModel == nil {
			t.Error("Expected updated model to be returned")
		}
		// cmd can be nil or not, both are valid
		_ = cmd
	})
}