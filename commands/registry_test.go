package commands

import (
	"context"
	"testing"
)

// TestCommand implements Command interface for testing
type TestCommand struct {
	BaseCommand
	executed bool
}

func NewTestCommand(id, name, description string) *TestCommand {
	return &TestCommand{
		BaseCommand: BaseCommand{
			id:          id,
			name:        name,
			description: description,
			category:    "test",
			commandType: BuiltinCommand,
			arguments:   []ArgumentDefinition{},
			aliases:     []string{},
		},
	}
}

func (tc *TestCommand) Execute(ctx context.Context, args map[string]interface{}) error {
	tc.executed = true
	return nil
}

func TestNewCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()
	if registry == nil {
		t.Fatal("NewCommandRegistry() returned nil")
	}
	
	// Registry should start empty
	commands := registry.List()
	if len(commands) != 0 {
		t.Errorf("Expected empty registry, got %d commands", len(commands))
	}
}

func TestCommandRegistration(t *testing.T) {
	registry := NewCommandRegistry()
	cmd := NewTestCommand("test1", "Test Command", "A test command")
	
	// Register command
	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}
	
	// Verify command is registered
	retrievedCmd, exists := registry.Get("test1")
	if !exists {
		t.Fatal("Command not found after registration")
	}
	
	if retrievedCmd.ID() != cmd.ID() {
		t.Errorf("Expected command ID %s, got %s", cmd.ID(), retrievedCmd.ID())
	}
}

func TestCommandUnregistration(t *testing.T) {
	registry := NewCommandRegistry()
	cmd := NewTestCommand("test1", "Test Command", "A test command")
	
	// Register and then unregister
	registry.Register(cmd)
	err := registry.Unregister("test1")
	if err != nil {
		t.Fatalf("Failed to unregister command: %v", err)
	}
	
	// Verify command is gone
	_, exists := registry.Get("test1")
	if exists {
		t.Fatal("Command still exists after unregistration")
	}
}

func TestDuplicateRegistration(t *testing.T) {
	registry := NewCommandRegistry()
	cmd1 := NewTestCommand("test1", "Test Command 1", "First test command")
	cmd2 := NewTestCommand("test1", "Test Command 2", "Second test command")
	
	// Register first command
	err := registry.Register(cmd1)
	if err != nil {
		t.Fatalf("Failed to register first command: %v", err)
	}
	
	// Try to register second command with same ID
	err = registry.Register(cmd2)
	if err == nil {
		t.Fatal("Expected error when registering duplicate command ID")
	}
}

func TestCommandList(t *testing.T) {
	registry := NewCommandRegistry()
	cmd1 := NewTestCommand("test1", "Test Command 1", "First test command")
	cmd2 := NewTestCommand("test2", "Test Command 2", "Second test command")
	
	registry.Register(cmd1)
	registry.Register(cmd2)
	
	commands := registry.List()
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
}

func TestCommandsByCategory(t *testing.T) {
	registry := NewCommandRegistry()
	cmd1 := NewTestCommand("test1", "Test Command 1", "First test command")
	cmd1.category = "category1"
	
	cmd2 := NewTestCommand("test2", "Test Command 2", "Second test command")
	cmd2.category = "category2"
	
	registry.Register(cmd1)
	registry.Register(cmd2)
	
	category1Commands := registry.GetByCategory("category1")
	if len(category1Commands) != 1 {
		t.Errorf("Expected 1 command in category1, got %d", len(category1Commands))
	}
	
	if category1Commands[0].ID() != "test1" {
		t.Errorf("Expected test1 in category1, got %s", category1Commands[0].ID())
	}
}

func TestCommandsByType(t *testing.T) {
	registry := NewCommandRegistry()
	cmd1 := NewTestCommand("test1", "Test Command 1", "First test command")
	cmd1.commandType = BuiltinCommand
	
	cmd2 := NewTestCommand("test2", "Test Command 2", "Second test command")
	cmd2.commandType = UserCommand
	
	registry.Register(cmd1)
	registry.Register(cmd2)
	
	builtinCommands := registry.GetByType(BuiltinCommand)
	if len(builtinCommands) != 1 {
		t.Errorf("Expected 1 builtin command, got %d", len(builtinCommands))
	}
	
	userCommands := registry.GetByType(UserCommand)
	if len(userCommands) != 1 {
		t.Errorf("Expected 1 user command, got %d", len(userCommands))
	}
}

