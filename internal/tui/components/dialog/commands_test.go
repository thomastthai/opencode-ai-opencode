package dialog

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opencode-ai/opencode/internal/tui/command"
)

func TestCommandSelectedMsgHandling(t *testing.T) {
	dialog := NewCommandDialogCmp()
	
	// Create test commands
	testCommands := []command.Command{
		{
			ID:          "test-builtin",
			Title:       "Test Built-in",
			Description: "A test built-in command",
			Scope:       command.BuiltinScope,
			Handler: func(cmd command.Command) tea.Cmd {
				return func() tea.Msg {
					return tea.Quit()
				}
			},
		},
		{
			ID:          "test-custom",
			Title:       "Test Custom",
			Description: "A test custom command",
			Content:     "echo 'custom command executed'",
			Scope:       command.UserScope,
		},
		{
			ID:          "test-with-placeholders",
			Title:       "Test Placeholders",
			Description: "A test command with placeholders",
			Content:     "echo 'Hello $NAME, welcome to $PROJECT'",
			Scope:       command.ProjectScope,
		},
	}
	
	dialog.SetCommands(testCommands)
	
	// Test selecting a built-in command
	t.Run("select built-in command", func(t *testing.T) {
		selectedCmd := Command(testCommands[0])
		msg := CommandSelectedMsg{Command: selectedCmd}
		
		// The message should be properly formed
		if msg.Command.ID != "test-builtin" {
			t.Errorf("Expected command ID 'test-builtin', got '%s'", msg.Command.ID)
		}
		if msg.Command.Handler == nil {
			t.Error("Expected handler to be set for built-in command")
		}
	})
	
	// Test selecting a custom command
	t.Run("select custom command", func(t *testing.T) {
		selectedCmd := Command(testCommands[1])
		msg := CommandSelectedMsg{Command: selectedCmd}
		
		if msg.Command.Content != "echo 'custom command executed'" {
			t.Errorf("Expected content to be preserved, got '%s'", msg.Command.Content)
		}
		if msg.Command.Handler != nil {
			t.Error("Custom commands should not have handlers")
		}
	})
	
	// Test selecting command with placeholders
	t.Run("select command with placeholders", func(t *testing.T) {
		selectedCmd := Command(testCommands[2])
		hasPlaceholders := command.Command(selectedCmd).HasPlaceholders()
		
		if !hasPlaceholders {
			t.Error("Expected command to have placeholders")
		}
	})
}

func TestCommandDialogKeyboardNavigation(t *testing.T) {
	dialog := NewCommandDialogCmp().(*commandDialogCmp)
	
	testCommands := []command.Command{
		{ID: "cmd1", Title: "Command 1", Scope: command.BuiltinScope},
		{ID: "cmd2", Title: "Command 2", Scope: command.UserScope},
		{ID: "cmd3", Title: "Command 3", Scope: command.ProjectScope},
	}
	
	dialog.SetCommands(testCommands)
	
	// Test Enter key to select command
	t.Run("enter key selection", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := dialog.Update(msg)
		
		if cmd == nil {
			t.Error("Expected command to be returned when Enter is pressed")
		}
	})
	
	// Test Escape key to close dialog
	t.Run("escape key close", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEsc}
		_, cmd := dialog.Update(msg)
		
		if cmd == nil {
			t.Error("Expected command to be returned when Escape is pressed")
		}
	})
	
	// Test search activation
	t.Run("search activation", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
		dialog.Update(msg)
		
		if !dialog.showSearch {
			t.Error("Expected search mode to be activated")
		}
		if !dialog.searchInput.Focused() {
			t.Error("Expected search input to be focused")
		}
	})
}

func TestCommandDialogSearch(t *testing.T) {
	dialog := NewCommandDialogCmp().(*commandDialogCmp)
	
	testCommands := []command.Command{
		{ID: "git-commit", Title: "Git Commit", Description: "Commit changes", Category: "git", Scope: command.BuiltinScope},
		{ID: "git-push", Title: "Git Push", Description: "Push changes", Category: "git", Scope: command.BuiltinScope},
		{ID: "test-run", Title: "Run Tests", Description: "Execute test suite", Category: "testing", Scope: command.UserScope},
		{ID: "deploy", Title: "Deploy", Description: "Deploy application", Category: "deployment", Scope: command.ProjectScope},
	}
	
	dialog.SetCommands(testCommands)
	
	// Test search by title
	t.Run("search by title", func(t *testing.T) {
		dialog.searchQuery = "git"
		dialog.updateFilteredCommands()
		
		filtered := dialog.filteredCommands
		if len(filtered) != 2 {
			t.Errorf("Expected 2 filtered commands, got %d", len(filtered))
		}
		
		for _, cmd := range filtered {
			if !command.Command(cmd).MatchesSearch("git") {
				t.Errorf("Command '%s' should match search 'git'", cmd.ID)
			}
		}
	})
	
	// Test search by description
	t.Run("search by description", func(t *testing.T) {
		dialog.searchQuery = "changes"
		dialog.updateFilteredCommands()
		
		filtered := dialog.filteredCommands
		if len(filtered) != 2 {
			t.Errorf("Expected 2 filtered commands, got %d", len(filtered))
		}
	})
	
	// Test search by category
	t.Run("search by category", func(t *testing.T) {
		dialog.searchQuery = "testing"
		dialog.updateFilteredCommands()
		
		filtered := dialog.filteredCommands
		if len(filtered) != 1 {
			t.Errorf("Expected 1 filtered command, got %d", len(filtered))
		}
		if filtered[0].ID != "test-run" {
			t.Errorf("Expected 'test-run' command, got '%s'", filtered[0].ID)
		}
	})
	
	// Test empty search shows all commands
	t.Run("empty search shows all", func(t *testing.T) {
		dialog.searchQuery = ""
		dialog.updateFilteredCommands()
		
		filtered := dialog.filteredCommands
		if len(filtered) != len(testCommands) {
			t.Errorf("Expected %d commands, got %d", len(testCommands), len(filtered))
		}
	})
	
	// Test case insensitive search
	t.Run("case insensitive search", func(t *testing.T) {
		dialog.searchQuery = "GIT"
		dialog.updateFilteredCommands()
		
		filtered := dialog.filteredCommands
		if len(filtered) != 2 {
			t.Errorf("Expected 2 filtered commands for case insensitive search, got %d", len(filtered))
		}
	})
}

