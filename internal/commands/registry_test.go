package commands

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestNewCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()
	if registry == nil {
		t.Fatal("NewCommandRegistry() returned nil")
	}
	if len(registry.List()) != 0 {
		t.Errorf("Expected empty registry, got %d commands", len(registry.List()))
	}
}

func TestCommandRegistration(t *testing.T) {
	registry := NewCommandRegistry()
	cmd := NewCommand("test", "Test", "A test command").Build()
	err := registry.Register(cmd)
	if err != nil {
		t.Fatalf("Failed to register command: %v", err)
	}

	retrievedCmd, exists := registry.Get("test")
	if !exists {
		t.Fatal("Command not found after registration")
	}
	if retrievedCmd.ID() != "test" {
		t.Errorf("Expected command ID 'test', got '%s'", retrievedCmd.ID())
	}
}

func TestUnregisterCommand(t *testing.T) {
	registry := NewCommandRegistry()
	cmd := NewCommand("test-unregister", "Test Unregister", "A command to be unregistered").Build()
	
	registry.Register(cmd)

	_, exists := registry.Get("test-unregister")
	if !exists {
		t.Fatal("Command not found after registration")
	}

	err := registry.Unregister("test-unregister")
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	_, exists = registry.Get("test-unregister")
	if exists {
		t.Fatal("Command still exists after unregistration")
	}
}

func TestListCommands(t *testing.T) {
	registry := NewCommandRegistry()
	cmd1 := NewCommand("test1", "Test 1", "First test command").Build()
	cmd2 := NewCommand("test2", "Test 2", "Second test command").Build()

	registry.Register(cmd1)
	registry.Register(cmd2)

	commands := registry.List()
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, but got %d", len(commands))
	}
}

func TestDuplicateRegistration(t *testing.T) {
	registry := NewCommandRegistry()
	cmd1 := NewCommand("test", "Test 1", "First test command").Build()
	cmd2 := NewCommand("test", "Test 2", "Second test command").Build()

	err := registry.Register(cmd1)
	if err != nil {
		t.Fatalf("Failed to register first command: %v", err)
	}

	err = registry.Register(cmd2)
	if err == nil {
		t.Fatal("Expected error when registering duplicate command")
	}
}

func TestRegisterBuiltIn(t *testing.T) {
	// Reset global registry for a clean test
	globalRegistry = NewCommandRegistry()

	cmd := NewCommand("builtin-test", "Built-in Test", "A test built-in command").
		WithType(BuiltinCommand).
		Build()
	
	// Use a separate function to avoid log.Fatalf
	registerTestBuiltIn := func(c Command) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = r.(error)
			}
		}()
		RegisterBuiltIn(c)
		return
	}

	err := registerTestBuiltIn(cmd)
	if err != nil {
		t.Fatalf("RegisterBuiltIn failed: %v", err)
	}

	retrievedCmd, exists := GetGlobalRegistry().Get("builtin-test")
	if !exists {
		t.Fatal("Built-in command not found after registration")
	}
	if retrievedCmd.Type() != BuiltinCommand {
		t.Errorf("Expected command type BuiltinCommand, got %v", retrievedCmd.Type())
	}
}

func TestRegisterBuiltInHierarchy(t *testing.T) {
	// Reset global registry
	globalRegistry = NewCommandRegistry()

	subCmd := NewCommand("sub", "Sub", "A sub command").Build()
	parentCmd := NewCommand("parent", "Parent", "A parent command").
		WithSubCommands(subCmd).
		Build()

	RegisterBuiltInHierarchy(parentCmd)

	if _, exists := GetGlobalRegistry().Get("parent"); !exists {
		t.Fatal("Parent command not registered")
	}
	if _, exists := GetGlobalRegistry().Get("sub"); !exists {
		t.Fatal("Sub command not registered")
	}
}