func TestBaseCommandValidation(t *testing.T) {
	cmd := NewTestCommand("test1", "Test Command", "A test command")
	
	// Add a required argument
	cmd.arguments = []ArgumentDefinition{
		{
			Name:     "required_arg",
			Type:     "string",
			Required: true,
		},
	}
	
	// Test with missing required argument
	args := map[string]interface{}{}
	err := cmd.ValidateArgs(args)
	if err == nil {
		t.Fatal("Expected validation error for missing required argument")
	}
	
	// Test with valid arguments
	args = map[string]interface{}{
		"required_arg": "value",
	}
	err = cmd.ValidateArgs(args)
	if err != nil {
		t.Fatalf("Unexpected validation error: %v", err)
	}
}

func TestCommandType_String(t *testing.T) {
	tests := []struct {
		commandType CommandType
		expected    string
	}{
		{BuiltinCommand, "builtin"},
		{UserCommand, "user"},
		{ProjectCommand, "project"},
		{PluginCommand, "plugin"},
		{CommandType(999), "unknown"},
	}
	
	for _, test := range tests {
		result := test.commandType.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Save original registry
	originalRegistry := globalRegistry
	defer func() {
		globalRegistry = originalRegistry
	}()
	
	// Create fresh registry for test
	SetGlobalRegistry(NewCommandRegistry())
	
	cmd := NewTestCommand("global_test", "Global Test Command", "Test global registry")
	
	// Test global registration
	err := RegisterCommand(cmd)
	if err != nil {
		t.Fatalf("Failed to register command globally: %v", err)
	}
	
	// Test global retrieval
	retrievedCmd, exists := GetCommand("global_test")
	if !exists {
		t.Fatal("Command not found in global registry")
	}
	
	if retrievedCmd.ID() != cmd.ID() {
		t.Errorf("Expected command ID %s, got %s", cmd.ID(), retrievedCmd.ID())
	}
	
	// Test global listing
	commands := ListCommands()
	if len(commands) != 1 {
		t.Errorf("Expected 1 command in global registry, got %d", len(commands))
	}
}

func TestCommandBuilder(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) error {
		return nil
	}
	
	cmd := NewCommand("test_builder", "Test Builder", "Test command builder").
		WithCategory("test").
		WithType(BuiltinCommand).
		WithHandler(handler).
		WithAliases([]string{"tb", "test"}).
		WithExample("test_builder --help").
		WithMetadataValue("version", "1.0").
		WithMetadataValue("author", "test").
		Build()
	
	if cmd.ID() != "test_builder" {
		t.Errorf("Expected ID 'test_builder', got %s", cmd.ID())
	}
	
	if cmd.Category() != "test" {
		t.Errorf("Expected category 'test', got %s", cmd.Category())
	}
	
	if cmd.Type() != BuiltinCommand {
		t.Errorf("Expected type BuiltinCommand, got %s", cmd.Type().String())
	}
	
	aliases := cmd.GetAliases()
	if len(aliases) != 2 || aliases[0] != "tb" || aliases[1] != "test" {
		t.Errorf("Expected aliases [tb, test], got %v", aliases)
	}
	
	metadata := cmd.GetMetadata()
	if metadata["version"] != "1.0" || metadata["author"] != "test" {
		t.Errorf("Expected metadata with version=1.0 and author=test, got %v", metadata)
	}
	
	// Test handler execution
	err := cmd.Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Errorf("Expected handler execution to succeed, got error: %v", err)
	}
}

func TestCommandHierarchy(t *testing.T) {
	// Create sub-command
	subCmd := NewCommand("commit", "Commit", "Commit changes").
		WithCategory("vcs").
		WithType(BuiltinCommand).
		Build()
	
	// Add sub-command to parent
	parentWithSub := NewCommand("git", "Git", "Git version control").
		WithCategory("vcs").
		WithType(BuiltinCommand).
		AddSubCommand(subCmd).
		Build()
	
	// Test parent-child relationship
	subCommands := parentWithSub.GetSubCommands()
	if len(subCommands) != 1 {
		t.Errorf("Expected 1 sub-command, got %d", len(subCommands))
	}
	
	if subCommands[0].ID() != "commit" {
		t.Errorf("Expected sub-command ID 'commit', got %s", subCommands[0].ID())
	}
	
	// Test parent reference
	if subCmd.GetParent() == nil {
		t.Error("Expected sub-command to have parent reference")
	} else if subCmd.GetParent().ID() != "git" {
		t.Errorf("Expected parent ID 'git', got %s", subCmd.GetParent().ID())
	}
	
	// Test command path
	expectedPath := "git commit"
	if subCmd.GetPath() != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, subCmd.GetPath())
	}
}

