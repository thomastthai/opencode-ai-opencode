package dialog

import (
	"fmt"
	"os"
	
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opencode-ai/opencode/internal/logging"
	utilComponents "github.com/opencode-ai/opencode/internal/tui/components/util"
	"github.com/opencode-ai/opencode/internal/tui/layout"
	"github.com/opencode-ai/opencode/internal/tui/styles"
	"github.com/opencode-ai/opencode/internal/tui/theme"
	"github.com/opencode-ai/opencode/internal/tui/util"
)

type CompletionItem struct {
	title string
	Title string
	Value string
}

type CompletionItemI interface {
	utilComponents.SimpleListItem
	GetValue() string
	DisplayValue() string
}

func (ci *CompletionItem) Render(selected bool, width int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	itemStyle := baseStyle.
		Width(width).
		Padding(0, 1)

	if selected {
		itemStyle = itemStyle.
			Background(t.Background()).
			Foreground(t.Primary()).
			Bold(true)
	}

	title := itemStyle.Render(
		ci.GetValue(),
	)

	return title
}

func (ci *CompletionItem) DisplayValue() string {
	return ci.Title
}

func (ci *CompletionItem) GetValue() string {
	return ci.Value
}

func NewCompletionItem(completionItem CompletionItem) CompletionItemI {
	return &completionItem
}

type CompletionProvider interface {
	GetId() string
	GetEntry() CompletionItemI
	GetChildEntries(query string) ([]CompletionItemI, error)
}

type CompletionSelectedMsg struct {
	SearchString    string
	CompletionValue string
}

type CompletionDialogCompleteItemMsg struct {
	Value string
}

type CompletionDialogCloseMsg struct{}

type CompletionDialog interface {
	tea.Model
	layout.Bindings
	SetWidth(width int)
	SetProvider(provider CompletionProvider)
	GetId() string
	// Test helper methods
	GetListItems() []CompletionItemI
	GetEmptyMessage() string
}

func (c *completionDialogCmp) GetId() string {
	return c.completionProvider.GetId()
}

// GetListItems returns the current items in the completion list (for testing)
func (c *completionDialogCmp) GetListItems() []CompletionItemI {
	return c.listView.GetItems()
}

// GetEmptyMessage returns the current empty message (for testing)
func (c *completionDialogCmp) GetEmptyMessage() string {
	return c.listView.GetEmptyMessage()
}


func (c *completionDialogCmp) SetProvider(provider CompletionProvider) {
	c.completionProvider = provider
	// Reset query when switching providers
	c.query = ""
	// Update the empty message based on provider type
	emptyMsg := "No matches found"
	if provider.GetId() == "commands" {
		emptyMsg = "No commands found"
	} else if provider.GetId() == "files" {
		emptyMsg = "No file matches found"
	} else if provider.GetId() == "slash-commands" {
		emptyMsg = "No command matches found"
	}
	// Update the list with new provider's items
	items, err := provider.GetChildEntries("")
	if err != nil {
		logging.Error("Failed to get child entries", err)
	}
	c.listView.SetItems(items)
	c.listView.SetEmptyMessage(emptyMsg)
}


type completionDialogCmp struct {
	query                string
	completionProvider   CompletionProvider
	width                int
	height               int
	pseudoSearchTextArea textarea.Model
	listView             utilComponents.SimpleList[CompletionItemI]
}

type completionDialogKeyMap struct {
	Complete key.Binding
	Cancel   key.Binding
}

var completionDialogKeys = completionDialogKeyMap{
	Complete: key.NewBinding(
		key.WithKeys("tab", "enter"),
	),
	Cancel: key.NewBinding(
		key.WithKeys(" ", "esc", "backspace"),
	),
}

func (c *completionDialogCmp) Init() tea.Cmd {
	return nil
}

func (c *completionDialogCmp) complete(item CompletionItemI) tea.Cmd {
	value := c.pseudoSearchTextArea.Value()

	if value == "" {
		return nil
	}

	// Use enhanced handling for slash commands
	return HandleSlashCommandCompletion(c.completionProvider, value, item)
}

func (c *completionDialogCmp) close() tea.Cmd {
	c.listView.SetItems([]CompletionItemI{})
	c.pseudoSearchTextArea.Reset()
	c.pseudoSearchTextArea.Blur()

	return util.CmdHandler(CompletionDialogCloseMsg{})
}

