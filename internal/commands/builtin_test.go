package commands

import (
	"context"
	"testing"
	"strings"
	"os"
	"io"
)

func TestBuiltInCommandHandlers(t *testing.T) {
	ctx := context.Background()

	t.Run("handleHelp", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := handleHelp(ctx, nil)
		
		w.Close()
		os.Stdout = oldStdout
		
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		if err != nil {
			t.Errorf("handleHelp should not return error, got: %v", err)
		}
		
		// Check that help contains expected commands
		expectedCommands := []string{"/help", "/exit", "/clear", "/list", "/init", "/compact"}
		for _, cmd := range expectedCommands {
			if !strings.Contains(outputStr, cmd) {
				t.Errorf("Help output should contain %s, got: %s", cmd, outputStr)
			}
		}
		
		// Check that keyboard shortcuts are included
		expectedShortcuts := []string{"Ctrl+K", "Ctrl+S", "Ctrl+O"}
		for _, shortcut := range expectedShortcuts {
			if !strings.Contains(outputStr, shortcut) {
				t.Errorf("Help output should contain shortcut %s, got: %s", shortcut, outputStr)
			}
		}
	})

	t.Run("handleExit", func(t *testing.T) {
		err := handleExit(ctx, nil)
		
		if err == nil {
			t.Error("handleExit should return an error to signal exit")
		}
		
		if !strings.Contains(err.Error(), "exit_requested") {
			t.Errorf("Expected error to contain 'exit_requested', got: %v", err)
		}
	})

	t.Run("handleClear", func(t *testing.T) {
		err := handleClear(ctx, nil)
		
		if err == nil {
			t.Error("handleClear should return an error to signal session clear")
		}
		
		if !strings.Contains(err.Error(), "clear_session_requested") {
			t.Errorf("Expected error to contain 'clear_session_requested', got: %v", err)
		}
	})

	t.Run("handleListCommands", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := handleListCommands(ctx, nil)
		
		w.Close()
		os.Stdout = oldStdout
		
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		if err != nil {
			t.Errorf("handleListCommands should not return error, got: %v", err)
		}
		
		// Should contain "Available Commands"
		if !strings.Contains(outputStr, "Available Commands") {
			t.Errorf("Expected output to contain 'Available Commands', got: %s", outputStr)
		}
	})

	t.Run("handleInit", func(t *testing.T) {
		err := handleInit(ctx, nil)
		
		if err == nil {
			t.Error("handleInit should return an error to signal init request")
		}
		
		if !strings.Contains(err.Error(), "init_project_requested") {
			t.Errorf("Expected error to contain 'init_project_requested', got: %v", err)
		}
	})

	t.Run("handleCompact", func(t *testing.T) {
		err := handleCompact(ctx, nil)
		
		if err == nil {
			t.Error("handleCompact should return an error to signal compact request")
		}
		
		if !strings.Contains(err.Error(), "compact_session_requested") {
			t.Errorf("Expected error to contain 'compact_session_requested', got: %v", err)
		}
	})

	t.Run("handleGitCommit", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := handleGitCommit(ctx, nil)
		
		w.Close()
		os.Stdout = oldStdout
		
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		if err != nil {
			t.Errorf("handleGitCommit should not return error, got: %v", err)
		}
		
		if !strings.Contains(outputStr, "git commit") {
			t.Errorf("Expected output to contain 'git commit', got: %s", outputStr)
		}
	})

	t.Run("handleGitPush", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := handleGitPush(ctx, nil)
		
		w.Close()
		os.Stdout = oldStdout
		
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		if err != nil {
			t.Errorf("handleGitPush should not return error, got: %v", err)
		}
		
		if !strings.Contains(outputStr, "git push") {
			t.Errorf("Expected output to contain 'git push', got: %s", outputStr)
		}
	})
}

func TestBuiltInCommandRegistration(t *testing.T) {
	// Save original registry state
	originalRegistry := globalRegistry
	defer func() {
		globalRegistry = originalRegistry
	}()
	
	// Reset global registry for clean test
	globalRegistry = NewCommandRegistry()
	
	// The init() function should have registered built-in commands
	// Let's manually trigger registration for testing
	commands := []struct {
		id          string
		expectType  CommandType
		expectAlias bool
	}{
		{"help", BuiltinCommand, true},
		{"exit", BuiltinCommand, true},
		{"clear", BuiltinCommand, true},
		{"list", BuiltinCommand, true},
		{"init", BuiltinCommand, false},
		{"compact", BuiltinCommand, true},
	}
	
	// Re-register commands for testing
	registerTestBuiltIns()
	
	for _, tc := range commands {
		t.Run("command_"+tc.id, func(t *testing.T) {
			cmd, exists := GetGlobalRegistry().Get(tc.id)
			if !exists {
				t.Errorf("Command %s should be registered", tc.id)
				return
			}
			
			if cmd.Type() != tc.expectType {
				t.Errorf("Command %s should have type %v, got %v", tc.id, tc.expectType, cmd.Type())
			}
			
			if tc.expectAlias && len(cmd.GetAliases()) == 0 {
				t.Errorf("Command %s should have aliases", tc.id)
			}
		})
	}
}