func TestRegistryHierarchy(t *testing.T) {
	registry := NewCommandRegistry()
	
	// Create parent command with sub-commands
	subCmd := NewCommand("commit", "Commit", "Commit changes").
		WithType(BuiltinCommand).
		Build()
	
	parentCmd := NewCommand("git", "Git", "Git version control").
		WithType(BuiltinCommand).
		AddSubCommand(subCmd).
		Build()
	
	// Register hierarchy
	err := registry.RegisterHierarchy(parentCmd)
	if err != nil {
		t.Fatalf("Failed to register hierarchy: %v", err)
	}
	
	// Verify both commands are registered
	_, exists := registry.Get("git")
	if !exists {
		t.Error("Parent command not found in registry")
	}
	
	_, exists = registry.Get("commit")
	if !exists {
		t.Error("Sub-command not found in registry")
	}
	
	// Test GetByPath
	cmd, exists := registry.GetByPath("git commit")
	if !exists {
		t.Error("Command not found by path")
	} else if cmd.ID() != "commit" {
		t.Errorf("Expected command ID 'commit', got %s", cmd.ID())
	}
	
	// Test GetRootCommands
	rootCommands := registry.GetRootCommands()
	if len(rootCommands) != 1 {
		t.Errorf("Expected 1 root command, got %d", len(rootCommands))
	} else if rootCommands[0].ID() != "git" {
		t.Errorf("Expected root command ID 'git', got %s", rootCommands[0].ID())
	}
}

func TestCommandMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"version":     "1.0",
		"author":      "test",
		"deprecated":  false,
		"priority":    10,
	}
	
	cmd := NewCommand("test_meta", "Test Metadata", "Test metadata handling").
		WithMetadata(metadata).
		WithMetadataValue("extra", "value").
		Build()
	
	retrievedMetadata := cmd.GetMetadata()
	
	// Check all metadata values
	if retrievedMetadata["version"] != "1.0" {
		t.Errorf("Expected version '1.0', got %v", retrievedMetadata["version"])
	}
	
	if retrievedMetadata["author"] != "test" {
		t.Errorf("Expected author 'test', got %v", retrievedMetadata["author"])
	}
	
	if retrievedMetadata["deprecated"] != false {
		t.Errorf("Expected deprecated false, got %v", retrievedMetadata["deprecated"])
	}
	
	if retrievedMetadata["priority"] != 10 {
		t.Errorf("Expected priority 10, got %v", retrievedMetadata["priority"])
	}
	
	if retrievedMetadata["extra"] != "value" {
		t.Errorf("Expected extra 'value', got %v", retrievedMetadata["extra"])
	}
	
	// Test metadata isolation (returned copy)
	retrievedMetadata["modified"] = true
	newMetadata := cmd.GetMetadata()
	if _, exists := newMetadata["modified"]; exists {
		t.Error("Metadata should be isolated - external modification should not affect original")
	}
}

func TestCommandHandler(t *testing.T) {
	executed := false
	handler := func(ctx context.Context, args map[string]interface{}) error {
		executed = true
		return nil
	}
	
	cmd := NewCommand("test_handler", "Test Handler", "Test command handler").
		WithHandler(handler).
		Build()
	
	// Execute command
	err := cmd.Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Errorf("Expected handler execution to succeed, got error: %v", err)
	}
	
	if !executed {
		t.Error("Handler was not executed")
	}
}

func TestGlobalHierarchyFunctions(t *testing.T) {
	// Save original registry
	originalRegistry := globalRegistry
	defer func() {
		globalRegistry = originalRegistry
	}()
	
	// Create fresh registry for test
	SetGlobalRegistry(NewCommandRegistry())
	
	// Create command hierarchy
	subCmd := NewCommand("push", "Push", "Push changes").
		WithType(BuiltinCommand).
		Build()
	
	parentCmd := NewCommand("git", "Git", "Git version control").
		WithType(BuiltinCommand).
		AddSubCommand(subCmd).
		Build()
	
	// Test global hierarchy registration
	err := RegisterCommandHierarchy(parentCmd)
	if err != nil {
		t.Fatalf("Failed to register hierarchy globally: %v", err)
	}
	
	// Test global path retrieval
	cmd, exists := GetCommandByPath("git push")
	if !exists {
		t.Error("Command not found by path globally")
	} else if cmd.ID() != "push" {
		t.Errorf("Expected command ID 'push', got %s", cmd.ID())
	}
	
	// Test global root commands
	rootCommands := GetRootCommands()
	if len(rootCommands) != 1 {
		t.Errorf("Expected 1 root command globally, got %d", len(rootCommands))
	} else if rootCommands[0].ID() != "git" {
		t.Errorf("Expected root command ID 'git', got %s", rootCommands[0].ID())
	}
}