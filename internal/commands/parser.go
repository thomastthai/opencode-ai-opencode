package commands

import (
	"fmt"
	"strings"

	"github.com/opencode-ai/opencode/internal/app"
)

// SlashCommand represents a parsed slash command with its components
type SlashCommand struct {
	Raw        string         // Original input (e.g., "/session new my-session")
	Topic      string         // The topic/subject (e.g., "session")
	Command    string         // The command (e.g., "new")
	Args       []string       // Additional arguments (e.g., ["my-session"])
	Options    *ParsedOptions // Parsed options/switches
	Incomplete bool           // Whether the command is still being typed
}

// ParseState represents the current parsing context for completions
type ParseState int

const (
	ParseStateTopic ParseState = iota // User is selecting/typing topic
	ParseStateCommand                 // User is selecting/typing command
	ParseStateArgs                    // User is typing arguments
)

// CommandParser handles parsing of slash commands
type CommandParser struct {
	registry         *CommandRegistry
	hierarchicalReg  *HierarchicalRegistry
	app             *app.App
}

// NewCommandParser creates a new command parser
func NewCommandParser(registry *CommandRegistry) *CommandParser {
	// Initialize hierarchical registry
	hierarchicalReg := NewHierarchicalRegistry()
	InitializeBuiltinCommands(hierarchicalReg)
	
	return &CommandParser{
		registry:        registry,
		hierarchicalReg: hierarchicalReg,
	}
}

// NewCommandParserWithApp creates a parser with app context for dynamic completions
func NewCommandParserWithApp(registry *CommandRegistry, app *app.App) *CommandParser {
	parser := NewCommandParser(registry)
	parser.app = app
	return parser
}

// Parse parses a command string into its components
func (p *CommandParser) Parse(input string) SlashCommand {
	result := SlashCommand{
		Raw:        input,
		Options:    NewParsedOptions(),
		Incomplete: true,
	}

	// Remove leading slash and trim
	if !strings.HasPrefix(input, "/") {
		return result
	}
	
	trimmed := strings.TrimPrefix(input, "/")
	trimmed = strings.TrimSpace(trimmed)
	
	if trimmed == "" {
		// Just "/" - ready to show topics
		return result
	}

	// Split by spaces, handling empty strings
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return result
	}

	// First part is the topic
	result.Topic = parts[0]

	// Check if we have a command
	if len(parts) > 1 {
		result.Command = parts[1]
		
		// Remaining parts are arguments and options
		if len(parts) > 2 {
			remainingArgs := parts[2:]
			
			// Parse options if we have a complete command
			if result.Topic != "" && result.Command != "" {
				// Get all applicable options for this command
				allOptions := p.hierarchicalReg.GetAllOptions(result.Topic, result.Command)
				
				if len(allOptions) > 0 {
					// Parse options from the remaining arguments
					optParser := NewOptionParser(allOptions)
					parsedOpts, err := optParser.ParseArgs(remainingArgs)
					if err == nil {
						result.Options = parsedOpts
						result.Args = parsedOpts.GetPositional()
					} else {
						// If parsing fails, still keep track of args for completion
						result.Args = remainingArgs
					}
				} else {
					// No options defined, all are positional args
					result.Args = remainingArgs
				}
			} else {
				// Command not complete yet, treat all as args
				result.Args = remainingArgs
			}
		}
	}

	// Determine if command is complete based on trailing space
	result.Incomplete = !strings.HasSuffix(input, " ") || len(parts) < 2

	return result
}

// GetParseState determines what phase of parsing we're in
func (p *CommandParser) GetParseState(parsed SlashCommand) ParseState {
	// If no topic or typing topic
	if parsed.Topic == "" || (!strings.HasSuffix(parsed.Raw, " ") && parsed.Command == "") {
		return ParseStateTopic
	}
	
	// If we have a topic but no command, or typing command
	if parsed.Command == "" {
		return ParseStateCommand
	}
	
	// If we have both topic and command, we're in args state
	// This includes option parsing
	return ParseStateArgs
}

