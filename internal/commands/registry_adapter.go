// Package commands provides registry integration for scanned commands
package commands

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// ErrCustomCommand is a special error used to signal that a custom command
	// should be executed by the TUI.
	ErrCustomCommand = errors.New("custom command execution")
)

// namedArgPattern is a regex pattern to find named arguments in the format $NAME
var namedArgPattern = regexp.MustCompile(`\$([A-Z][A-Z0-9_]*)`)

// RegistryAdapter adapts scanned commands to work with the command registry system
type RegistryAdapter struct{}

// NewRegistryAdapter creates a new registry adapter
func NewRegistryAdapter() *RegistryAdapter {
	return &RegistryAdapter{}
}

// ConvertToRegistryCommand converts a ParsedCommand to a Command interface implementation
func (ra *RegistryAdapter) ConvertToRegistryCommand(parsed ParsedCommand) Command {
	name := parsed.Metadata.Name
	if name == "" {
		name = generateDisplayName(parsed.ID, parsed.SourceType)
	}

	description := parsed.Metadata.Description
	if description == "" {
		description = fmt.Sprintf("Custom command from %s", parsed.RelativePath)
	}

	var arguments []ArgumentDefinition
	if parsed.Metadata.Arguments != nil {
		arguments = parsed.Metadata.Arguments
	} else {
		arguments = extractArgumentsFromContent(parsed.Content)
	}

	cmd := NewCommand(parsed.ID, name, description).
		WithType(parsed.SourceType).
		WithCategory(parsed.Metadata.Category).
		WithArguments(arguments).
		WithAliases(parsed.Metadata.Aliases).
		WithExample(parsed.Metadata.Example).
		WithHandler(createCommandHandler(parsed)).
		Build()

	return cmd
}

// generateDisplayName generates a display name with appropriate prefix
func generateDisplayName(id string, sourceType CommandType) string {
	switch sourceType {
	case UserCommand:
		return "user:" + id
	case ProjectCommand:
		return "project:" + id
	default:
		return id
	}
}

// extractArgumentsFromContent extracts argument definitions from command content using regex
func extractArgumentsFromContent(content string) []ArgumentDefinition {
	matches := namedArgPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	argMap := make(map[string]bool)
	var arguments []ArgumentDefinition

	for _, match := range matches {
		argName := match[1]
		if !argMap[argName] {
			argMap[argName] = true
			arguments = append(arguments, ArgumentDefinition{
				Name:        argName,
				Description: fmt.Sprintf("Argument %s extracted from command content", argName),
				Type:        "string",
				Required:    true,
			})
		}
	}

	return arguments
}

// createCommandHandler creates a command handler that works with the TUI system
func createCommandHandler(parsed ParsedCommand) CommandHandler {
	return func(ctx context.Context, args map[string]interface{}) error {
		commandContent := parsed.Content
		matches := namedArgPattern.FindAllStringSubmatch(commandContent, -1)
		if len(matches) > 0 {
			processedContent := commandContent
			for argName, value := range args {
				placeholder := "$" + argName
				processedContent = strings.ReplaceAll(processedContent, placeholder, fmt.Sprintf("%v", value))
			}
			commandContent = processedContent
		}
		return fmt.Errorf("%w: %s", ErrCustomCommand, commandContent)
	}
}

// LoadAndRegisterCommands scans for user and project commands and registers them
// in the provided registry.
func LoadAndRegisterCommands(registry Registry) error {
	adapter := NewRegistryAdapter()
	var allErrors []error

	// Scan and register user commands
	userResult, err := ScanUserCommands()
	if err != nil {
		allErrors = append(allErrors, fmt.Errorf("error scanning user commands: %w", err))
	} else {
		for _, parsed := range userResult.Commands {
			cmd := adapter.ConvertToRegistryCommand(parsed)
			if err := registry.Register(cmd); err != nil {
				allErrors = append(allErrors, fmt.Errorf("error registering user command %s: %w", parsed.ID, err))
			}
		}
	}

	// Scan and register project commands
	projectResult, err := ScanProjectCommands(".")
	if err != nil {
		allErrors = append(allErrors, fmt.Errorf("error scanning project commands: %w", err))
	} else {
		for _, parsed := range projectResult.Commands {
			cmd := adapter.ConvertToRegistryCommand(parsed)
			if err := registry.Register(cmd); err != nil {
				allErrors = append(allErrors, fmt.Errorf("error registering project command %s: %w", parsed.ID, err))
			}
		}
	}

	if len(allErrors) > 0 {
		var errorMessages []string
		for _, err := range allErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("errors during command loading: %s", strings.Join(errorMessages, "; "))
	}

	return nil
}

// LoadCustomCommands loads commands from user and project directories and registers them.
func LoadCustomCommands() error {
	adapter := NewRegistryAdapter()
	registry := GetGlobalRegistry()
	var allErrors []error

	// Scan and register user commands
	userResult, err := ScanUserCommands()
	if err != nil {
		allErrors = append(allErrors, fmt.Errorf("error scanning user commands: %w", err))
	} else {
		for _, parsed := range userResult.Commands {
			cmd := adapter.ConvertToRegistryCommand(parsed)
			if err := registry.Register(cmd); err != nil {
				allErrors = append(allErrors, fmt.Errorf("error registering user command %s: %w", parsed.ID, err))
			}
		}
	}

	// Scan and register project commands
	projectResult, err := ScanProjectCommands(".")
	if err != nil {
		allErrors = append(allErrors, fmt.Errorf("error scanning project commands: %w", err))
	} else {
		for _, parsed := range projectResult.Commands {
			cmd := adapter.ConvertToRegistryCommand(parsed)
			if err := registry.Register(cmd); err != nil {
				allErrors = append(allErrors, fmt.Errorf("error registering project command %s: %w", parsed.ID, err))
			}
		}
	}

	if len(allErrors) > 0 {
		var errorMessages []string
		for _, err := range allErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("errors during command loading: %s", strings.Join(errorMessages, "; "))
	}

	return nil
}