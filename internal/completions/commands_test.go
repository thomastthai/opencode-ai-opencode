package completions

import (
	"testing"

	"github.com/opencode-ai/opencode/internal/commands"
	"github.com/stretchr/testify/assert"
)

func TestCommandCompletionProvider_GetChildEntries(t *testing.T) {
	// Setup: Clear the global registry and register some mock commands
	commands.GetGlobalRegistry().Clear()
	cmd1 := commands.NewCommand("test1", "Test 1", "A test command").Build()
	cmd2 := commands.NewCommand("test2", "Test 2", "Another test command").Build()
	gitCmd := commands.NewCommand("git", "Git", "Git commands").
		WithSubCommands(
			commands.NewCommand("git:commit", "Commit", "Commit changes").Build(),
			commands.NewCommand("git:push", "Push", "Push changes").Build(),
		).
		Build()

	registry := commands.GetGlobalRegistry()
	registry.Register(cmd1)
	registry.Register(cmd2)
	registry.RegisterHierarchy(gitCmd)

	provider := NewCommandCompletionProvider()

	// Test: Top-level completion with no query
	items, err := provider.GetChildEntries("")
	assert.NoError(t, err)
	assert.Len(t, items, 3) // Should only show top-level commands

	// Test: Top-level completion with a query
	items, err = provider.GetChildEntries("test")
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	// Test: Sub-command completion
	items, err = provider.GetChildEntries("git:")
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	// Test: Sub-command completion with a query
	items, err = provider.GetChildEntries("git:com")
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "/git:commit", items[0].GetValue())
}
