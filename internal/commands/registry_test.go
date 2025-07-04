package commands

import (
	"context"
	"fmt"
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