func (c *completionDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Check for slash command specific handling first
	if slashModel, slashCmd := UpdateCompletionDialogForSlashCommands(c, msg); slashModel != nil {
		return slashModel, slashCmd
	}
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if c.pseudoSearchTextArea.Focused() {

			if !key.Matches(msg, completionDialogKeys.Complete) {

				// Check for tab key first
				if key.Matches(msg, slashCompletionDialogKeys.Tab) {
					// Tab is handled by UpdateCompletionDialogForSlashCommands
					return c, tea.Batch(cmds...)
				}
				
				var cmd tea.Cmd
				c.pseudoSearchTextArea, cmd = c.pseudoSearchTextArea.Update(msg)
				cmds = append(cmds, cmd)

				var query string
				fullValue := c.pseudoSearchTextArea.Value()
				
				// Log textarea value updates
				debugLog := fmt.Sprintf("[completionDialog] TextArea updated: value=%q (key: %v)\n", fullValue, msg.String())
				if f, err := os.OpenFile("/tmp/opencode-tab-debug.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644); err == nil {
					f.WriteString(debugLog)
					f.Close()
				}
				// Extract query by removing the trigger character (/ or @)
				if len(fullValue) > 0 && (fullValue[0] == '/' || fullValue[0] == '@') {
					query = fullValue[1:]
				} else {
					query = fullValue
				}

				if query != c.query {
					logging.Debug("[completionDialogCmp] Query update:", "fullValue", fullValue, "query", query, "previousQuery", c.query)
					items, err := c.completionProvider.GetChildEntries(query)
					if err != nil {
						logging.Error("Failed to get child entries", err)
					}

					c.listView.SetItems(items)
					c.query = query
				}

				u, cmd := c.listView.Update(msg)
				c.listView = u.(utilComponents.SimpleList[CompletionItemI])

				cmds = append(cmds, cmd)
			}

			switch {
			case key.Matches(msg, completionDialogKeys.Complete):
				item, i := c.listView.GetSelectedItem()
				if i == -1 {
					return c, nil
				}

				cmd := c.complete(item)

				return c, cmd
			case key.Matches(msg, completionDialogKeys.Cancel):
				// For backspace, check if we should close based on the current value
				// The textarea has already processed the backspace at this point
				if msg.String() == "backspace" {
					// Log the value after backspace
					currentValue := c.pseudoSearchTextArea.Value()
					debugLog := fmt.Sprintf("[completionDialog] After backspace: value=%q\n", currentValue)
					if f, err := os.OpenFile("/tmp/opencode-tab-debug.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644); err == nil {
						f.WriteString(debugLog)
						f.Close()
					}
					
					// If after backspace we only have the trigger character, close the dialog
					// This prevents the issue where typing "/" again would create "//"
					if len(currentValue) == 1 && (currentValue == "/" || currentValue == "@") {
						return c, c.close()
					}
					// If we have more than just the trigger character, keep dialog open
					if len(currentValue) > 1 {
						return c, tea.Batch(cmds...)
					}
				}
				// For other cancel keys (space, esc) or empty backspace, close
				return c, c.close()
			}

			return c, tea.Batch(cmds...)
		} else {
			// Dialog is not focused, initialize it with the trigger character
			triggerChar := msg.String()
			if triggerChar == "/" || triggerChar == "@" {
				// Log when dialog is initialized
				debugLog := fmt.Sprintf("[completionDialog] Initializing with trigger: %q\n", triggerChar)
				if f, err := os.OpenFile("/tmp/opencode-tab-debug.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644); err == nil {
					f.WriteString(debugLog)
					f.Close()
				}
				
				// Load initial completions for empty query
				items, err := c.completionProvider.GetChildEntries("")
				if err != nil {
					logging.Error("Failed to get child entries", err)
				}

				c.listView.SetItems(items)
				c.pseudoSearchTextArea.SetValue(triggerChar)
				c.query = "" // Reset query to ensure proper state
				return c, c.pseudoSearchTextArea.Focus()
			}
		}
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
	}

	return c, tea.Batch(cmds...)
}

func (c *completionDialogCmp) View() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	maxWidth := 40

	completions := c.listView.GetItems()

	for _, cmd := range completions {
		title := cmd.DisplayValue()
		if len(title) > maxWidth-4 {
			maxWidth = len(title) + 4
		}
	}

	c.listView.SetMaxWidth(maxWidth)

	return baseStyle.Padding(0, 0).
		Border(lipgloss.NormalBorder()).
		BorderBottom(false).
		BorderRight(false).
		BorderLeft(false).
		BorderBackground(t.Background()).
		BorderForeground(t.TextMuted()).
		Width(c.width).
		Render(c.listView.View())
}

func (c *completionDialogCmp) SetWidth(width int) {
	c.width = width
}

func (c *completionDialogCmp) BindingKeys() []key.Binding {
	return layout.KeyMapToSlice(completionDialogKeys)
}

func NewCompletionDialogCmp(completionProvider CompletionProvider) CompletionDialog {
	ti := textarea.New()

	items, err := completionProvider.GetChildEntries("")
	if err != nil {
		logging.Error("Failed to get child entries", err)
	}

	li := utilComponents.NewSimpleList(
		items,
		7,
		"No file matches found",
		true, // Enable vim navigation keys (j/k)
	)

	return &completionDialogCmp{
		query:                "",
		completionProvider:   completionProvider,
		pseudoSearchTextArea: ti,
		listView:             li,
	}
}
