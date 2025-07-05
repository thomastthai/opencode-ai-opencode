package commands

import (
	"context"
	"fmt"
)

func init() {
	// Register essential built-in commands
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

func handleHelp(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("OpenCode Commands:")
	fmt.Println("  /help, /h        - Show this help")
	fmt.Println("  /exit, /quit, /q - Exit OpenCode")
	fmt.Println("  /clear, /cls     - Clear current session")
	fmt.Println("  /list, /ls       - List all commands")
	fmt.Println("  /init            - Initialize project")
	fmt.Println("  /compact         - Compact session")
	fmt.Println("\nKeyboard Shortcuts:")
	fmt.Println("  Ctrl+K  - Command palette")
	fmt.Println("  Ctrl+S  - Switch session")
	fmt.Println("  Ctrl+O  - Model selection")
	fmt.Println("  Ctrl+H  - Toggle help")
	fmt.Println("  Ctrl+C  - Quit")
	return nil
}

func handleExit(ctx context.Context, args map[string]interface{}) error {
	// This should trigger application exit
	fmt.Println("Exiting OpenCode...")
	return fmt.Errorf("exit_requested")
}

func handleClear(ctx context.Context, args map[string]interface{}) error {
	// This should trigger a new session
	fmt.Println("Starting new session...")
	return fmt.Errorf("clear_session_requested")
}

func handleListCommands(ctx context.Context, args map[string]interface{}) error {
	registry := GetGlobalRegistry()
	commands := registry.List()
	
	fmt.Println("Available Commands:")
	for _, cmd := range commands {
		fmt.Printf("  /%s - %s\n", cmd.ID(), cmd.Description())
		if len(cmd.GetAliases()) > 0 {
			fmt.Printf("    Aliases: %v\n", cmd.GetAliases())
		}
	}
	return nil
}

func handleInit(ctx context.Context, args map[string]interface{}) error {
	// This should send the init prompt to the AI
	return fmt.Errorf("init_project_requested")
}

func handleCompact(ctx context.Context, args map[string]interface{}) error {
	// This should trigger session compacting
	return fmt.Errorf("compact_session_requested")
}

func handleGitCommit(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("git commit")
	return nil
}

func handleGitPush(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("git push")
	return nil
}