func TestRegisterHierarchyFailure(t *testing.T) {
	registry := NewCommandRegistry()
	
	// Hierarchy 1
	subCmd1 := NewCommand("sub", "Sub", "A sub command").Build()
	parentCmd1 := NewCommand("parent1", "Parent 1", "A parent command").
		WithSubCommands(subCmd1).
		Build()

	// Hierarchy 2 with duplicate sub-command ID
	subCmd2 := NewCommand("sub", "Sub", "Another sub command").Build()
	parentCmd2 := NewCommand("parent2", "Parent 2", "Another parent command").
		WithSubCommands(subCmd2).
		Build()

	err := registry.RegisterHierarchy(parentCmd1)
	if err != nil {
		t.Fatalf("Registering first hierarchy failed: %v", err)
	}

	err = registry.RegisterHierarchy(parentCmd2)
	if err == nil {
		t.Fatal("Expected error when registering hierarchy with duplicate sub-command ID")
	}

	// Ensure the second parent was not partially registered
	if _, exists := registry.Get("parent2"); exists {
		t.Fatal("Parent of duplicate sub-command should not be registered")
	}
}

// Enhancement: Edge cases and additional validation tests
func TestCommandRegistrationEdgeCases(t *testing.T) {
	registry := NewCommandRegistry()
	
	// Test with commands containing special characters
	t.Run("commands with special characters", func(t *testing.T) {
		specialCommands := []struct {
			id   string
			name string
		}{
			{"git-commit", "Git Commit"},
			{"test_command", "Test Command"},
			{"command.with.dots", "Command With Dots"},
			{"command@symbol", "Command At Symbol"},
			{"123numeric", "Numeric Start"},
			{"UPPERCASE", "Uppercase Command"},
		}
		
		for _, tc := range specialCommands {
			cmd := NewCommand(tc.id, tc.name, "Test command with special characters").Build()
			err := registry.Register(cmd)
			if err != nil {
				t.Errorf("Failed to register command with ID '%s': %v", tc.id, err)
			}
			
			// Verify retrieval
			retrieved, exists := registry.Get(tc.id)
			if !exists {
				t.Errorf("Command with ID '%s' should exist", tc.id)
			}
			if retrieved.ID() != tc.id {
				t.Errorf("Retrieved command ID should be '%s', got '%s'", tc.id, retrieved.ID())
			}
		}
	})
	
	// Test command lookup by aliases
	t.Run("command lookup by aliases", func(t *testing.T) {
		cmd := NewCommand("test-aliases", "Test Aliases", "Command with aliases").
			WithAliases([]string{"ta", "test", "alias-cmd"}).
			Build()
		
		err := registry.Register(cmd)
		if err != nil {
			t.Fatalf("Failed to register command: %v", err)
		}
		
		// Note: Current registry doesn't support lookup by aliases
		// This test documents the current behavior
		for _, alias := range cmd.GetAliases() {
			_, exists := registry.Get(alias)
			if exists {
				t.Errorf("Lookup by alias '%s' should not work with current implementation", alias)
			}
		}
		
		// Only ID lookup should work
		retrieved, exists := registry.Get("test-aliases")
		if !exists {
			t.Error("Command should be found by ID")
		}
		if len(retrieved.GetAliases()) != 3 {
			t.Errorf("Expected 3 aliases, got %d", len(retrieved.GetAliases()))
		}
	})
	
	// Test registry state after partial hierarchy registration failure
	t.Run("registry state after partial failure", func(t *testing.T) {
		// Start with clean registry
		testRegistry := NewCommandRegistry()
		
		// Register a command that will conflict
		conflictCmd := NewCommand("conflict", "Conflict", "A conflicting command").Build()
		err := testRegistry.Register(conflictCmd)
		if err != nil {
			t.Fatalf("Failed to register initial command: %v", err)
		}
		
		// Try to register hierarchy with conflicting sub-command
		subCmd := NewCommand("conflict", "Sub Conflict", "A conflicting sub-command").Build()
		parentCmd := NewCommand("parent-conflict", "Parent Conflict", "Parent with conflicting sub").
			WithSubCommands(subCmd).
			Build()
		
		initialCount := len(testRegistry.List())
		
		err = testRegistry.RegisterHierarchy(parentCmd)
		if err == nil {
			t.Error("Expected error when registering hierarchy with conflicting sub-command")
		}
		
		// Verify registry state is unchanged
		finalCount := len(testRegistry.List())
		if finalCount != initialCount {
			t.Errorf("Registry should be unchanged after failed hierarchy registration. Initial: %d, Final: %d", initialCount, finalCount)
		}
		
		// Verify original command still exists
		original, exists := testRegistry.Get("conflict")
		if !exists {
			t.Error("Original conflicting command should still exist")
		}
		if original.Name() != "Conflict" {
			t.Errorf("Original command should be unchanged, got name: %s", original.Name())
		}
		
		// Verify parent was not registered
		if _, exists := testRegistry.Get("parent-conflict"); exists {
			t.Error("Parent command should not have been registered due to conflict")
		}
	})
	
	// Test with large command sets and memory efficiency
	t.Run("memory usage with large command sets", func(t *testing.T) {
		largeRegistry := NewCommandRegistry()
		
		// Register many commands to test memory efficiency
		numCommands := 1000
		for i := 0; i < numCommands; i++ {
			cmd := NewCommand(
				fmt.Sprintf("large-cmd-%d", i),
				fmt.Sprintf("Large Command %d", i),
				fmt.Sprintf("Description for large command %d with some extra text to increase memory usage", i),
			).
				WithCategory(fmt.Sprintf("category-%d", i%10)). // 10 different categories
				WithAliases([]string{fmt.Sprintf("lc%d", i), fmt.Sprintf("large%d", i)}).
				Build()
			
			err := largeRegistry.Register(cmd)
			if err != nil {
				t.Fatalf("Failed to register large command %d: %v", i, err)
			}
		}
		
		// Verify all commands are registered
		commands := largeRegistry.List()
		if len(commands) != numCommands {
			t.Errorf("Expected %d commands, got %d", numCommands, len(commands))
		}
		
		// Test random access
		for i := 0; i < 100; i++ {
			randomID := fmt.Sprintf("large-cmd-%d", i*10) // Every 10th command
			cmd, exists := largeRegistry.Get(randomID)
			if !exists {
				t.Errorf("Command %s should exist", randomID)
			}
			if cmd.ID() != randomID {
				t.Errorf("Expected command ID %s, got %s", randomID, cmd.ID())
			}
		}
	})
}