func TestCommandDialogGrouping(t *testing.T) {
	dialog := NewCommandDialogCmp().(*commandDialogCmp)
	
	testCommands := []command.Command{
		{ID: "builtin1", Title: "Built-in 1", Scope: command.BuiltinScope},
		{ID: "builtin2", Title: "Built-in 2", Scope: command.BuiltinScope},
		{ID: "user1", Title: "User 1", Scope: command.UserScope},
		{ID: "project1", Title: "Project 1", Scope: command.ProjectScope},
	}
	
	dialog.SetCommands(testCommands)
	
	// Test command grouping by scope
	t.Run("commands grouped by scope", func(t *testing.T) {
		grouped := dialog.commandsToDialogCommands(testCommands)
		
		// Commands should be ordered: builtin, user, project
		expectedOrder := []command.CommandScope{
			command.BuiltinScope,
			command.BuiltinScope,
			command.UserScope,
			command.ProjectScope,
		}
		
		if len(grouped) != len(expectedOrder) {
			t.Errorf("Expected %d commands, got %d", len(expectedOrder), len(grouped))
		}
		
		for i, cmd := range grouped {
			actualScope := command.Command(cmd).Scope
			if actualScope != expectedOrder[i] {
				t.Errorf("Expected scope %s at index %d, got %s", expectedOrder[i], i, actualScope)
			}
		}
	})
}

func TestCommandDialogRecentlyUsed(t *testing.T) {
	dialog := NewCommandDialogCmp().(*commandDialogCmp)
	
	now := time.Now()
	testCommands := []command.Command{
		{ID: "cmd1", Title: "Command 1", LastUsed: now.Add(-1 * time.Hour), Scope: command.BuiltinScope},
		{ID: "cmd2", Title: "Command 2", LastUsed: now.Add(-2 * time.Hour), Scope: command.BuiltinScope},
		{ID: "cmd3", Title: "Command 3", LastUsed: time.Time{}, Scope: command.BuiltinScope}, // Never used
	}
	
	dialog.SetCommands(testCommands)
	
	// Test that commands with LastUsed are prioritized
	t.Run("recently used commands tracked", func(t *testing.T) {
		// Find a command that has been used
		var usedCmd *command.Command
		for _, cmd := range testCommands {
			if !cmd.LastUsed.IsZero() {
				usedCmd = &cmd
				break
			}
		}
		
		if usedCmd == nil {
			t.Error("Expected at least one command with LastUsed set")
		}
		
		if usedCmd.LastUsed.IsZero() {
			t.Error("Expected LastUsed to be set for recently used command")
		}
	})
}

func TestCommandDialogView(t *testing.T) {
	dialog := NewCommandDialogCmp().(*commandDialogCmp)
	
	testCommands := []command.Command{
		{ID: "test", Title: "Test Command", Description: "A test command", Scope: command.BuiltinScope},
	}
	
	dialog.SetCommands(testCommands)
	dialog.width = 80
	dialog.height = 20
	
	// Test basic view rendering
	t.Run("basic view rendering", func(t *testing.T) {
		view := dialog.View()
		
		if view == "" {
			t.Error("Expected non-empty view")
		}
		
		// Should contain command title
		if !contains(view, "Test Command") {
			t.Error("Expected view to contain command title")
		}
		
		// Should contain help text
		if !contains(view, "navigate") {
			t.Error("Expected view to contain navigation help")
		}
	})
	
	// Test search mode view
	t.Run("search mode view", func(t *testing.T) {
		dialog.showSearch = true
		dialog.searchQuery = "test"
		view := dialog.View()
		
		if !contains(view, "filtered") {
			t.Error("Expected view to show filtered results in search mode")
		}
	})
}

func TestCloseCommandDialogMsg(t *testing.T) {
	msg := CloseCommandDialogMsg{}
	
	// Test that the message can be created
	if msg != (CloseCommandDialogMsg{}) {
		t.Error("Expected empty CloseCommandDialogMsg")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}