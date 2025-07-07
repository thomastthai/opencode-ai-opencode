package commands

import (
	"context"
	"fmt"
	"strings"
)

// Topic represents a command topic (e.g., session, config, auth)
type Topic struct {
	ID          string
	Name        string
	Description string
	Icon        string
	Commands    map[string]*HierCommand
	Options     []*Option // Global options available for all commands in this topic
}

// HierCommand represents an action within a topic (e.g., new, list, delete)
type HierCommand struct {
	ID          string
	Name        string
	Description string
	Handler     CommandHandler
	ArgsHelp    string    // Help text for arguments
	MinArgs     int
	MaxArgs     int
	Options     []*Option // Command-specific options
}

// HierarchicalRegistry manages commands in a topic/command structure
type HierarchicalRegistry struct {
	topics map[string]*Topic
}

// NewHierarchicalRegistry creates a new hierarchical command registry
func NewHierarchicalRegistry() *HierarchicalRegistry {
	return &HierarchicalRegistry{
		topics: make(map[string]*Topic),
	}
}

// RegisterTopic registers a new topic
func (r *HierarchicalRegistry) RegisterTopic(topic *Topic) error {
	if _, exists := r.topics[topic.ID]; exists {
		return fmt.Errorf("topic '%s' already registered", topic.ID)
	}
	if topic.Commands == nil {
		topic.Commands = make(map[string]*HierCommand)
	}
	r.topics[topic.ID] = topic
	return nil
}

// RegisterCommand registers a command under a topic
func (r *HierarchicalRegistry) RegisterCommand(topicID string, command *HierCommand) error {
	topic, exists := r.topics[topicID]
	if !exists {
		return fmt.Errorf("topic '%s' not found", topicID)
	}
	if _, exists := topic.Commands[command.ID]; exists {
		return fmt.Errorf("command '%s' already registered in topic '%s'", command.ID, topicID)
	}
	topic.Commands[command.ID] = command
	return nil
}

// GetTopic retrieves a topic by ID
func (r *HierarchicalRegistry) GetTopic(id string) (*Topic, bool) {
	topic, exists := r.topics[id]
	return topic, exists
}

// GetCommand retrieves a command from a topic
func (r *HierarchicalRegistry) GetCommand(topicID, commandID string) (*HierCommand, bool) {
	topic, exists := r.topics[topicID]
	if !exists {
		return nil, false
	}
	command, exists := topic.Commands[commandID]
	return command, exists
}

// ListTopics returns all registered topics
func (r *HierarchicalRegistry) ListTopics() []*Topic {
	topics := make([]*Topic, 0, len(r.topics))
	for _, topic := range r.topics {
		topics = append(topics, topic)
	}
	return topics
}

// Execute runs a command based on parsed input
func (r *HierarchicalRegistry) Execute(ctx context.Context, cmd SlashCommand) error {
	command, exists := r.GetCommand(cmd.Topic, cmd.Command)
	if !exists {
		return fmt.Errorf("command not found: /%s %s", cmd.Topic, cmd.Command)
	}

	// Validate args
	if len(cmd.Args) < command.MinArgs {
		return fmt.Errorf("insufficient arguments: expected at least %d, got %d", command.MinArgs, len(cmd.Args))
	}
	if command.MaxArgs >= 0 && len(cmd.Args) > command.MaxArgs {
		return fmt.Errorf("too many arguments: expected at most %d, got %d", command.MaxArgs, len(cmd.Args))
	}

	// Validate options
	allOptions := r.GetAllOptions(cmd.Topic, cmd.Command)
	if err := cmd.Options.Validate(allOptions); err != nil {
		return fmt.Errorf("option validation failed: %w", err)
	}

	// Convert args to map for handler
	args := map[string]interface{}{
		"args":    cmd.Args,
		"raw":     cmd.Raw,
		"options": cmd.Options,
	}

	if command.Handler == nil {
		return fmt.Errorf("no handler for command: /%s %s", cmd.Topic, cmd.Command)
	}

	return command.Handler(ctx, args)
}

// GetCompletionsForTopic returns command completions for a topic
func (r *HierarchicalRegistry) GetCompletionsForTopic(topicID string) []CommandCompletion {
	topic, exists := r.topics[topicID]
	if !exists {
		return nil
	}

	completions := make([]CommandCompletion, 0, len(topic.Commands))
	for _, command := range topic.Commands {
		completions = append(completions, CommandCompletion{
			Value:       command.ID,
			Display:     command.Name,
			Description: command.Description,
			Complete:    fmt.Sprintf("/%s %s ", topicID, command.ID),
		})
	}
	return completions
}

