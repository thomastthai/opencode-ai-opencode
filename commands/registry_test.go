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