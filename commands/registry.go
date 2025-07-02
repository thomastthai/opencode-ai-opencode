// Package commands provides a comprehensive command registry system for OpenCode.
//
// This package implements a hierarchical command structure that supports:
//   - Built-in system commands
//   - User-defined commands
//   - Project-specific commands
//   - Plugin-provided commands
//
// # Core Components
//
// Command Interface: Defines the contract that all commands must implement,
// including execution, validation, metadata, and hierarchical relationships.
//
// BaseCommand: A concrete implementation of the Command interface that other
// commands can embed to reduce boilerplate code. Supports command handlers,
// metadata, sub-commands, and parent relationships.
//
// CommandRegistry: Thread-safe registry for command registration, discovery,
// and management. Supports hierarchical registration and path-based lookup.
//
// CommandBuilder: Fluent interface for constructing commands with various
// options and configurations.
//
// # Usage Example
//
//	// Create a command with sub-commands
//	subCmd := NewCommand("commit", "Commit", "Commit changes").
//		WithType(BuiltinCommand).
//		WithHandler(func(ctx context.Context, args map[string]interface{}) error {
//			// Handle commit logic
//			return nil
//		}).
//		Build()
//
//	parentCmd := NewCommand("git", "Git", "Git version control").
//		WithType(BuiltinCommand).
//		AddSubCommand(subCmd).
//		Build()
//
//	// Register the hierarchy
//	err := RegisterCommandHierarchy(parentCmd)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Find command by path
//	cmd, found := GetCommandByPath("git commit")
//	if found {
//		err := cmd.Execute(context.Background(), args)
//	}
//
// # Thread Safety
//
// All registry operations are thread-safe using read-write mutexes.
// Command instances themselves are not inherently thread-safe and should
// be designed with concurrent execution in mind.
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
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`        // "string", "int", "bool", "file", etc.
	Required    bool     `json:"required"`
	Default     string   `json:"default,omitempty"`
	Options     []string `json:"options,omitempty"` // For enum-like arguments
}

// CommandHandler represents a function that executes a command
// This provides flexibility for different execution patterns
type CommandHandler func(ctx context.Context, args map[string]interface{}) error

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
	
	// GetMetadata returns command-specific metadata
	GetMetadata() map[string]interface{}
	
	// GetParent returns the parent command (nil for root commands)
	GetParent() Command
	
	// GetPath returns the full command path (e.g., "git commit --amend")
	GetPath() string
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
	handler     CommandHandler                 // Explicit handler function
	metadata    map[string]interface{}         // Extensible metadata storage
	parent      Command                        // Parent command reference
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

// GetMetadata returns command-specific metadata
func (bc *BaseCommand) GetMetadata() map[string]interface{} {
	if bc.metadata == nil {
		return make(map[string]interface{})
	}
	// Return a copy to prevent external modification
	metadata := make(map[string]interface{})
	for k, v := range bc.metadata {
		metadata[k] = v
	}
	return metadata
}

// GetParent returns the parent command (nil for root commands)
func (bc *BaseCommand) GetParent() Command {
	return bc.parent
}

// GetPath returns the full command path (e.g., "git commit --amend")
func (bc *BaseCommand) GetPath() string {
	if bc.parent == nil {
		return bc.id
	}
	return bc.parent.GetPath() + " " + bc.id
}

// Execute provides a default implementation that uses the handler if available
// Commands can override this method or set a handler function
func (bc *BaseCommand) Execute(ctx context.Context, args map[string]interface{}) error {
	if bc.handler != nil {
		return bc.handler(ctx, args)
	}
	return fmt.Errorf("command %s does not implement Execute method or have a handler", bc.id)
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

// CommandBuilder provides a fluent interface for building commands
type CommandBuilder struct {
	command *BaseCommand
}

// NewCommand creates a new command builder with required fields
func NewCommand(id, name, description string) *CommandBuilder {
	return &CommandBuilder{
		command: &BaseCommand{
			id:          id,
			name:        name,
			description: description,
			arguments:   []ArgumentDefinition{},
			subCommands: []Command{},
			aliases:     []string{},
			metadata:    make(map[string]interface{}),
		},
	}
}

// WithCategory sets the command category
func (cb *CommandBuilder) WithCategory(category string) *CommandBuilder {
	cb.command.category = category
	return cb
}

// WithType sets the command type
func (cb *CommandBuilder) WithType(commandType CommandType) *CommandBuilder {
	cb.command.commandType = commandType
	return cb
}

// WithHandler sets the command handler function
func (cb *CommandBuilder) WithHandler(handler CommandHandler) *CommandBuilder {
	cb.command.handler = handler
	return cb
}

// WithArguments sets the command arguments
func (cb *CommandBuilder) WithArguments(args []ArgumentDefinition) *CommandBuilder {
	cb.command.arguments = args
	return cb
}

// WithAliases sets the command aliases
func (cb *CommandBuilder) WithAliases(aliases []string) *CommandBuilder {
	cb.command.aliases = aliases
	return cb
}

// WithExample sets the command example
func (cb *CommandBuilder) WithExample(example string) *CommandBuilder {
	cb.command.example = example
	return cb
}

// WithMetadata sets command metadata
func (cb *CommandBuilder) WithMetadata(metadata map[string]interface{}) *CommandBuilder {
	if cb.command.metadata == nil {
		cb.command.metadata = make(map[string]interface{})
	}
	for k, v := range metadata {
		cb.command.metadata[k] = v
	}
	return cb
}

// WithMetadataValue sets a single metadata value
func (cb *CommandBuilder) WithMetadataValue(key string, value interface{}) *CommandBuilder {
	if cb.command.metadata == nil {
		cb.command.metadata = make(map[string]interface{})
	}
	cb.command.metadata[key] = value
	return cb
}

// Hidden marks the command as hidden
func (cb *CommandBuilder) Hidden() *CommandBuilder {
	cb.command.hidden = true
	return cb
}

// AddSubCommand adds a sub-command and sets the parent relationship
func (cb *CommandBuilder) AddSubCommand(subCmd Command) *CommandBuilder {
	cb.command.subCommands = append(cb.command.subCommands, subCmd)
	// Set parent reference if the sub-command is a BaseCommand
	if baseCmd, ok := subCmd.(*BaseCommand); ok {
		baseCmd.parent = cb.command
	}
	return cb
}

// Build returns the constructed command
func (cb *CommandBuilder) Build() *BaseCommand {
	return cb.command
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
	
	// RegisterHierarchy registers a command and all its sub-commands recursively
	RegisterHierarchy(cmd Command) error
	
	// GetByPath retrieves a command by its full path (e.g., "git commit")
	GetByPath(path string) (Command, bool)
	
	// GetRootCommands returns all top-level commands (commands without parents)
	GetRootCommands() []Command
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

// RegisterHierarchy registers a command and all its sub-commands recursively
func (r *DefaultCommandRegistry) RegisterHierarchy(cmd Command) error {
	// Register the main command first
	if err := r.Register(cmd); err != nil {
		return err
	}
	
	// Register all sub-commands recursively
	for _, subCmd := range cmd.GetSubCommands() {
		if err := r.RegisterHierarchy(subCmd); err != nil {
			return fmt.Errorf("failed to register sub-command %s: %w", subCmd.ID(), err)
		}
	}
	
	return nil
}

// GetByPath retrieves a command by its full path (e.g., "git commit")
func (r *DefaultCommandRegistry) GetByPath(path string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// For now, search through all commands to find one with matching path
	// This could be optimized with a path-based index in the future
	for _, cmd := range r.commands {
		if cmd.GetPath() == path {
			return cmd, true
		}
	}
	
	return nil, false
}

// GetRootCommands returns all top-level commands (commands without parents)
func (r *DefaultCommandRegistry) GetRootCommands() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	commands := make([]Command, 0)
	for _, cmd := range r.commands {
		if cmd.GetParent() == nil {
			commands = append(commands, cmd)
		}
	}
	
	return commands
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

// RegisterCommandHierarchy registers a command and all its sub-commands in the global registry
func RegisterCommandHierarchy(cmd Command) error {
	return globalRegistry.RegisterHierarchy(cmd)
}

// GetCommandByPath retrieves a command by its full path from the global registry
func GetCommandByPath(path string) (Command, bool) {
	return globalRegistry.GetByPath(path)
}

// GetRootCommands returns all top-level commands from the global registry
func GetRootCommands() []Command {
	return globalRegistry.GetRootCommands()
}