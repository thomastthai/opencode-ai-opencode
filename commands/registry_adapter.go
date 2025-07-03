// Package commands provides registry integration for scanned commands
package commands

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opencode-ai/opencode/internal/tui/util"
)

// namedArgPattern is a regex pattern to find named arguments in the format $NAME
var namedArgPattern = regexp.MustCompile(`\$([A-Z][A-Z0-9_]*)`)

// CommandRunCustomMsg is sent when a custom command is executed
type CommandRunCustomMsg struct {
	Content string
	Args    map[string]string // Map of argument names to values
}

// ShowMultiArgumentsDialogMsg is sent when a command needs multiple arguments
type ShowMultiArgumentsDialogMsg struct {
	CommandID string
	Content   string
	ArgNames  []string
}

// RegistryAdapter adapts scanned commands to work with the command registry system
type RegistryAdapter struct{}

// NewRegistryAdapter creates a new registry adapter
func NewRegistryAdapter() *RegistryAdapter {
	return &RegistryAdapter{}
}

// ConvertToRegistryCommand converts a ParsedCommand to a Command interface implementation
func (ra *RegistryAdapter) ConvertToRegistryCommand(parsed ParsedCommand) Command {
	// Create metadata map
	metadata := make(map[string]interface{})
	
	// Add source information
	metadata["source_type"] = parsed.SourceType.String()
	metadata["file_path"] = parsed.FilePath
	metadata["relative_path"] = parsed.RelativePath
	
	// Add custom metadata from frontmatter
	for k, v := range parsed.Metadata.Custom {
		metadata[k] = v
	}
	
	// Determine display name
	name := parsed.Metadata.Name
	if name == "" {
		name = ra.generateDisplayName(parsed.ID, parsed.SourceType)
	}
	
	// Determine description
	description := parsed.Metadata.Description
	if description == "" {
		description = fmt.Sprintf("Custom command from %s", parsed.RelativePath)
	}
	
	// Convert arguments
	var arguments []ArgumentDefinition
	if parsed.Metadata.Arguments != nil {
		arguments = parsed.Metadata.Arguments
	} else {
		// Extract arguments from content if not defined in frontmatter
		arguments = ra.extractArgumentsFromContent(parsed.Content)
	}
	
	// Create the command
	cmd := NewCommand(parsed.ID, name, description).
		WithType(parsed.SourceType).
		WithCategory(parsed.Metadata.Category).
		WithArguments(arguments).
		WithAliases(parsed.Metadata.Aliases).
		WithExample(parsed.Metadata.Example).
		WithMetadata(metadata).
		WithHandler(ra.createCommandHandler(parsed)).
		Build()
	
	if parsed.Metadata.Hidden {
		cmd.hidden = true
	}
	
	return cmd
}

