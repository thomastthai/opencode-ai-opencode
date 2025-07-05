package commands

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRegistryPerformance(t *testing.T) {
	// Test performance with large number of commands
	t.Run("large command set registration", func(t *testing.T) {
		registry := NewCommandRegistry()
		
		start := time.Now()
		
		// Register 1000 commands
		for i := 0; i < 1000; i++ {
			cmd := NewCommand(fmt.Sprintf("perf-cmd-%d", i), 
				fmt.Sprintf("Performance Command %d", i), 
				fmt.Sprintf("A performance test command %d", i)).
				WithCategory("performance").
				Build()
			
			err := registry.Register(cmd)
			if err != nil {
				t.Fatalf("Failed to register command %d: %v", i, err)
			}
		}
		
		duration := time.Since(start)
		
		// Should complete registration within reasonable time (1 second)
		if duration > time.Second {
			t.Errorf("Registration took too long: %v", duration)
		}
		
		// Verify all commands were registered
		commands := registry.List()
		if len(commands) != 1000 {
			t.Errorf("Expected 1000 commands, got %d", len(commands))
		}
	})
	
	// Test lookup performance
	t.Run("command lookup performance", func(t *testing.T) {
		registry := NewCommandRegistry()
		
		// Register 1000 commands
		for i := 0; i < 1000; i++ {
			cmd := NewCommand(fmt.Sprintf("lookup-cmd-%d", i), 
				fmt.Sprintf("Lookup Command %d", i), 
				"A lookup test command").Build()
			registry.Register(cmd)
		}
		
		start := time.Now()
		
		// Perform 10000 lookups
		for i := 0; i < 10000; i++ {
			cmdID := fmt.Sprintf("lookup-cmd-%d", i%1000)
			_, exists := registry.Get(cmdID)
			if !exists {
				t.Errorf("Command %s should exist", cmdID)
			}
		}
		
		duration := time.Since(start)
		
		// 10000 lookups should complete quickly (100ms)
		if duration > 100*time.Millisecond {
			t.Errorf("Lookups took too long: %v", duration)
		}
	})
	
	// Test concurrent access performance
	t.Run("concurrent access performance", func(t *testing.T) {
		registry := NewCommandRegistry()
		
		// Pre-populate with some commands
		for i := 0; i < 100; i++ {
			cmd := NewCommand(fmt.Sprintf("concurrent-cmd-%d", i), 
				fmt.Sprintf("Concurrent Command %d", i), 
				"A concurrent test command").Build()
			registry.Register(cmd)
		}
		
		var wg sync.WaitGroup
		errors := make(chan error, 100)
		
		start := time.Now()
		
		// Launch 100 concurrent readers
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				// Each goroutine performs multiple operations
				for j := 0; j < 100; j++ {
					cmdID := fmt.Sprintf("concurrent-cmd-%d", j%100)
					_, exists := registry.Get(cmdID)
					if !exists {
						errors <- fmt.Errorf("command %s should exist", cmdID)
						return
					}
				}
			}(i)
		}
		
		wg.Wait()
		close(errors)
		
		duration := time.Since(start)
		
		// Check for errors
		for err := range errors {
			t.Error(err)
		}
		
		// Concurrent operations should complete quickly (500ms)
		if duration > 500*time.Millisecond {
			t.Errorf("Concurrent access took too long: %v", duration)
		}
	})
}

func TestCommandExecutionPerformance(t *testing.T) {
	t.Run("command execution overhead", func(t *testing.T) {
		executed := false
		handler := func(ctx context.Context, args map[string]interface{}) error {
			executed = true
			return nil
		}
		
		cmd := NewCommand("perf-exec", "Performance Execution", "Test execution performance").
			WithHandler(handler).
			Build()
		
		// Warm up
		cmd.Execute(context.Background(), nil)
		executed = false
		
		start := time.Now()
		
		// Execute command 1000 times
		for i := 0; i < 1000; i++ {
			err := cmd.Execute(context.Background(), nil)
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}
		}
		
		duration := time.Since(start)
		
		// 1000 executions should complete quickly (10ms)
		if duration > 10*time.Millisecond {
			t.Errorf("Command executions took too long: %v", duration)
		}
		
		if !executed {
			t.Error("Handler should have been executed")
		}
	})
}

func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}
	
	t.Run("memory usage with large command set", func(t *testing.T) {
		registry := NewCommandRegistry()
		
		// Register a large number of commands
		numCommands := 10000
		for i := 0; i < numCommands; i++ {
			cmd := NewCommand(fmt.Sprintf("mem-cmd-%d", i), 
				fmt.Sprintf("Memory Command %d", i), 
				fmt.Sprintf("A memory test command with longer description %d", i)).
				WithCategory("memory-test").
				WithAliases([]string{fmt.Sprintf("mc%d", i), fmt.Sprintf("mem%d", i)}).
				Build()
			
			err := registry.Register(cmd)
			if err != nil {
				t.Fatalf("Failed to register command %d: %v", i, err)
			}
		}
		
		// Verify all commands are accessible
		commands := registry.List()
		if len(commands) != numCommands {
			t.Errorf("Expected %d commands, got %d", numCommands, len(commands))
		}
		
		// Perform some operations to ensure data is not optimized away
		totalAliases := 0
		for _, cmd := range commands {
			totalAliases += len(cmd.GetAliases())
		}
		
		if totalAliases != numCommands*2 { // Each command has 2 aliases
			t.Errorf("Expected %d total aliases, got %d", numCommands*2, totalAliases)
		}
	})
}

func BenchmarkCommandRegistration(b *testing.B) {
	registry := NewCommandRegistry()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		cmd := NewCommand(fmt.Sprintf("bench-cmd-%d", i), 
			fmt.Sprintf("Benchmark Command %d", i), 
			"A benchmark test command").Build()
		registry.Register(cmd)
	}
}

func BenchmarkCommandLookup(b *testing.B) {
	registry := NewCommandRegistry()
	
	// Pre-populate registry
	for i := 0; i < 1000; i++ {
		cmd := NewCommand(fmt.Sprintf("bench-lookup-%d", i), 
			fmt.Sprintf("Benchmark Lookup %d", i), 
			"A benchmark lookup command").Build()
		registry.Register(cmd)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		cmdID := fmt.Sprintf("bench-lookup-%d", i%1000)
		registry.Get(cmdID)
	}
}

func BenchmarkCommandExecution(b *testing.B) {
	handler := func(ctx context.Context, args map[string]interface{}) error {
		return nil
	}
	
	cmd := NewCommand("bench-exec", "Benchmark Execution", "Benchmark execution test").
		WithHandler(handler).
		Build()
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		cmd.Execute(ctx, nil)
	}
}