// GetAllOptions returns all applicable options for a command (topic + command options)
func (r *HierarchicalRegistry) GetAllOptions(topicID, commandID string) []*Option {
	options := []*Option{}
	
	// Add topic-level options
	if topic, exists := r.topics[topicID]; exists {
		options = append(options, topic.Options...)
	}
	
	// Add command-level options
	if command, exists := r.GetCommand(topicID, commandID); exists {
		options = append(options, command.Options...)
	}
	
	return options
}

// InitializeBuiltinCommands sets up all built-in commands with the new structure
func InitializeBuiltinCommands(registry *HierarchicalRegistry) error {
	// Session commands
	sessionTopic := &Topic{
		ID:          "session",
		Name:        "Session",
		Description: "Manage chat sessions",
		Icon:        "💬",
		Options: []*Option{
			{
				Name:        "verbose",
				ShortName:   "v",
				Type:        OptionTypeBool,
				Description: "Show detailed output",
			},
		},
	}
	registry.RegisterTopic(sessionTopic)

	registry.RegisterCommand("session", &HierCommand{
		ID:          "new",
		Name:        "New",
		Description: "Create a new session",
		Handler:     handleHierSessionNew,
		ArgsHelp:    "[name]",
		MinArgs:     0,
		MaxArgs:     1,
		Options: []*Option{
			{
				Name:        "model",
				ShortName:   "m",
				Type:        OptionTypeString,
				Description: "Specify the AI model to use",
				Example:     "--model=gpt-4",
			},
			{
				Name:        "system",
				Type:        OptionTypeString,
				Description: "Set system prompt for the session",
			},
		},
	})

	registry.RegisterCommand("session", &HierCommand{
		ID:          "list",
		Name:        "List",
		Description: "List all sessions",
		Handler:     handleHierSessionList,
		MinArgs:     0,
		MaxArgs:     0,
		Options: []*Option{
			{
				Name:        "format",
				ShortName:   "f",
				Type:        OptionTypeString,
				Description: "Output format",
				Choices:     []string{"table", "json", "csv"},
				DefaultValue: "table",
			},
			{
				Name:        "all",
				ShortName:   "a",
				Type:        OptionTypeBool,
				Description: "Show all sessions including archived",
			},
		},
	})

	registry.RegisterCommand("session", &HierCommand{
		ID:          "clear",
		Name:        "Clear",
		Description: "Clear current session",
		Handler:     handleHierSessionClear,
		MinArgs:     0,
		MaxArgs:     0,
	})

	registry.RegisterCommand("session", &HierCommand{
		ID:          "compact",
		Name:        "Compact",
		Description: "Compact current session",
		Handler:     handleHierSessionCompact,
		ArgsHelp:    "[instructions]",
		MinArgs:     0,
		MaxArgs:     -1, // Unlimited args for instructions
	})

	// Config commands
	configTopic := &Topic{
		ID:          "config",
		Name:        "Configuration",
		Description: "Configuration and settings",
		Icon:        "⚙️",
	}
	registry.RegisterTopic(configTopic)

	registry.RegisterCommand("config", &HierCommand{
		ID:          "show",
		Name:        "Show",
		Description: "View configuration",
		Handler:     handleHierConfigShow,
		MinArgs:     0,
		MaxArgs:     0,
	})

	registry.RegisterCommand("config", &HierCommand{
		ID:          "model",
		Name:        "Model",
		Description: "Select AI model",
		Handler:     handleHierConfigModel,
		ArgsHelp:    "[model-name]",
		MinArgs:     0,
		MaxArgs:     1,
		Options: []*Option{
			{
				Name:        "global",
				ShortName:   "g",
				Type:        OptionTypeBool,
				Description: "Set as global default model",
			},
			{
				Name:        "temperature",
				ShortName:   "t",
				Type:        OptionTypeFloat,
				Description: "Set model temperature",
				MinValue:    0.0,
				MaxValue:    2.0,
				DefaultValue: 0.7,
			},
		},
	})

	// Project commands
	projectTopic := &Topic{
		ID:          "project",
		Name:        "Project",
		Description: "Project management",
		Icon:        "📁",
	}
	registry.RegisterTopic(projectTopic)

	registry.RegisterCommand("project", &HierCommand{
		ID:          "init",
		Name:        "Initialize",
		Description: "Initialize project with CLAUDE.md",
		Handler:     handleHierProjectInit,
		MinArgs:     0,
		MaxArgs:     0,
	})

	// Auth commands
	authTopic := &Topic{
		ID:          "auth",
		Name:        "Authentication",
		Description: "Authentication and login",
		Icon:        "🔐",
	}
	registry.RegisterTopic(authTopic)

	registry.RegisterCommand("auth", &HierCommand{
		ID:          "login",
		Name:        "Login",
		Description: "Login to provider",
		Handler:     handleHierAuthLogin,
		ArgsHelp:    "<provider>",
		MinArgs:     1,
		MaxArgs:     1,
		Options: []*Option{
			{
				Name:        "force",
				ShortName:   "f",
				Type:        OptionTypeBool,
				Description: "Force re-authentication even if already logged in",
			},
			{
				Name:        "no-browser",
				Type:        OptionTypeBool,
				Description: "Don't open browser automatically",
			},
		},
	})

	registry.RegisterCommand("auth", &HierCommand{
		ID:          "logout",
		Name:        "Logout",
		Description: "Logout from provider",
		Handler:     handleHierAuthLogout,
		ArgsHelp:    "[provider]",
		MinArgs:     0,
		MaxArgs:     1,
	})

	registry.RegisterCommand("auth", &HierCommand{
		ID:          "status",
		Name:        "Status",
		Description: "Show authentication status",
		Handler:     handleHierAuthStatus,
		MinArgs:     0,
		MaxArgs:     0,
	})

	// System commands
	systemTopic := &Topic{
		ID:          "system",
		Name:        "System",
		Description: "System commands",
		Icon:        "🖥️",
	}
	registry.RegisterTopic(systemTopic)

	registry.RegisterCommand("system", &HierCommand{
		ID:          "help",
		Name:        "Help",
		Description: "Show help information",
		Handler:     handleHierSystemHelp,
		MinArgs:     0,
		MaxArgs:     1,
	})

	registry.RegisterCommand("system", &HierCommand{
		ID:          "exit",
		Name:        "Exit",
		Description: "Exit application",
		Handler:     handleHierSystemExit,
		MinArgs:     0,
		MaxArgs:     0,
	})

	// Help as a special top-level command
	helpTopic := &Topic{
		ID:          "help",
		Name:        "Help",
		Description: "Show help information",
		Icon:        "❓",
		Commands: map[string]*HierCommand{
			"": { // Empty command for just "/help"
				ID:      "",
				Name:    "General Help",
				Handler: handleHierHelp,
				MinArgs: 0,
				MaxArgs: 0,
			},
		},
	}
	registry.RegisterTopic(helpTopic)

	return nil
}