func TestCommandBuilderValidation(t *testing.T) {
	// Test builder with invalid input
	t.Run("builder with empty strings", func(t *testing.T) {
		cmd := NewCommand("", "", "").Build()
		
		if cmd.ID() != "" {
			t.Error("Empty ID should remain empty")
		}
		if cmd.Name() != "" {
			t.Error("Empty name should remain empty")
		}
		if cmd.Description() != "" {
			t.Error("Empty description should remain empty")
		}
	})
	
	// Test builder with nil values (where applicable)
	t.Run("builder with nil values", func(t *testing.T) {
		cmd := NewCommand("test-nil", "Test Nil", "Test with nil values").
			WithAliases(nil).
			WithArguments(nil).
			WithMetadata(nil).
			WithHandler(nil).
			Build()
		
		// Check that the command handles nil values gracefully
		aliases := cmd.GetAliases()
		if aliases == nil {
			t.Log("Aliases are nil when set to nil - this is current behavior")
		} else if len(aliases) != 0 {
			t.Errorf("Expected empty aliases, got %v", aliases)
		}
		
		arguments := cmd.GetArguments()
		if arguments == nil {
			t.Log("Arguments are nil when set to nil - this is current behavior")
		} else if len(arguments) != 0 {
			t.Errorf("Expected empty arguments, got %v", arguments)
		}
		
		metadata := cmd.GetMetadata()
		if metadata == nil {
			t.Log("Metadata are nil when set to nil - this is current behavior")
		} else if len(metadata) != 0 {
			t.Errorf("Expected empty metadata, got %v", metadata)
		}
		
		// Handler can be nil
		err := cmd.Execute(context.Background(), nil)
		if err == nil {
			t.Error("Expected error when executing command with nil handler")
		}
	})
	
	// Test command metadata handling
	t.Run("command metadata", func(t *testing.T) {
		metadata := map[string]interface{}{
			"version":    "1.0.0",
			"author":     "test",
			"deprecated": false,
			"tags":       []string{"test", "example"},
		}
		
		cmd := NewCommand("test-metadata", "Test Metadata", "Test metadata handling").
			WithMetadata(metadata).
			Build()
		
		cmdMeta := cmd.GetMetadata()
		if cmdMeta["version"] != "1.0.0" {
			t.Errorf("Expected version 1.0.0, got %v", cmdMeta["version"])
		}
		if cmdMeta["author"] != "test" {
			t.Errorf("Expected author test, got %v", cmdMeta["author"])
		}
		if cmdMeta["deprecated"] != false {
			t.Errorf("Expected deprecated false, got %v", cmdMeta["deprecated"])
		}
		
		tags, ok := cmdMeta["tags"].([]string)
		if !ok {
			t.Error("Expected tags to be []string")
		}
		if len(tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(tags))
		}
	})
}

