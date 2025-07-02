package commands

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// CommandType represents the type/category of a command
type CommandType int

const (
	// BuiltinCommand represents system built-in commands
	BuiltinCommand CommandType = iota
	// UserCommand represents user-defined commands
	UserCommand
	// ProjectCommand represents project-specific commands
	ProjectCommand
	// PluginCommand represents plugin-provided commands
	PluginCommand
)

// String returns the string representation of the command type
func (ct CommandType) String() string {
	switch ct {
	case BuiltinCommand:
		return "builtin"
	case UserCommand:
		return "user"
	case ProjectCommand:
		return "project"
	case PluginCommand:
		return "plugin"
	default:
		return "unknown"
	}
}

// ArgumentDefinition defines a command argument
type ArgumentDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`        // "string", "int", "bool", "file", etc.
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"` // For enum-like arguments
}

// CommandContext provides context information for command execution
type CommandContext struct {
	WorkingDir    string                 `json:"working_dir"`
	ProjectPath   string                 `json:"project_path"`
	UserConfig    map[string]interface{} `json:"user_config,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
}

// CommandSuggestion represents an autocomplete suggestion
type CommandSuggestion struct {
	Command     Command `json:"command"`
	Score       float64 `json:"score"`       // Relevance score for ranking
	MatchedText string  `json:"matched_text"` // The part of the command that matched
}

// Command interface defines the contract that all commands must implement
type Command interface {
	// ID returns the unique identifier for the command
	ID() string
	
	// Name returns the display name of the command
	Name() string
	
	// Description returns a detailed description of what the command does
	Description() string
	
	// Category returns the category this command belongs to
	Category() string
	
	// Type returns the type of command (builtin, user, project, plugin)
	Type() CommandType
	
	// Execute runs the command with the given arguments
	Execute(ctx context.Context, args map[string]interface{}) error
	
	// ValidateArgs validates the provided arguments
	ValidateArgs(args map[string]interface{}) error
	
	// GetArguments returns the argument definitions for this command
	GetArguments() []ArgumentDefinition
	
	// GetSubCommands returns any sub-commands this command may have
	GetSubCommands() []Command
	
	// IsHidden returns true if this command should be hidden from general listings
	IsHidden() bool
	
	// GetAliases returns alternative names for this command
	GetAliases() []string
	
	// GetExample returns usage examples for this command
	GetExample() string
}

// BaseCommand provides a basic implementation of the Command interface
// that other commands can embed to reduce boilerplate
type BaseCommand struct {
	id          string
	name        string
	description string
	category    string
	commandType CommandType
	arguments   []ArgumentDefinition
	subCommands []Command
	hidden      bool
	aliases     []string
	example     string
}

// ID returns the unique identifier for the command
func (bc *BaseCommand) ID() string {
	return bc.id
}

// Name returns the display name of the command
func (bc *BaseCommand) Name() string {
	return bc.name
}

// Description returns a detailed description of what the command does
func (bc *BaseCommand) Description() string {
	return bc.description
}

// Category returns the category this command belongs to
func (bc *BaseCommand) Category() string {
	return bc.category
}

// Type returns the type of command
func (bc *BaseCommand) Type() CommandType {
	return bc.commandType
}

// GetArguments returns the argument definitions for this command
func (bc *BaseCommand) GetArguments() []ArgumentDefinition {
	return bc.arguments
}

// GetSubCommands returns any sub-commands this command may have
func (bc *BaseCommand) GetSubCommands() []Command {
	return bc.subCommands
}

// IsHidden returns true if this command should be hidden from general listings
func (bc *BaseCommand) IsHidden() bool {
	return bc.hidden
}

// GetAliases returns alternative names for this command
func (bc *BaseCommand) GetAliases() []string {
	return bc.aliases
}

// GetExample returns usage examples for this command
func (bc *BaseCommand) GetExample() string {
	return bc.example
}

// Execute provides a default implementation that returns an error
// Commands should override this method
func (bc *BaseCommand) Execute(ctx context.Context, args map[string]interface{}) error {
	return fmt.Errorf("command %s does not implement Execute method", bc.id)
}

// ValidateArgs provides a basic validation implementation
// Commands can override this for custom validation
func (bc *BaseCommand) ValidateArgs(args map[string]interface{}) error {
	for _, arg := range bc.arguments {
		value, exists := args[arg.Name]
		
		// Check required arguments
		if arg.Required && !exists {
			return fmt.Errorf("required argument '%s' is missing", arg.Name)
		}
		
		// Skip validation for optional arguments that aren't provided
		if !exists {
			continue
		}
		
		// Basic type validation
		switch arg.Type {
		case "string":
			if _, ok := value.(string); !ok {
				return fmt.Errorf("argument '%s' must be a string", arg.Name)
			}
		case "int":
			if _, ok := value.(int); !ok {
				return fmt.Errorf("argument '%s' must be an integer", arg.Name)
			}
		case "bool":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("argument '%s' must be a boolean", arg.Name)
			}
		}
		
		// Validate enum options
		if len(arg.Options) > 0 {
			valueStr := fmt.Sprintf("%v", value)
			found := false
			for _, option := range arg.Options {
				if option == valueStr {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("argument '%s' must be one of: %v", arg.Name, arg.Options)
			}
		}
	}
	
	return nil
}

// CommandRegistry interface defines the contract for command registration and discovery
type CommandRegistry interface {
	// Register adds a command to the registry
	Register(cmd Command) error
	
	// Unregister removes a command from the registry
	Unregister(id string) error
	
	// Get retrieves a command by its ID
	Get(id string) (Command, bool)
	
	// List returns all registered commands
	List() []Command
	
	// Search finds commands matching the given query
	Search(query string) []CommandSuggestion
	
	// GetByCategory returns commands in a specific category
	GetByCategory(category string) []Command
	
	// GetByType returns commands of a specific type
	GetByType(commandType CommandType) []Command
	
	// Clear removes all commands from the registry (primarily for testing)
	Clear()
}

// DefaultCommandRegistry provides a thread-safe implementation of CommandRegistry
type DefaultCommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]Command
	aliases  map[string]string // alias -> command ID mapping
}

// NewCommandRegistry creates a new command registry instance
func NewCommandRegistry() CommandRegistry {
	return &DefaultCommandRegistry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

// Register adds a command to the registry
func (r *DefaultCommandRegistry) Register(cmd Command) error {
	if cmd == nil {
		return errors.New("command cannot be nil")
	}
	
	if cmd.ID() == "" {
		return errors.New("command ID cannot be empty")
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if command already exists
	if _, exists := r.commands[cmd.ID()]; exists {
		return fmt.Errorf("command with ID '%s' already registered", cmd.ID())
	}
	
	// Register the command
	r.commands[cmd.ID()] = cmd
	
	// Register aliases
	for _, alias := range cmd.GetAliases() {
		if alias == "" {
			continue
		}
		if existingID, exists := r.aliases[alias]; exists {
			return fmt.Errorf("alias '%s' already registered for command '%s'", alias, existingID)
		}
		r.aliases[alias] = cmd.ID()
	}
	
	return nil
}

// Unregister removes a command from the registry
func (r *DefaultCommandRegistry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	cmd, exists := r.commands[id]
	if !exists {
		return fmt.Errorf("command '%s' not found", id)
	}
	
	// Remove aliases
	for _, alias := range cmd.GetAliases() {
		delete(r.aliases, alias)
	}
	
	// Remove command
	delete(r.commands, id)
	
	return nil
}

// Get retrieves a command by its ID or alias
func (r *DefaultCommandRegistry) Get(id string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Check direct ID match first
	if cmd, exists := r.commands[id]; exists {
		return cmd, true
	}
	
	// Check aliases
	if realID, exists := r.aliases[id]; exists {
		if cmd, exists := r.commands[realID]; exists {
			return cmd, true
		}
	}
	
	return nil, false
}

// List returns all registered commands
func (r *DefaultCommandRegistry) List() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	commands := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}
	
	return commands
}

// Search finds commands matching the given query
// This is a basic implementation - can be enhanced with fuzzy search later
func (r *DefaultCommandRegistry) Search(query string) []CommandSuggestion {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	suggestions := make([]CommandSuggestion, 0)
	
	for _, cmd := range r.commands {
		score := r.calculateMatchScore(query, cmd)
		if score > 0 {
			suggestions = append(suggestions, CommandSuggestion{
				Command:     cmd,
				Score:       score,
				MatchedText: query,
			})
		}
	}
	
	// Sort by score (basic implementation)
	// TODO: Implement proper sorting
	
	return suggestions
}

// calculateMatchScore provides basic scoring for command matching
// TODO: Implement proper fuzzy matching algorithm
func (r *DefaultCommandRegistry) calculateMatchScore(query string, cmd Command) float64 {
	query = fmt.Sprintf("%s", query) // Convert to lowercase equivalent when implementing
	
	// Exact ID match
	if cmd.ID() == query {
		return 1.0
	}
	
	// Exact name match
	if cmd.Name() == query {
		return 0.9
	}
	
	// Alias match
	for _, alias := range cmd.GetAliases() {
		if alias == query {
			return 0.8
		}
	}
	
	// Basic substring matching - will be replaced with fuzzy search
	// This is just placeholder logic
	return 0.0
}

// GetByCategory returns commands in a specific category
func (r *DefaultCommandRegistry) GetByCategory(category string) []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	commands := make([]Command, 0)
	for _, cmd := range r.commands {
		if cmd.Category() == category {
			commands = append(commands, cmd)
		}
	}
	
	return commands
}

// GetByType returns commands of a specific type
func (r *DefaultCommandRegistry) GetByType(commandType CommandType) []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	commands := make([]Command, 0)
	for _, cmd := range r.commands {
		if cmd.Type() == commandType {
			commands = append(commands, cmd)
		}
	}
	
	return commands
}

// Clear removes all commands from the registry
func (r *DefaultCommandRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.commands = make(map[string]Command)
	r.aliases = make(map[string]string)
}

// Global registry instance
var globalRegistry CommandRegistry = NewCommandRegistry()

// RegisterCommand registers a command in the global registry
func RegisterCommand(cmd Command) error {
	return globalRegistry.Register(cmd)
}

// GetCommand retrieves a command from the global registry
func GetCommand(id string) (Command, bool) {
	return globalRegistry.Get(id)
}

// ListCommands returns all commands from the global registry
func ListCommands() []Command {
	return globalRegistry.List()
}

// SearchCommands searches for commands in the global registry
func SearchCommands(query string) []CommandSuggestion {
	return globalRegistry.Search(query)
}

// GetCommandsByCategory returns commands by category from the global registry
func GetCommandsByCategory(category string) []Command {
	return globalRegistry.GetByCategory(category)
}

// GetCommandsByType returns commands by type from the global registry
func GetCommandsByType(commandType CommandType) []Command {
	return globalRegistry.GetByType(commandType)
}

// GetGlobalRegistry returns the global command registry instance
// Primarily for testing and advanced use cases
func GetGlobalRegistry() CommandRegistry {
	return globalRegistry
}

// SetGlobalRegistry sets the global command registry instance
// Primarily for testing
func SetGlobalRegistry(registry CommandRegistry) {
	globalRegistry = registry
}