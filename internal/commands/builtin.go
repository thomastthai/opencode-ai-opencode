package commands

import (
	"context"
	"fmt"
)

func init() {
	// Register top-level built-in commands here.
	// Sub-commands should be added to their parent command.
	RegisterBuiltIn(
		NewCommand("hello", "Hello", "A simple hello command").
			WithType(BuiltinCommand).
			WithHandler(handleHello).
			Build(),
	)

	// Register a command with sub-commands
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

func handleHello(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("Hello, world!")
	return nil
}

func handleGitCommit(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("git commit")
	return nil
}

func handleGitPush(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("git push")
	return nil
}