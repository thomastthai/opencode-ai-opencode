package commands

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// CommandType defines the origin of a command.
type CommandType int

const (
	BuiltinCommand CommandType = iota
	UserCommand
	ProjectCommand
	PluginCommand
)

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

// ArgumentDefinition defines a command-line argument.
type ArgumentDefinition struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     interface{}
}

// CommandHandler is the function that executes a command's logic.
type CommandHandler func(ctx context.Context, args map[string]interface{}) error

// Command is the interface that all commands must implement.
type Command interface {
	ID() string
	Name() string
	Description() string
	Category() string
	Type() CommandType
	Example() string
	GetAliases() []string
	GetArguments() []ArgumentDefinition
	GetSubCommands() []Command
	GetParent() Command
	GetPath() string
	GetMetadata() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) error
	ValidateArgs(args map[string]interface{}) error
	AddSubCommand(sub Command)
}

// Registry is the interface for a command registry.
type Registry interface {
	Register(cmd Command) error
	Unregister(id string) error
	Get(id string) (Command, bool)
	List() []Command
	RegisterHierarchy(cmd Command) error
}

// CommandRegistry manages all available commands.
type CommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]Command
	aliases  map[string]string
}

// NewCommandRegistry creates a new, empty command registry.
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

// Register adds a new command to the registry.
func (r *CommandRegistry) Register(cmd Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[cmd.ID()]; exists {
		return fmt.Errorf("command with ID '%s' already exists", cmd.ID())
	}

	r.commands[cmd.ID()] = cmd
	return nil
}

// Unregister removes a command from the registry.
func (r *CommandRegistry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[id]; !exists {
		return fmt.Errorf("command with ID '%s' not found", id)
	}

	delete(r.commands, id)
	return nil
}

// Get retrieves a command by its ID.
func (r *CommandRegistry) Get(id string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, exists := r.commands[id]
	return cmd, exists
}

// List returns all registered commands.
func (r *CommandRegistry) List() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmds := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// RegisterHierarchy registers a command and all its sub-commands recursively.
// If any sub-command fails to register, any commands in the hierarchy that were
// already registered will be unregistered.
func (r *CommandRegistry) RegisterHierarchy(cmd Command) error {
	if err := r.Register(cmd); err != nil {
		return err
	}

	for _, sub := range cmd.GetSubCommands() {
		if err := r.RegisterHierarchy(sub); err != nil {
			// If a sub-command fails, unregister the parent to keep the registry clean.
			r.Unregister(cmd.ID())
			return err
		}
	}

	return nil
}

// globalRegistry is the singleton instance of the command registry.
var globalRegistry *CommandRegistry
var once sync.Once

// GetGlobalRegistry returns the global command registry.
func GetGlobalRegistry() *CommandRegistry {
	once.Do(func() {
		globalRegistry = NewCommandRegistry()
	})
	return globalRegistry
}

// RegisterBuiltIn registers a command as a built-in command.
// Built-in commands are core to the application and cannot be overridden.
// This function is for registering single commands. For commands with sub-commands,
// use RegisterBuiltInHierarchy.
func RegisterBuiltIn(cmd Command) {
	registry := GetGlobalRegistry()
	if err := registry.Register(cmd); err != nil {
		log.Fatalf("Failed to register built-in command '%s': %v", cmd.ID(), err)
	}
}

// RegisterBuiltInHierarchy registers a command and its entire sub-command tree.
func RegisterBuiltInHierarchy(cmd Command) {
	registry := GetGlobalRegistry()
	if err := registry.RegisterHierarchy(cmd); err != nil {
		log.Fatalf("Failed to register built-in command hierarchy '%s': %v", cmd.ID(), err)
	}
}