func TestArgumentValidationEnhancements(t *testing.T) {
	// Test argument validation with complex scenarios
	t.Run("complex argument validation", func(t *testing.T) {
		args := []ArgumentDefinition{
			{Name: "required_string", Type: "string", Required: true},
			{Name: "optional_int", Type: "int", Required: false},
			{Name: "required_bool", Type: "bool", Required: true},
			{Name: "optional_array", Type: "[]string", Required: false},
		}
		
		cmd := NewCommand("test-complex-args", "Test Complex Args", "Test complex argument validation").
			WithArguments(args).
			Build()
		
		// Test case 1: All required arguments present
		validArgs := map[string]interface{}{
			"required_string": "test",
			"required_bool":   true,
			"optional_int":    42,
		}
		
		err := cmd.ValidateArgs(validArgs)
		if err != nil {
			t.Errorf("Expected no error for valid args, got: %v", err)
		}
		
		// Test case 2: Missing required string
		invalidArgs1 := map[string]interface{}{
			"required_bool": true,
		}
		
		err = cmd.ValidateArgs(invalidArgs1)
		if err == nil {
			t.Error("Expected error for missing required_string")
		}
		if !strings.Contains(err.Error(), "required_string") {
			t.Errorf("Error should mention required_string, got: %v", err)
		}
		
		// Test case 3: Missing required bool
		invalidArgs2 := map[string]interface{}{
			"required_string": "test",
		}
		
		err = cmd.ValidateArgs(invalidArgs2)
		if err == nil {
			t.Error("Expected error for missing required_bool")
		}
		if !strings.Contains(err.Error(), "required_bool") {
			t.Errorf("Error should mention required_bool, got: %v", err)
		}
		
		// Test case 4: Empty args map
		err = cmd.ValidateArgs(map[string]interface{}{})
		if err == nil {
			t.Error("Expected error for empty args map")
		}
		
		// Test case 5: Nil args map
		err = cmd.ValidateArgs(nil)
		if err == nil {
			t.Error("Expected error for nil args map")
		}
	})
}

func TestConcurrency(t *testing.T) {
	registry := NewCommandRegistry()
	
	// Run registration in parallel
	t.Run("parallel register", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			i := i
			t.Run(fmt.Sprintf("goroutine-%d", i), func(t *testing.T) {
				t.Parallel()
				id := fmt.Sprintf("cmd-%d", i)
				cmd := NewCommand(id, "Test", "A test command").Build()
				err := registry.Register(cmd)
				if err != nil {
					t.Errorf("Failed to register command in parallel: %v", err)
				}
			})
		}
	})

	// Check that all commands were registered
	if len(registry.List()) != 100 {
		t.Errorf("Expected 100 commands, got %d", len(registry.List()))
	}

	// Run get in parallel
	t.Run("parallel get", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			i := i
			t.Run(fmt.Sprintf("goroutine-%d", i), func(t *testing.T) {
				t.Parallel()
				id := fmt.Sprintf("cmd-%d", i)
				if _, exists := registry.Get(id); !exists {
					t.Errorf("Command %s not found in parallel get", id)
				}
			})
		}
	})
}

func TestCommandHandler(t *testing.T) {
	executed := false
	handler := func(ctx context.Context, args map[string]interface{}) error {
		executed = true
		return nil
	}

	cmd := NewCommand("test-handler", "Test Handler", "Tests handler execution").
		WithHandler(handler).
		Build()

	err := cmd.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	if !executed {
		t.Fatal("Command handler was not executed")
	}
}

func TestMissingHandler(t *testing.T) {
	cmd := NewCommand("no-handler", "No Handler", "Tests missing handler").Build()
	err := cmd.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("Expected error for missing handler, but got nil")
	}
}
