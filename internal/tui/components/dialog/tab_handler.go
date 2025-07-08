package dialog

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opencode-ai/opencode/internal/logging"
)

// TabHandler is an interface for providers that support tab completion
type TabHandler interface {
	HandleTabCompletion(input string) (string, []CompletionItemI, error)
}

// HandleTabKey handles tab key press in the completion dialog
func HandleTabKey(c *completionDialogCmp) tea.Cmd {
	if c.completionProvider.GetId() != "slash-commands" {
		// Tab not supported for non-slash providers
		return nil
	}
	
	// Get current input value
	input := c.pseudoSearchTextArea.Value()
	logging.Debug("[HandleTabKey] Input from pseudoSearchTextArea:", "input", input)
	
	// Try to get tab handler from provider
	if handler, ok := c.completionProvider.(TabHandler); ok {
		completed, options, err := handler.HandleTabCompletion(input)
		logging.Debug("[HandleTabKey] After HandleTabCompletion:", "completed", completed, "hasOptions", options != nil)
		if err != nil {
			return nil
		}
		
		if options == nil {
			// Single match - complete it
			return func() tea.Msg {
				return SlashCommandCompleteMsg{
					OriginalValue: input,
					NewValue:      completed,
					CursorPos:     len(completed),
					KeepOpen:      true, // Keep open for further completions
				}
			}
		}
		
		// Multiple matches - update the list
		c.listView.SetItems(options)
	}
	
	return nil
}