func TestBuiltInCommandAliases(t *testing.T) {
	// Save original registry state
	originalRegistry := globalRegistry
	defer func() {
		globalRegistry = originalRegistry
	}()
	
	// Reset and re-register
	globalRegistry = NewCommandRegistry()
	registerTestBuiltIns()
	
	aliasTests := []struct {
		commandID string
		aliases   []string
	}{
		{"help", []string{"h"}},
		{"exit", []string{"quit", "q"}},
		{"clear", []string{"cls", "new"}},
		{"list", []string{"ls", "commands"}},
		{"compact", []string{"summary"}},
	}
	
	for _, tc := range aliasTests {
		t.Run("aliases_"+tc.commandID, func(t *testing.T) {
			cmd, exists := GetGlobalRegistry().Get(tc.commandID)
			if !exists {
				t.Errorf("Command %s should be registered", tc.commandID)
				return
			}
			
			aliases := cmd.GetAliases()
			if len(aliases) != len(tc.aliases) {
				t.Errorf("Command %s should have %d aliases, got %d", tc.commandID, len(tc.aliases), len(aliases))
				return
			}
			
			for i, expectedAlias := range tc.aliases {
				if i >= len(aliases) || aliases[i] != expectedAlias {
					t.Errorf("Command %s alias %d should be %s, got %s", tc.commandID, i, expectedAlias, aliases[i])
				}
			}
		})
	}
}

func TestBuiltInCommandExecution(t *testing.T) {
	ctx := context.Background()
	
	// Test that commands can be executed through the registry
	t.Run("execute through registry", func(t *testing.T) {
		// Save original registry state
		originalRegistry := globalRegistry
		defer func() {
			globalRegistry = originalRegistry
		}()
		
		globalRegistry = NewCommandRegistry()
		registerTestBuiltIns()
		
		cmd, exists := GetGlobalRegistry().Get("help")
		if !exists {
			t.Fatal("Help command should be registered")
		}
		
		// Capture stdout for help command
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		err := cmd.Execute(ctx, nil)
		
		w.Close()
		os.Stdout = oldStdout
		
		output, _ := io.ReadAll(r)
		outputStr := string(output)
		
		if err != nil {
			t.Errorf("Command execution should not return error, got: %v", err)
		}
		
		if !strings.Contains(outputStr, "OpenCode Commands") {
			t.Errorf("Expected help output, got: %s", outputStr)
		}
	})
}

func TestBuiltInCommandHierarchy(t *testing.T) {
	// Save original registry state
	originalRegistry := globalRegistry
	defer func() {
		globalRegistry = originalRegistry
	}()
	
	globalRegistry = NewCommandRegistry()
	registerTestBuiltIns()
	
	// Test that git command and its sub-commands are registered
	t.Run("git hierarchy", func(t *testing.T) {
		gitCmd, exists := GetGlobalRegistry().Get("git")
		if !exists {
			t.Error("Git command should be registered")
			return
		}
		
		subCommands := gitCmd.GetSubCommands()
		if len(subCommands) != 2 {
			t.Errorf("Git command should have 2 sub-commands, got %d", len(subCommands))
		}
		
		// Check that sub-commands are also registered individually
		commitCmd, exists := GetGlobalRegistry().Get("commit")
		if !exists {
			t.Error("Git commit sub-command should be registered")
		}
		
		pushCmd, exists := GetGlobalRegistry().Get("push")
		if !exists {
			t.Error("Git push sub-command should be registered")
		}
		
		// Check that sub-commands have correct parent
		if commitCmd.GetParent() == nil || commitCmd.GetParent().ID() != "git" {
			t.Error("Commit command should have git as parent")
		}
		
		if pushCmd.GetParent() == nil || pushCmd.GetParent().ID() != "git" {
			t.Error("Push command should have git as parent")
		}
	})
}

// Helper function to re-register built-in commands for testing
func registerTestBuiltIns() {
	RegisterBuiltIn(
		NewCommand("help", "Help", "Show available commands and keyboard shortcuts").
			WithType(BuiltinCommand).
			WithHandler(handleHelp).
			WithAliases([]string{"h"}).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("exit", "Exit", "Exit OpenCode").
			WithType(BuiltinCommand).
			WithHandler(handleExit).
			WithAliases([]string{"quit", "q"}).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("clear", "Clear", "Clear current session and start new").
			WithType(BuiltinCommand).
			WithHandler(handleClear).
			WithAliases([]string{"cls", "new"}).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("list", "List Commands", "List all available commands").
			WithType(BuiltinCommand).
			WithHandler(handleListCommands).
			WithAliases([]string{"ls", "commands"}).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("init", "Initialize", "Create/Update the OpenCode.md memory file").
			WithType(BuiltinCommand).
			WithHandler(handleInit).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("compact", "Compact", "Summarize current session and create new one").
			WithType(BuiltinCommand).
			WithHandler(handleCompact).
			WithAliases([]string{"summary"}).
			Build(),
	)

	// Register a command with sub-commands for testing
	gitCmd := NewCommand("git", "Git", "Git commands").
		WithType(BuiltinCommand).
		WithSubCommands(
			NewCommand("commit", "Commit", "Commit changes").
				WithType(BuiltinCommand).
				WithHandler(handleGitCommit).
				Build(),
			NewCommand("push", "Push", "Push changes").
				WithType(BuiltinCommand).
				WithHandler(handleGitPush).
				Build(),
		).
		Build()
	RegisterBuiltInHierarchy(gitCmd)
}