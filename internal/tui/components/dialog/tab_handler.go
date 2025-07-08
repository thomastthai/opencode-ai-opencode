package dialog

import (
	"fmt"
	"os"
	
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
	logging.Info("[HandleTabKey] Input from pseudoSearchTextArea:", "input", input)
	
	// Also write to debug file for easier access
	debugLog := fmt.Sprintf("[HandleTabKey] Input: %q\n", input)
	os.WriteFile("/tmp/opencode-tab-debug.log", []byte(debugLog), 0644)
	
	// Try to get tab handler from provider
	if handler, ok := c.completionProvider.(TabHandler); ok {
		completed, options, err := handler.HandleTabCompletion(input)
		logging.Info("[HandleTabKey] After HandleTabCompletion:", "completed", completed, "hasOptions", options != nil)
		
		// Append to debug file
		debugLog2 := fmt.Sprintf("[HandleTabKey] Completed: %q (hasOptions: %v)\n", completed, options != nil)
		f, _ := os.OpenFile("/tmp/opencode-tab-debug.log", os.O_APPEND|os.O_WRONLY, 0644)
		if f != nil {
			f.WriteString(debugLog2)
			f.Close()
		}
		if err != nil {
			return nil
		}
		
		if options == nil {
			// Single match - complete it
			// Log what we're sending
			debugLog3 := fmt.Sprintf("[HandleTabKey] Sending completion message: original=%q new=%q\n", input, completed)
			f2, _ := os.OpenFile("/tmp/opencode-tab-debug.log", os.O_APPEND|os.O_WRONLY, 0644)
			if f2 != nil {
				f2.WriteString(debugLog3)
				f2.Close()
			}
			
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