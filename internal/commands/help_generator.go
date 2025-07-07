package commands

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateCommandHelp generates help text for a specific command
func GenerateCommandHelp(registry *HierarchicalRegistry, topicID, commandID string) string {
	topic, exists := registry.GetTopic(topicID)
	if !exists {
		return fmt.Sprintf("Topic '%s' not found", topicID)
	}
	
	command, exists := registry.GetCommand(topicID, commandID)
	if !exists {
		return fmt.Sprintf("Command '%s %s' not found", topicID, commandID)
	}
	
	var help strings.Builder
	
	// Command header
	help.WriteString(fmt.Sprintf("# %s %s\n\n", topic.Name, command.Name))
	help.WriteString(fmt.Sprintf("%s\n\n", command.Description))
	
	// Usage
	help.WriteString("## Usage\n")
	help.WriteString(fmt.Sprintf("  /%s %s", topicID, commandID))
	if command.ArgsHelp != "" {
		help.WriteString(" " + command.ArgsHelp)
	}
	help.WriteString(" [options]\n\n")
	
	// Options
	options := registry.GetAllOptions(topicID, commandID)
	if len(options) > 0 {
		help.WriteString("## Options\n")
		help.WriteString(generateOptionsHelp(options, false))
		help.WriteString("\n")
	}
	
	// Examples if available
	examples := getCommandExamples(topicID, commandID, options)
	if len(examples) > 0 {
		help.WriteString("## Examples\n")
		for _, example := range examples {
			help.WriteString(fmt.Sprintf("  %s\n", example))
		}
		help.WriteString("\n")
	}
	
	return help.String()
}

// GenerateTopicHelp generates help text for a topic
func GenerateTopicHelp(registry *HierarchicalRegistry, topicID string) string {
	topic, exists := registry.GetTopic(topicID)
	if !exists {
		return fmt.Sprintf("Topic '%s' not found", topicID)
	}
	
	var help strings.Builder
	
	// Topic header
	help.WriteString(fmt.Sprintf("# %s %s\n\n", topic.Icon, topic.Name))
	help.WriteString(fmt.Sprintf("%s\n\n", topic.Description))
	
	// Commands
	help.WriteString("## Commands\n")
	commands := make([]*HierCommand, 0, len(topic.Commands))
	for _, cmd := range topic.Commands {
		if cmd.ID != "" { // Skip empty command
			commands = append(commands, cmd)
		}
	}
	
	// Sort commands by name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
	
	// Display commands
	for _, cmd := range commands {
		help.WriteString(fmt.Sprintf("  %-12s %s", cmd.ID, cmd.Description))
		if cmd.ArgsHelp != "" {
			help.WriteString(fmt.Sprintf(" %s", cmd.ArgsHelp))
		}
		help.WriteString("\n")
	}
	
	// Global options
	if len(topic.Options) > 0 {
		help.WriteString("\n## Global Options\n")
		help.WriteString("These options are available for all commands in this topic:\n")
		help.WriteString(generateOptionsHelp(topic.Options, true))
	}
	
	return help.String()
}

// GenerateAllHelp generates help text showing all available commands
func GenerateAllHelp(registry *HierarchicalRegistry) string {
	var help strings.Builder
	
	help.WriteString("# Available Commands\n\n")
	
	topics := registry.ListTopics()
	
	// Sort topics by name
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Name < topics[j].Name
	})
	
	for _, topic := range topics {
		if topic.ID == "help" {
			continue // Skip help topic in general listing
		}
		
		help.WriteString(fmt.Sprintf("## %s %s\n", topic.Icon, topic.Name))
		help.WriteString(fmt.Sprintf("%s\n\n", topic.Description))
		
		// List commands
		commands := make([]*HierCommand, 0, len(topic.Commands))
		for _, cmd := range topic.Commands {
			if cmd.ID != "" {
				commands = append(commands, cmd)
			}
		}
		
		// Sort commands
		sort.Slice(commands, func(i, j int) bool {
			return commands[i].Name < commands[j].Name
		})
		
		for _, cmd := range commands {
			help.WriteString(fmt.Sprintf("  /%s %-10s %s\n", topic.ID, cmd.ID, cmd.Description))
		}
		help.WriteString("\n")
	}
	
	help.WriteString("Use `/help <topic>` or `/help <topic> <command>` for more details.\n")
	
	return help.String()
}

