package completions

import (
	"fmt"
	"strings"

	"github.com/opencode-ai/opencode/internal/app"
	"github.com/opencode-ai/opencode/internal/commands"
	"github.com/opencode-ai/opencode/internal/logging"
	"github.com/opencode-ai/opencode/internal/tui/components/dialog"
)

// slashCommandProvider provides completions for slash commands
type slashCommandProvider struct {
	parser       *commands.CommandParser
	registry     *commands.CommandRegistry
	currentQuery string // Track current query for progressive completion
}

// NewSlashCommandProvider creates a new slash command completion provider
func NewSlashCommandProvider() dialog.CompletionProvider {
	registry := commands.GetGlobalRegistry()
	return &slashCommandProvider{
		parser:   commands.NewCommandParser(registry),
		registry: registry,
	}
}

// NewSlashCommandProviderWithApp creates a new slash command completion provider with app context
func NewSlashCommandProviderWithApp(app *app.App) dialog.CompletionProvider {
	registry := commands.GetGlobalRegistry()
	return &slashCommandProvider{
		parser:   commands.NewCommandParserWithApp(registry, app),
		registry: registry,
	}
}

// GetId returns the provider ID
func (p *slashCommandProvider) GetId() string {
	return "slash-commands"
}

// GetEntry returns the top-level entry for this provider
func (p *slashCommandProvider) GetEntry() dialog.CompletionItemI {
	return &dialog.CompletionItem{
		Title: "Commands",
		Value: "/",
	}
}

// GetChildEntries returns completion items based on the query
func (p *slashCommandProvider) GetChildEntries(query string) ([]dialog.CompletionItemI, error) {
	// Store the query for later use
	p.currentQuery = query
	
	// The query might come with or without the leading slash
	// In the UI, it comes without (stripped by complete.go)
	// In tests, it comes with the slash
	fullQuery := query
	if !strings.HasPrefix(query, "/") {
		fullQuery = "/" + query
	}
	logging.Debug("[slashCommandProvider.GetChildEntries]", "inputQuery", query, "fullQuery", fullQuery)
	
	
	// Parse the current input
	parsed := p.parser.Parse(fullQuery)
	completions := p.parser.GetCompletions(parsed)
	
	items := make([]dialog.CompletionItemI, len(completions))
	for i, comp := range completions {
		// Create a custom completion item that stores the complete text
		items[i] = &SlashCommandItem{
			CompletionItem: dialog.CompletionItem{
				Title: fmt.Sprintf("%s %s", comp.Icon, comp.Display),
				Value: comp.Value, // Use the actual value, not the complete text
			},
			Description: comp.Description,
			Complete:    comp.Complete,
		}
	}
	
	return items, nil
}

// SlashCommandItem extends CompletionItem with additional metadata
type SlashCommandItem struct {
	dialog.CompletionItem
	Description string
	Complete    string
}

// GetCompleteValue returns the complete command text for this item
func (s *SlashCommandItem) GetCompleteValue() string {
	return s.Complete
}

// Render customizes the display of slash command items
func (s *SlashCommandItem) Render(selected bool, width int) string {
	// Use the parent's rendering for now
	// This can be customized later to show descriptions
	return s.CompletionItem.Render(selected, width)
}

// HandleTabCompletion implements TabHandler interface for tab completion
func (p *slashCommandProvider) HandleTabCompletion(input string) (string, []dialog.CompletionItemI, error) {
	logging.Debug("[slashCommandProvider.HandleTabCompletion] Input:", "input", input)
	completed, options := p.parser.GetTabCompletion(input)
	logging.Debug("[slashCommandProvider.HandleTabCompletion] After GetTabCompletion:", "completed", completed, "optionsCount", len(options))
	
	if options == nil {
		// Single match or no change
		return completed, nil, nil
	}
	
	// Convert options to completion items
	items := make([]dialog.CompletionItemI, len(options))
	for i, opt := range options {
		items[i] = &SlashCommandItem{
			CompletionItem: dialog.CompletionItem{
				Title: fmt.Sprintf("%s %s", opt.Icon, opt.Display),
				Value: opt.Value,
			},
			Description: opt.Description,
			Complete:    opt.Complete,
		}
	}
	
	return completed, items, nil
}

// ParseSlashCommand parses a slash command string
func ParseSlashCommand(input string) commands.SlashCommand {
	registry := commands.GetGlobalRegistry()
	parser := commands.NewCommandParser(registry)
	return parser.Parse(input)
}

// IsSlashCommand checks if input is a slash command
func IsSlashCommand(input string) bool {
	return strings.HasPrefix(input, "/")
}

// SlashCommandCompleteMsg is sent when a slash command should be completed
type SlashCommandCompleteMsg struct {
	Value string
}

// SlashCommandExecuteMsg is sent when a slash command should be executed
type SlashCommandExecuteMsg struct {
	Command commands.SlashCommand
}