// GetCompletions returns possible completions for the current parse state
func (p *CommandParser) GetCompletions(parsed SlashCommand) []CommandCompletion {
	state := p.GetParseState(parsed)
	
	switch state {
	case ParseStateTopic:
		return p.getTopicCompletions(parsed.Topic)
	case ParseStateCommand:
		return p.getCommandCompletions(parsed.Topic, parsed.Command)
	case ParseStateArgs:
		// Check if the command actually exists
		if _, exists := p.hierarchicalReg.GetCommand(parsed.Topic, parsed.Command); !exists {
			// Command doesn't exist, so this might be a partial command
			// If not ending with space, treat as command completion
			if !strings.HasSuffix(parsed.Raw, " ") {
				return p.getCommandCompletions(parsed.Topic, parsed.Command)
			}
		}
		return p.getArgCompletions(parsed.Topic, parsed.Command, parsed.Args)
	}
	
	return nil
}

// CommandCompletion represents a possible completion
type CommandCompletion struct {
	Value       string // The completion value (e.g., "session")
	Display     string // Display text (e.g., "session - Manage sessions")
	Description string // Longer description
	Icon        string // Icon to display
	Complete    string // The full text to insert when selected
}

// getTopicCompletions returns completions for topics
func (p *CommandParser) getTopicCompletions(partial string) []CommandCompletion {
	allTopics := p.hierarchicalReg.ListTopics()
	completions := make([]CommandCompletion, 0, len(allTopics))
	
	for _, topic := range allTopics {
		// Filter by partial match if provided
		if partial != "" && !strings.HasPrefix(topic.ID, strings.ToLower(partial)) {
			continue
		}
		
		completion := CommandCompletion{
			Value:       topic.ID,
			Display:     topic.Name,
			Description: topic.Description,
			Icon:        topic.Icon,
			Complete:    "/" + topic.ID + " ",
		}
		
		// Special case for help - it can be executed without a verb
		if topic.ID == "help" {
			completion.Complete = "/help "
		}
		
		completions = append(completions, completion)
	}
	
	return completions
}

// getCommandCompletions returns completions for commands based on topic
func (p *CommandParser) getCommandCompletions(topic, partial string) []CommandCompletion {
	topicObj, exists := p.hierarchicalReg.GetTopic(topic)
	if !exists {
		return nil
	}
	
	completions := make([]CommandCompletion, 0, len(topicObj.Commands))
	
	for _, command := range topicObj.Commands {
		// Skip empty command (used for special cases like /help)
		if command.ID == "" && topic != "help" {
			continue
		}
		
		// Filter by partial match if provided
		if partial != "" && !strings.HasPrefix(command.ID, strings.ToLower(partial)) {
			continue
		}
		
		completion := CommandCompletion{
			Value:       command.ID,
			Display:     command.Name,
			Description: command.Description,
			Complete:    fmt.Sprintf("/%s %s ", topic, command.ID),
		}
		
		// Add args help if available
		if command.ArgsHelp != "" {
			completion.Description += " " + command.ArgsHelp
		}
		
		completions = append(completions, completion)
	}
	
	return completions
}

// getArgCompletions returns completions for arguments
func (p *CommandParser) getArgCompletions(topic, command string, args []string) []CommandCompletion {
	completions := []CommandCompletion{}
	
	// Build the command prefix for option completions
	commandPrefix := fmt.Sprintf("/%s %s", topic, command)
	
	// Check if the last argument looks like an option prefix
	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if strings.HasPrefix(lastArg, "-") {
			// Get option completions
			options := p.hierarchicalReg.GetAllOptions(topic, command)
			return p.getOptionCompletions(lastArg, options, commandPrefix, args[:len(args)-1])
		}
	}
	
	// Use dynamic completions if app context is available
	if p.app != nil {
		dynamicCompletions := GetDynamicCompletions(topic, command, args, p.app)
		completions = append(completions, dynamicCompletions...)
	}
	
	// Also add option completions with -- prefix
	options := p.hierarchicalReg.GetAllOptions(topic, command)
	optionCompletions := p.getAllOptionCompletions(options, commandPrefix, args)
	completions = append(completions, optionCompletions...)
	
	return completions
}

