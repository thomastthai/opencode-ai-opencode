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
	err = LoadAndRegisterCommands(GetGlobalRegistry())
	if err != nil {
		fmt.Printf("Warning during command loading: %v\n", err)
	}

	// 4. Show registry contents
	fmt.Println("\n4. Commands in registry:")
	commands := GetGlobalRegistry().List()
	fmt.Printf("Total commands registered: %d\n", len(commands))
	for _, cmd := range commands {
		fmt.Printf("  - %s (%s): %s\n", cmd.ID(), cmd.Type().String(), cmd.Description())
	}

	return nil
}

// RunExample runs the integration example if called directly
func RunExample() {
	if len(os.Args) > 1 && os.Args[1] == "example" {
		if err := IntegrationExample(); err != nil {
			fmt.Printf("Error running example: %v\n", err)
			os.Exit(1)
		}
	}
}
