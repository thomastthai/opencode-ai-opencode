package dialog

import (
	"strings"
	
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/opencode-ai/opencode/internal/logging"
)

// SlashCommandCompleteMsg is sent when a slash command completion is selected
// This replaces CompletionSelectedMsg for progressive command building
type SlashCommandCompleteMsg struct {
	OriginalValue string // The original text in the editor
	NewValue      string // The new text to set in the editor
	CursorPos     int    // Where to place the cursor
	KeepOpen      bool   // Whether to keep the completion dialog open
}

// HandleSlashCommandCompletion handles slash command completions with progressive building
func HandleSlashCommandCompletion(provider CompletionProvider, editorValue string, selectedItem CompletionItemI) tea.Cmd {
	// Check if this is a slash command provider
	if provider.GetId() != "slash-commands" {
		// Fall back to standard completion
		return tea.Batch(
			func() tea.Msg {
				return CompletionSelectedMsg{
					SearchString:    editorValue,
					CompletionValue: selectedItem.GetValue(),
				}
			},
			func() tea.Msg {
				return CompletionDialogCloseMsg{}
			},
		)
	}

	// For slash commands, check if this is a SlashCommandItem with complete value
	var newValue string
	if slashItem, ok := selectedItem.(interface{ GetCompleteValue() string }); ok {
		// Use the complete command text
		newValue = slashItem.GetCompleteValue()
	} else {
		// Fall back to the item value
		newValue = selectedItem.GetValue()
	}
	
	// Simple check to see if we should keep the dialog open
	// Keep open if the value ends with a space (indicating more input expected)
	keepOpen := len(newValue) > 0 && newValue[len(newValue)-1] == ' '
	
	return func() tea.Msg {
		return SlashCommandCompleteMsg{
			OriginalValue: editorValue,
			NewValue:      newValue,
			CursorPos:     len(newValue),
			KeepOpen:      keepOpen,
		}
	}
}


// completionDialogKeys with tab support
var slashCompletionDialogKeys = struct {
	Complete key.Binding
	Cancel   key.Binding
	Tab      key.Binding
}{
	Complete: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("space", "esc", "backspace"),
		key.WithHelp("esc", "close"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "complete"),
	),
}

// UpdateCompletionDialogForSlashCommands modifies the completion dialog update behavior
// This should be called from the completion dialog's Update method
func UpdateCompletionDialogForSlashCommands(c *completionDialogCmp, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, slashCompletionDialogKeys.Tab) {
			// Handle tab completion
			cmd := HandleTabKey(c)
			if cmd != nil {
				return c, cmd
			}
		}
	case SlashCommandCompleteMsg:
		// Update the search text area with the new value
		c.pseudoSearchTextArea.SetValue(msg.NewValue)
		c.query = msg.NewValue
		
		if !msg.KeepOpen {
			return c, c.close()
		}
		
		// Update completions for the new value
		// Strip the leading slash if present since GetChildEntries expects query without it
		query := msg.NewValue
		if strings.HasPrefix(query, "/") {
			query = query[1:]
		}
		items, err := c.completionProvider.GetChildEntries(query)
		if err != nil {
			logging.Error("Failed to get completions", err)
			return c, nil
		}
		
		c.listView.SetItems(items)
		
		// If no more completions and we have a complete command, close
		if len(items) == 0 && !msg.KeepOpen {
			return c, c.close()
		}
		
		return c, nil
	}
	
	return nil, nil
}