// Command handlers (these will send appropriate messages to the TUI)
func handleHierSessionNew(ctx context.Context, args map[string]interface{}) error {
	// This will be implemented to send the appropriate message
	return fmt.Errorf("session_new_requested")
}

func handleHierSessionList(ctx context.Context, args map[string]interface{}) error {
	return fmt.Errorf("session_list_requested")
}

func handleHierSessionClear(ctx context.Context, args map[string]interface{}) error {
	return fmt.Errorf("clear_session_requested")
}

func handleHierSessionCompact(ctx context.Context, args map[string]interface{}) error {
	cmdArgs := args["args"].([]string)
	instructions := strings.Join(cmdArgs, " ")
	// Store instructions in args for later use
	args["instructions"] = instructions
	return fmt.Errorf("compact_session_requested")
}

func handleHierConfigShow(ctx context.Context, args map[string]interface{}) error {
	return fmt.Errorf("config_show_requested")
}

func handleHierConfigModel(ctx context.Context, args map[string]interface{}) error {
	return fmt.Errorf("config_model_requested")
}

func handleHierProjectInit(ctx context.Context, args map[string]interface{}) error {
	return fmt.Errorf("init_project_requested")
}

func handleHierAuthLogin(ctx context.Context, args map[string]interface{}) error {
	cmdArgs := args["args"].([]string)
	if len(cmdArgs) > 0 {
		args["provider"] = cmdArgs[0]
	}
	return fmt.Errorf("auth_login_requested")
}

func handleHierAuthLogout(ctx context.Context, args map[string]interface{}) error {
	cmdArgs := args["args"].([]string)
	if len(cmdArgs) > 0 {
		args["provider"] = cmdArgs[0]
	}
	return fmt.Errorf("auth_logout_requested")
}

func handleHierAuthStatus(ctx context.Context, args map[string]interface{}) error {
	return fmt.Errorf("auth_status_requested")
}

func handleHierSystemHelp(ctx context.Context, args map[string]interface{}) error {
	return fmt.Errorf("help_requested")
}

func handleHierSystemExit(ctx context.Context, args map[string]interface{}) error {
	return fmt.Errorf("exit_requested")
}

func handleHierHelp(ctx context.Context, args map[string]interface{}) error {
	return fmt.Errorf("help_requested")
}