// getOptionCompletions returns completions for options based on prefix
func (p *CommandParser) getOptionCompletions(prefix string, options []*Option, commandPrefix string, existingArgs []string) []CommandCompletion {
	completions := []CommandCompletion{}
	
	// Determine if it's long or short option
	isLong := strings.HasPrefix(prefix, "--")
	searchPrefix := strings.TrimLeft(prefix, "-")
	
	// Build a map to avoid duplicates
	seen := make(map[string]bool)
	
	// Build the base command with existing args
	baseCommand := commandPrefix
	if len(existingArgs) > 0 {
		baseCommand += " " + strings.Join(existingArgs, " ")
	}
	
	for _, opt := range options {
		if opt.Hidden {
			continue
		}
		
		// Check long option
		if (isLong || prefix == "-") && strings.HasPrefix(opt.Name, searchPrefix) {
			if !seen[opt.Name] {
				optionText := "--" + opt.Name
				completions = append(completions, CommandCompletion{
					Value:       optionText,
					Display:     FormatOptionName(opt),
					Description: opt.Description,
					Complete:    baseCommand + " " + optionText + " ",
				})
				seen[opt.Name] = true
			}
		}
		
		// Check short option
		if !isLong && opt.ShortName != "" && strings.HasPrefix(opt.ShortName, searchPrefix) {
			if !seen[opt.ShortName] {
				optionText := "-" + opt.ShortName
				completions = append(completions, CommandCompletion{
					Value:       optionText,
					Display:     FormatOptionName(opt),
					Description: opt.Description,
					Complete:    baseCommand + " " + optionText + " ",
				})
				seen[opt.ShortName] = true
			}
		}
	}
	
	return completions
}

// getAllOptionCompletions returns all available option completions
func (p *CommandParser) getAllOptionCompletions(options []*Option, commandPrefix string, existingArgs []string) []CommandCompletion {
	completions := []CommandCompletion{}
	
	// Build a map to avoid duplicates
	seen := make(map[string]bool)
	
	// Build the base command with existing args
	baseCommand := commandPrefix
	if len(existingArgs) > 0 {
		baseCommand += " " + strings.Join(existingArgs, " ")
	}
	
	for _, opt := range options {
		if opt.Hidden || seen[opt.Name] {
			continue
		}
		
		optionText := "--" + opt.Name
		completions = append(completions, CommandCompletion{
			Value:       optionText,
			Display:     FormatOptionName(opt),
			Description: opt.Description,
			Complete:    baseCommand + " " + optionText + " ",
		})
		seen[opt.Name] = true
	}
	
	return completions
}

// GetTabCompletion returns the best completion for tab key
func (p *CommandParser) GetTabCompletion(input string) (string, []CommandCompletion) {
	parsed := p.Parse(input)
	completions := p.GetCompletions(parsed)
	
	if len(completions) == 0 {
		return input, nil
	}
	
	if len(completions) == 1 {
		// Single match - complete it
		return completions[0].Complete, nil
	}
	
	// Multiple matches - try to find common prefix
	state := p.GetParseState(parsed)
	switch state {
	case ParseStateTopic:
		if parsed.Topic != "" {
			// Find longest common prefix
			prefix := p.findCommonPrefix(completions, parsed.Topic)
			if prefix != parsed.Topic {
				return "/" + prefix, completions
			}
		}
	case ParseStateCommand:
		if parsed.Command != "" {
			prefix := p.findCommonPrefix(completions, parsed.Command)
			if prefix != parsed.Command {
				return "/" + parsed.Topic + " " + prefix, completions
			}
		}
	}
	
	// Return original input with available completions
	return input, completions
}

// findCommonPrefix finds the longest common prefix among completions
func (p *CommandParser) findCommonPrefix(completions []CommandCompletion, current string) string {
	if len(completions) == 0 {
		return current
	}
	
	prefix := completions[0].Value
	for _, comp := range completions[1:] {
		prefix = commonPrefix(prefix, comp.Value)
		if prefix == "" {
			return current
		}
	}
	
	// Only return if it's longer than current
	if len(prefix) > len(current) {
		return prefix
	}
	return current
}

// commonPrefix returns the common prefix of two strings
func commonPrefix(a, b string) string {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	
	return a[:minLen]
}

// CanExecute checks if a slash command is complete and can be executed
func (p *CommandParser) CanExecute(cmd SlashCommand) bool {
	// Special case for help
	if cmd.Topic == "help" && cmd.Command == "" {
		return true
	}
	
	// Normal commands need topic and command
	if cmd.Topic == "" || cmd.Command == "" {
		return false
	}
	
	// Check if the command exists
	_, exists := p.hierarchicalReg.GetCommand(cmd.Topic, cmd.Command)
	return exists
}

// GetHierarchicalRegistry returns the hierarchical registry for execution
func (p *CommandParser) GetHierarchicalRegistry() *HierarchicalRegistry {
	return p.hierarchicalReg
}