// Package commands provides an integration example showing how to use the new scanning
// functionality with the existing TUI system.
//
// This example demonstrates how to load custom commands with YAML frontmatter
// and integrate them with both the new registry system and the existing TUI
// command dialog.
package commands

import (
	"fmt"
	"os"
)

// IntegrationExample demonstrates how to integrate the scanning functionality
// with the existing OpenCode system.
func IntegrationExample() error {
	fmt.Println("OpenCode Command Scanner Integration Example")
	fmt.Println("==========================================")

	// 1. Load commands using the new scanning system
	fmt.Println("\n1. Scanning user commands...")
	userResult, err := ScanUserCommands()
	if err != nil {
		fmt.Printf("Error scanning user commands: %v\n", err)
	} else {
		fmt.Printf("Found %d user commands with %d errors\n",
			len(userResult.Commands), len(userResult.Errors))
		for _, cmd := range userResult.Commands {
			fmt.Printf("  - %s: %s\n", cmd.ID, cmd.Metadata.Name)
		}
	}

	// 2. Load project commands
	fmt.Println("\n2. Scanning project commands...")
	projectResult, err := ScanProjectCommands(".")
	if err != nil {
		fmt.Printf("Error scanning project commands: %v\n", err)
	} else {
		fmt.Printf("Found %d project commands with %d errors\n",
			len(projectResult.Commands), len(projectResult.Errors))
		for _, cmd := range projectResult.Commands {
			fmt.Printf("  - %s: %s\n", cmd.ID, cmd.Metadata.Name)
		}
	}

	// 3. Register commands in the global registry
	fmt.Println("\n3. Loading commands into registry...")
	err = LoadAndRegisterCommandsInGlobalRegistry()
	if err != nil {
		fmt.Printf("Warning during command loading: %v\n", err)
	}

	// 4. Show registry contents
	fmt.Println("\n4. Commands in registry:")
	commands := ListCommands()
	fmt.Printf("Total commands registered: %d\n", len(commands))
	for _, cmd := range commands {
		fmt.Printf("  - %s (%s): %s\n", cmd.ID(), cmd.Type().String(), cmd.Description())
	}

	// 5. Load TUI-compatible commands
	fmt.Println("\n5. Loading TUI commands...")
	tuiAdapter := NewTUICommandAdapter()
	tuiCommands, err := tuiAdapter.LoadTUICommands()
	if err != nil {
		fmt.Printf("Warning during TUI command loading: %v\n", err)
	}
	fmt.Printf("TUI commands loaded: %d\n", len(tuiCommands))
	for _, cmd := range tuiCommands {
		fmt.Printf("  - %s: %s\n", cmd.ID, cmd.Description)
	}

	return nil
}

// TUIIntegrationHelpers provides helper functions for integrating with the existing TUI
type TUIIntegrationHelpers struct {
	adapter *TUICommandAdapter
}

// NewTUIIntegrationHelpers creates new TUI integration helpers
func NewTUIIntegrationHelpers() *TUIIntegrationHelpers {
	return &TUIIntegrationHelpers{
		adapter: NewTUICommandAdapter(),
	}
}

// LoadCommandsForTUI loads all custom commands and returns them in the format
// expected by the existing TUI command dialog
func (h *TUIIntegrationHelpers) LoadCommandsForTUI() ([]TUICommand, error) {
	return h.adapter.LoadTUICommands()
}

// GetCommandMetadata returns metadata for a command by ID
func (h *TUIIntegrationHelpers) GetCommandMetadata(commandID string) (CommandMetadata, error) {
	// Scan all commands to find the one with the given ID
	userResult, err := ScanUserCommands()
	if err == nil {
		for _, cmd := range userResult.Commands {
			if cmd.ID == commandID {
				return cmd.Metadata, nil
			}
		}
	}

	projectResult, err := ScanProjectCommands(".")
	if err == nil {
		for _, cmd := range projectResult.Commands {
			if cmd.ID == commandID {
				return cmd.Metadata, nil
			}
		}
	}

	return CommandMetadata{}, fmt.Errorf("command not found: %s", commandID)
}

// ListCommandsByCategory returns commands grouped by category
func (h *TUIIntegrationHelpers) ListCommandsByCategory() (map[string][]TUICommand, error) {
	commands, err := h.LoadCommandsForTUI()
	if err != nil {
		return nil, err
	}

	categories := make(map[string][]TUICommand)

	for _, cmd := range commands {
		// Get the command metadata to find its category
		metadata, err := h.GetCommandMetadata(cmd.ID)
		if err != nil {
			continue // Skip commands we can't get metadata for
		}

		category := metadata.Category
		if category == "" {
			category = "General"
		}

		categories[category] = append(categories[category], cmd)
	}

	return categories, nil
}

// CompatibilityLayer provides a compatibility layer for existing command loading
type CompatibilityLayer struct {
	helpers *TUIIntegrationHelpers
}

// NewCompatibilityLayer creates a new compatibility layer
func NewCompatibilityLayer() *CompatibilityLayer {
	return &CompatibilityLayer{
		helpers: NewTUIIntegrationHelpers(),
	}
}

// LoadCustomCommandsCompat loads custom commands in the format expected by the existing
// LoadCustomCommands function, providing backward compatibility
func (c *CompatibilityLayer) LoadCustomCommandsCompat() ([]TUICommand, error) {
	return c.helpers.LoadCommandsForTUI()
}

// Example of how to integrate with existing TUI code:
//
// Replace the existing LoadCustomCommands() call with:
//
//   compat := commands.NewCompatibilityLayer()
//   customCommands, err := compat.LoadCustomCommandsCompat()
//   if err != nil {
//       logging.Warn("Failed to load custom commands", "error", err)
//   } else {
//       for _, cmd := range customCommands {
//           model.RegisterCommand(cmd)
//       }
//   }

// RunExample runs the integration example if called directly
func RunExample() {
	if len(os.Args) > 1 && os.Args[1] == "example" {
		if err := IntegrationExample(); err != nil {
			fmt.Printf("Error running example: %v\n", err)
			os.Exit(1)
		}
	}
}