// generateOptionsHelp generates formatted help text for options
func generateOptionsHelp(options []*Option, indent bool) string {
	if len(options) == 0 {
		return ""
	}
	
	// Group options by category (required vs optional)
	required := []*Option{}
	optional := []*Option{}
	
	for _, opt := range options {
		if opt.Hidden {
			continue
		}
		if opt.Required {
			required = append(required, opt)
		} else {
			optional = append(optional, opt)
		}
	}
	
	var help strings.Builder
	prefix := ""
	if indent {
		prefix = "  "
	}
	
	// Required options first
	if len(required) > 0 {
		help.WriteString(fmt.Sprintf("%sRequired:\n", prefix))
		for _, opt := range required {
			help.WriteString(formatOptionHelp(opt, prefix+"  "))
		}
	}
	
	// Optional options
	if len(optional) > 0 {
		if len(required) > 0 {
			help.WriteString("\n")
		}
		help.WriteString(fmt.Sprintf("%sOptional:\n", prefix))
		for _, opt := range optional {
			help.WriteString(formatOptionHelp(opt, prefix+"  "))
		}
	}
	
	return help.String()
}

// formatOptionHelp formats a single option for help display
func formatOptionHelp(opt *Option, prefix string) string {
	var help strings.Builder
	
	// Option name(s)
	help.WriteString(prefix)
	help.WriteString(fmt.Sprintf("%-20s", FormatOptionName(opt)))
	
	// Description
	help.WriteString(opt.Description)
	
	// Additional details
	details := []string{}
	
	// Default value
	if opt.DefaultValue != nil {
		details = append(details, fmt.Sprintf("default: %v", opt.DefaultValue))
	}
	
	// Choices
	if len(opt.Choices) > 0 {
		details = append(details, fmt.Sprintf("choices: %s", strings.Join(opt.Choices, ", ")))
	}
	
	// Range for numeric options
	if opt.Type == OptionTypeInt || opt.Type == OptionTypeFloat {
		if opt.MinValue != 0 || opt.MaxValue != 0 {
			details = append(details, fmt.Sprintf("range: [%.0f, %.0f]", opt.MinValue, opt.MaxValue))
		}
	}
	
	// Repeatable
	if opt.Repeatable {
		details = append(details, "repeatable")
	}
	
	if len(details) > 0 {
		help.WriteString(fmt.Sprintf(" (%s)", strings.Join(details, ", ")))
	}
	
	help.WriteString("\n")
	
	// Example if provided
	if opt.Example != "" {
		help.WriteString(fmt.Sprintf("%s  Example: %s\n", prefix, opt.Example))
	}
	
	return help.String()
}

// getCommandExamples returns example usages for a command
func getCommandExamples(topicID, commandID string, options []*Option) []string {
	examples := []string{}
	
	// Add basic examples based on known commands
	switch topicID {
	case "session":
		switch commandID {
		case "new":
			examples = append(examples, "/session new")
			examples = append(examples, "/session new my-project")
			if hasOption(options, "model") {
				examples = append(examples, "/session new --model=gpt-4 my-project")
			}
		case "compact":
			examples = append(examples, "/session compact")
			examples = append(examples, "/session compact focus on errors")
			if hasOption(options, "instructions") {
				examples = append(examples, "/session compact --instructions=\"summarize key points\"")
			}
		}
	case "config":
		switch commandID {
		case "set":
			examples = append(examples, "/config set editor.theme dark")
			if hasOption(options, "global") {
				examples = append(examples, "/config set --global api.key YOUR_KEY")
			}
		}
	case "auth":
		switch commandID {
		case "login":
			examples = append(examples, "/auth login openai")
			examples = append(examples, "/auth login anthropic")
			if hasOption(options, "force") {
				examples = append(examples, "/auth login --force gemini")
			}
		}
	}
	
	return examples
}

// hasOption checks if an option with the given name exists
func hasOption(options []*Option, name string) bool {
	for _, opt := range options {
		if opt.Name == name {
			return true
		}
	}
	return false
}