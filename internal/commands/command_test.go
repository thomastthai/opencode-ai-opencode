package commands

import (
	"context"
	"testing"
)

func TestCommandBuilder(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) error {
		return nil
	}

	args := []ArgumentDefinition{
		{Name: "file", Type: "string", Required: true},
	}

	cmd := NewCommand("test-builder", "Test Builder", "A test command from the builder").
		WithCategory("testing").
		WithType(BuiltinCommand).
		WithExample("test-builder --file /path/to/file").
		WithAliases([]string{"tb", "testb"}).
		WithArguments(args).
		WithHandler(handler).
		Build()

	if cmd.ID() != "test-builder" {
		t.Errorf("Expected ID 'test-builder', got '%s'", cmd.ID())
	}
	if cmd.Name() != "Test Builder" {
		t.Errorf("Expected Name 'Test Builder', got '%s'", cmd.Name())
	}
	if cmd.Description() != "A test command from the builder" {
		t.Errorf("Expected Description, got '%s'", cmd.Description())
	}
	if cmd.Category() != "testing" {
		t.Errorf("Expected Category 'testing', got '%s'", cmd.Category())
	}
	if cmd.Type() != BuiltinCommand {
		t.Errorf("Expected Type BuiltinCommand, got %v", cmd.Type())
	}
	if cmd.Example() != "test-builder --file /path/to/file" {
		t.Errorf("Expected Example, got '%s'", cmd.Example())
	}
	if len(cmd.GetAliases()) != 2 || cmd.GetAliases()[0] != "tb" {
		t.Errorf("Expected Aliases [tb, testb], got %v", cmd.GetAliases())
	}
	if len(cmd.GetArguments()) != 1 || cmd.GetArguments()[0].Name != "file" {
		t.Errorf("Expected Arguments to contain 'file', got %v", cmd.GetArguments())
	}
	if err := cmd.Execute(context.Background(), nil); err != nil {
		t.Errorf("Handler was not set correctly")
	}
}

func TestCommandHierarchyAndPath(t *testing.T) {
	subSubCmd := NewCommand("add", "Add", "Add a remote").Build()
	subCmd := NewCommand("remote", "Remote", "Manage remotes").
		WithSubCommands(subSubCmd).
		Build()
	NewCommand("git", "Git", "Git VCS").
		WithSubCommands(subCmd).
		Build()

	if subSubCmd.GetParent() == nil || subSubCmd.GetParent().ID() != "remote" {
		t.Error("Sub-command parent is not set correctly")
	}

	expectedPath := "Git Remote Add"
	if path := subSubCmd.GetPath(); path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, path)
	}

	expectedParentPath := "Git Remote"
	if path := subCmd.GetPath(); path != expectedParentPath {
		t.Errorf("Expected path '%s', got '%s'", expectedParentPath, path)
	}
}

func TestArgumentValidation(t *testing.T) {
	cmd := NewCommand("test-args", "Test Args", "Test argument validation").
		WithArguments([]ArgumentDefinition{
			{Name: "required_arg", Type: "string", Required: true},
			{Name: "optional_arg", Type: "string", Required: false},
		}).
		Build()

	// Test case 1: Missing required argument
	args1 := map[string]interface{}{"optional_arg": "value"}
	if err := cmd.ValidateArgs(args1); err == nil {
		t.Error("Expected error for missing required argument, but got nil")
	}

	// Test case 2: All required arguments present
	args2 := map[string]interface{}{"required_arg": "value"}
	if err := cmd.ValidateArgs(args2); err != nil {
		t.Errorf("Did not expect error for valid arguments, but got: %v", err)
	}

	// Test case 3: No arguments for a command that has required ones
	if err := cmd.ValidateArgs(nil); err == nil {
		t.Error("Expected error when passing nil arguments to a command with required args")
	}
}