// generateDisplayName generates a display name with appropriate prefix
func (ra *RegistryAdapter) generateDisplayName(id string, sourceType CommandType) string {
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
func (ra *RegistryAdapter) extractArgumentsFromContent(content string) []ArgumentDefinition {
	matches := namedArgPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	
	// Extract unique argument names
	argMap := make(map[string]bool)
	var arguments []ArgumentDefinition
	
	for _, match := range matches {
		argName := match[1] // Group 1 is the name without $
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
func (ra *RegistryAdapter) createCommandHandler(parsed ParsedCommand) CommandHandler {
	return func(ctx context.Context, args map[string]interface{}) error {
		commandContent := parsed.Content
		
		// Check for named arguments in content
		matches := namedArgPattern.FindAllStringSubmatch(commandContent, -1)
		if len(matches) > 0 {
			// Extract unique argument names
			argNames := make([]string, 0)
			argMap := make(map[string]bool)
			
			for _, match := range matches {
				argName := match[1] // Group 1 is the name without $
				if !argMap[argName] {
					argMap[argName] = true
					argNames = append(argNames, argName)
				}
			}
			
			// Check if we have all required arguments
			missingArgs := make([]string, 0)
			for _, argName := range argNames {
				if _, exists := args[argName]; !exists {
					missingArgs = append(missingArgs, argName)
				}
			}
			
			// If we have missing args, request them via the TUI
			if len(missingArgs) > 0 {
				// Note: This is a simplified approach. In a real implementation,
				// this would need to be handled by the TUI system differently
				return fmt.Errorf("missing required arguments: %v", missingArgs)
			}
			
			// Replace arguments in content
			processedContent := commandContent
			for argName, value := range args {
				placeholder := "$" + argName
				processedContent = strings.ReplaceAll(processedContent, placeholder, fmt.Sprintf("%v", value))
			}
			
			commandContent = processedContent
		}
		
		// For now, we'll return an error indicating that the command content should be processed
		// In the actual TUI integration, this would trigger the appropriate message handling
		return fmt.Errorf("EXECUTE_CUSTOM_COMMAND: %s", commandContent)
	}
}

// LoadAndRegisterCommands loads commands from directories and registers them in the registry
func (ra *RegistryAdapter) LoadAndRegisterCommands(registry CommandRegistry) error {
	var allErrors []error
	
	// Scan user commands
	userResult, err := ScanUserCommands()
	if err != nil {
		allErrors = append(allErrors, fmt.Errorf("error scanning user commands: %w", err))
	} else {
		// Register user commands
		for _, parsed := range userResult.Commands {
			cmd := ra.ConvertToRegistryCommand(parsed)
			if err := registry.Register(cmd); err != nil {
				allErrors = append(allErrors, fmt.Errorf("error registering user command %s: %w", parsed.ID, err))
			}
		}
		
		// Log scanning errors
		for _, scanErr := range userResult.Errors {
			allErrors = append(allErrors, fmt.Errorf("user command scan error: %w", scanErr))
		}
	}
	
	// Scan project commands (we'll need project directory from config)
	// For now, we'll try to get it from the current working directory
	// This could be enhanced to use the actual project directory from config
	projectResult, err := ScanProjectCommands(".")
	if err != nil {
		allErrors = append(allErrors, fmt.Errorf("error scanning project commands: %w", err))
	} else {
		// Register project commands
		for _, parsed := range projectResult.Commands {
			cmd := ra.ConvertToRegistryCommand(parsed)
			if err := registry.Register(cmd); err != nil {
				allErrors = append(allErrors, fmt.Errorf("error registering project command %s: %w", parsed.ID, err))
			}
		}
		
		// Log scanning errors
		for _, scanErr := range projectResult.Errors {
			allErrors = append(allErrors, fmt.Errorf("project command scan error: %w", scanErr))
		}
	}
	
	// Return combined errors if any
	if len(allErrors) > 0 {
		var errorMessages []string
		for _, err := range allErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("errors during command loading: %s", strings.Join(errorMessages, "; "))
	}
	
	return nil
}

// LoadAndRegisterCommandsInGlobalRegistry loads and registers commands in the global registry
func LoadAndRegisterCommandsInGlobalRegistry() error {
	adapter := NewRegistryAdapter()
	return adapter.LoadAndRegisterCommands(GetGlobalRegistry())
}

// TUICommandAdapter provides compatibility with the existing TUI command system
type TUICommandAdapter struct {
	adapter *RegistryAdapter
}

// NewTUICommandAdapter creates a new TUI command adapter
func NewTUICommandAdapter() *TUICommandAdapter {
	return &TUICommandAdapter{
		adapter: NewRegistryAdapter(),
	}
}

// TUICommand represents a command in the TUI system format
type TUICommand struct {
	ID          string
	Title       string
	Description string
	Handler     func(cmd TUICommand) tea.Cmd
}

// ConvertToTUICommand converts a ParsedCommand to a TUI-compatible command
func (tca *TUICommandAdapter) ConvertToTUICommand(parsed ParsedCommand) TUICommand {
	// Generate display name with prefix
	title := tca.adapter.generateDisplayName(parsed.ID, parsed.SourceType)
	if parsed.Metadata.Name != "" {
		title = parsed.Metadata.Name
	}
	
	// Use metadata description or generate one
	description := parsed.Metadata.Description
	if description == "" {
		description = fmt.Sprintf("Custom command from %s", parsed.RelativePath)
	}
	
	return TUICommand{
		ID:          title, // Use prefixed title as ID for TUI compatibility
		Title:       title,
		Description: description,
		Handler:     tca.createTUIHandler(parsed),
	}
}

// createTUIHandler creates a TUI-compatible handler for the command
func (tca *TUICommandAdapter) createTUIHandler(parsed ParsedCommand) func(cmd TUICommand) tea.Cmd {
	return func(cmd TUICommand) tea.Cmd {
		commandContent := parsed.Content
		
		// Check for named arguments
		matches := namedArgPattern.FindAllStringSubmatch(commandContent, -1)
		if len(matches) > 0 {
			// Extract unique argument names
			argNames := make([]string, 0)
			argMap := make(map[string]bool)
			
			for _, match := range matches {
				argName := match[1] // Group 1 is the name without $
				if !argMap[argName] {
					argMap[argName] = true
					argNames = append(argNames, argName)
				}
			}
			
			// Show multi-arguments dialog for all named arguments
			return util.CmdHandler(ShowMultiArgumentsDialogMsg{
				CommandID: cmd.ID,
				Content:   commandContent,
				ArgNames:  argNames,
			})
		}
		
		// No arguments needed, run command directly
		return util.CmdHandler(CommandRunCustomMsg{
			Content: commandContent,
			Args:    nil, // No arguments
		})
	}
}

// LoadTUICommands loads commands and returns them in TUI-compatible format
func (tca *TUICommandAdapter) LoadTUICommands() ([]TUICommand, error) {
	commands := make([]TUICommand, 0) // Always initialize as non-nil
	var allErrors []error
	
	// Load user commands
	userResult, err := ScanUserCommands()
	if err != nil {
		allErrors = append(allErrors, fmt.Errorf("error scanning user commands: %w", err))
	} else {
		for _, parsed := range userResult.Commands {
			commands = append(commands, tca.ConvertToTUICommand(parsed))
		}
		allErrors = append(allErrors, userResult.Errors...)
	}
	
	// Load project commands
	projectResult, err := ScanProjectCommands(".")
	if err != nil {
		allErrors = append(allErrors, fmt.Errorf("error scanning project commands: %w", err))
	} else {
		for _, parsed := range projectResult.Commands {
			commands = append(commands, tca.ConvertToTUICommand(parsed))
		}
		allErrors = append(allErrors, projectResult.Errors...)
	}
	
	// Return commands even if there were some errors
	if len(allErrors) > 0 {
		var errorMessages []string
		for _, err := range allErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		return commands, fmt.Errorf("errors during command loading: %s", strings.Join(errorMessages, "; "))
	}
	
	return commands, nil
}