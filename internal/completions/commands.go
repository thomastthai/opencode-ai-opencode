package completions

import (
	"strings"

	"github.com/opencode-ai/opencode/internal/commands"
	"github.com/opencode-ai/opencode/internal/tui/components/dialog"
)

// commandCompletionProvider provides command suggestions.
type commandCompletionProvider struct{}

// GetId returns the unique identifier for this provider.
func (p *commandCompletionProvider) GetId() string {
	return "commands"
}

// GetEntry returns the top-level entry for this provider.
func (p *commandCompletionProvider) GetEntry() dialog.CompletionItemI {
	return &dialog.CompletionItem{
		Title: "Commands",
		Value: "/",
	}
}

// GetChildEntries returns a list of commands matching the given query.
func (p *commandCompletionProvider) GetChildEntries(query string) ([]dialog.CompletionItemI, error) {
	var items []dialog.CompletionItemI
	allCommands := commands.GetGlobalRegistry().List()

	for _, cmd := range allCommands {
		// Filter commands based on the query.
		// The query will not include the leading slash.
		if strings.HasPrefix(cmd.ID(), query) {
			items = append(items, &dialog.CompletionItem{
				Title:       cmd.Name(),
				Value:       "/" + cmd.ID(),
			})
		}
	}
	return items, nil
}

// NewCommandCompletionProvider creates a new completion provider for commands.
func NewCommandCompletionProvider() dialog.CompletionProvider {
	return &commandCompletionProvider{}
}
