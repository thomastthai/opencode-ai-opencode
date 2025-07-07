package commands

import (
	"fmt"
	"strings"

	"github.com/opencode-ai/opencode/internal/app"
)

// SlashCommand represents a parsed slash command with its components
type SlashCommand struct {
	Raw        string   // Original input (e.g., "/session new my-session")
	Topic      string   // The topic/subject (e.g., "session")
	Verb       string   // The action/verb (e.g., "new")
	Args       []string // Additional arguments (e.g., ["my-session"])
	Incomplete bool     // Whether the command is still being typed
}

// ParseState represents the current parsing context for completions
type ParseState int

const (
	ParseStateTopic ParseState = iota // User is selecting/typing topic
	ParseStateVerb                    // User is selecting/typing verb
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

	// Check if we have a verb
	if len(parts) > 1 {
		result.Verb = parts[1]
		
		// Remaining parts are arguments
		if len(parts) > 2 {
			result.Args = parts[2:]
		}
	}

	// Determine if command is complete based on trailing space
	result.Incomplete = !strings.HasSuffix(input, " ") || len(parts) < 2

	return result
}

// GetParseState determines what phase of parsing we're in
func (p *CommandParser) GetParseState(parsed SlashCommand) ParseState {
	if parsed.Topic == "" || (!strings.HasSuffix(parsed.Raw, " ") && parsed.Verb == "") {
		return ParseStateTopic
	}
	if parsed.Verb == "" || (!strings.HasSuffix(parsed.Raw, " ") && len(parsed.Args) == 0) {
		return ParseStateVerb
	}
	return ParseStateArgs
}

// GetCompletions returns possible completions for the current parse state
func (p *CommandParser) GetCompletions(parsed SlashCommand) []CommandCompletion {
	state := p.GetParseState(parsed)
	
	switch state {
	case ParseStateTopic:
		return p.getTopicCompletions(parsed.Topic)
	case ParseStateVerb:
		return p.getVerbCompletions(parsed.Topic, parsed.Verb)
	case ParseStateArgs:
		return p.getArgCompletions(parsed.Topic, parsed.Verb, parsed.Args)
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

// getVerbCompletions returns completions for verbs based on topic
func (p *CommandParser) getVerbCompletions(topic, partial string) []CommandCompletion {
	topicObj, exists := p.hierarchicalReg.GetTopic(topic)
	if !exists {
		return nil
	}
	
	completions := make([]CommandCompletion, 0, len(topicObj.Verbs))
	
	for _, verb := range topicObj.Verbs {
		// Skip empty verb (used for special cases like /help)
		if verb.ID == "" && topic != "help" {
			continue
		}
		
		// Filter by partial match if provided
		if partial != "" && !strings.HasPrefix(verb.ID, strings.ToLower(partial)) {
			continue
		}
		
		completion := CommandCompletion{
			Value:       verb.ID,
			Display:     verb.Name,
			Description: verb.Description,
			Complete:    fmt.Sprintf("/%s %s ", topic, verb.ID),
		}
		
		// Add args help if available
		if verb.ArgsHelp != "" {
			completion.Description += " " + verb.ArgsHelp
		}
		
		completions = append(completions, completion)
	}
	
	return completions
}

// getArgCompletions returns completions for arguments
func (p *CommandParser) getArgCompletions(topic, verb string, args []string) []CommandCompletion {
	// Use dynamic completions if app context is available
	if p.app != nil {
		return GetDynamicCompletions(topic, verb, args, p.app)
	}
	
	// Return empty to indicate free-form input
	return nil
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
	case ParseStateVerb:
		if parsed.Verb != "" {
			prefix := p.findCommonPrefix(completions, parsed.Verb)
			if prefix != parsed.Verb {
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
	if cmd.Topic == "help" && cmd.Verb == "" {
		return true
	}
	
	// Normal commands need topic and verb
	if cmd.Topic == "" || cmd.Verb == "" {
		return false
	}
	
	// Check if the command exists
	_, exists := p.hierarchicalReg.GetVerb(cmd.Topic, cmd.Verb)
	return exists
}

// GetHierarchicalRegistry returns the hierarchical registry for execution
func (p *CommandParser) GetHierarchicalRegistry() *HierarchicalRegistry {
	return p.hierarchicalReg
}