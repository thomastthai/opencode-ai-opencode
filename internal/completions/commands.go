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

	parts := strings.Split(query, ":")
	if len(parts) == 1 {
		// Top-level command search
		for _, cmd := range allCommands {
			if cmd.GetParent() == nil && strings.HasPrefix(cmd.ID(), query) {
				items = append(items, &dialog.CompletionItem{
					Title: cmd.Name(),
					Value: "/" + cmd.ID(),
				})
			}
		}
	} else {
		// Sub-command search
		parentID := strings.Join(parts[:len(parts)-1], ":")
		subQuery := parts[len(parts)-1]

		parentCmd, found := commands.GetGlobalRegistry().Get(parentID)
		if found {
			for _, subCmd := range parentCmd.GetSubCommands() {
				if strings.HasPrefix(subCmd.ID(), parentID+":"+subQuery) {
					items = append(items, &dialog.CompletionItem{
						Title: subCmd.Name(),
						Value: "/" + subCmd.ID(),
					})
				}
			}
		}
	}

	return items, nil
}

// NewCommandCompletionProvider creates a new completion provider for commands.
func NewCommandCompletionProvider() dialog.CompletionProvider {
	return &commandCompletionProvider{}
}
