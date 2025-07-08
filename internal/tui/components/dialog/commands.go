package dialog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	utilComponents "github.com/opencode-ai/opencode/internal/tui/components/util"
	"github.com/opencode-ai/opencode/internal/tui/command"
	"github.com/opencode-ai/opencode/internal/tui/layout"
	"github.com/opencode-ai/opencode/internal/tui/styles"
	"github.com/opencode-ai/opencode/internal/tui/theme"
	"github.com/opencode-ai/opencode/internal/tui/util"
)

// Command represents a command that can be executed
type Command command.Command

// CommandGroup represents a group of commands with the same scope
type CommandGroup struct {
	Scope    command.CommandScope
	Commands []Command
}

func (ci Command) Render(selected bool, width int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	// Single line style for "Command - Description" format
	itemStyle := baseStyle.Width(width-4).
		Foreground(t.Text()).
		Background(t.Background()).
		Padding(0, 1)

	if selected {
		itemStyle = itemStyle.
			Background(t.Primary()).
			Foreground(t.Background()).
			Bold(true)
	}

	// Get command from underlying command.Command
	cmd := command.Command(ci)
	
	// Get emoji based on command type/category
	emoji := getCommandEmoji(ci)
	
	// Format as "emoji Command - Description"
	var text string
	if ci.Description != "" {
		// Command name style
		cmdStyle := lipgloss.NewStyle().Bold(true)
		if selected {
			cmdStyle = cmdStyle.Foreground(t.Background())
		} else {
			cmdStyle = cmdStyle.Foreground(t.Primary())
		}
		
		// Description style
		descStyle := lipgloss.NewStyle()
		if selected {
			descStyle = descStyle.Foreground(t.Background())
		} else {
			descStyle = descStyle.Foreground(t.TextMuted())
		}
		
		text = fmt.Sprintf("%s %s - %s", emoji, cmdStyle.Render(ci.Title), descStyle.Render(ci.Description))
		
		// Add aliases if any
		if len(cmd.Aliases) > 0 {
			aliasText := fmt.Sprintf(" (%s)", strings.Join(cmd.Aliases, ", "))
			text += descStyle.Render(aliasText)
		}
	} else {
		text = fmt.Sprintf("%s %s", emoji, ci.Title)
	}
	
	return itemStyle.Render(text)
}

// getCommandEmoji returns an appropriate emoji for the command based on its type
func getCommandEmoji(cmd Command) string {
	// Check for specific command patterns
	title := strings.ToLower(cmd.Title)
	
	// System/built-in commands
	if strings.HasPrefix(title, "/system") {
		return "🤖"
	}
	
	// Session commands
	if strings.HasPrefix(title, "/session") {
		return "💬"
	}
	
	// File commands
	if strings.HasPrefix(title, "/file") || strings.HasPrefix(title, "@") {
		return "📄"
	}
	
	// Project commands
	if strings.HasPrefix(title, "/project") {
		return "📁"
	}
	
	// Config commands
	if strings.HasPrefix(title, "/config") {
		return "⚙️"
	}
	
	// Auth commands
	if strings.HasPrefix(title, "/auth") {
		return "🔐"
	}
	
	// Help command
	if strings.HasPrefix(title, "/help") {
		return "❓"
	}
	
	// Git commands (custom)
	if strings.Contains(title, "git") || command.Command(cmd).Category == "git" {
		return "🌿"
	}
	
	// Test commands (custom)
	if strings.Contains(title, "test") || command.Command(cmd).Category == "testing" {
		return "🧪"
	}
	
	// Build/deployment commands (custom)
	if strings.Contains(title, "build") || strings.Contains(title, "deploy") || command.Command(cmd).Category == "deployment" {
		return "🚀"
	}
	
	// Default based on scope
	c := command.Command(cmd)
	switch c.Scope {
	case command.BuiltinScope:
		return "⚡"
	case command.UserScope:
		return "👤"
	case command.ProjectScope:
		return "📋"
	default:
		return "•"
	}
}

// CommandSelectedMsg is sent when a command is selected
type CommandSelectedMsg struct {
	Command Command
}

// CloseCommandDialogMsg is sent when the command dialog is closed
type CloseCommandDialogMsg struct{}

// ShowCommandDialogMsg is sent to request showing the command dialog
type ShowCommandDialogMsg struct{}

// CommandDialog interface for the command selection dialog
type CommandDialog interface {
	tea.Model
	layout.Bindings
	SetCommands(commands []command.Command)
}

type commandDialogCmp struct {
	listView        utilComponents.SimpleList[Command]
	searchInput     textinput.Model
	width           int
	height          int
	commands        []command.Command
	filteredCommands []Command
	groups          []CommandGroup
	showSearch      bool
	searchQuery     string
	watcher         *CommandWatcher
}

type commandKeyMap struct {
	Enter      key.Binding
	Escape     key.Binding
	Search     key.Binding
	ClearSearch key.Binding
}

var commandKeys = commandKeyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select command"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close/clear search"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search commands"),
	),
	ClearSearch: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "clear search"),
	),
}

func (c *commandDialogCmp) Init() tea.Cmd {
	// Initialize search input
	c.searchInput = textinput.New()
	c.searchInput.Placeholder = "Search commands..."
	c.searchInput.CharLimit = 50
	
	var cmds []tea.Cmd
	cmds = append(cmds, c.listView.Init())
	
	// Initialize and start command watcher
	watcher, err := NewCommandWatcher()
	if err == nil {
		c.watcher = watcher
		cmds = append(cmds, watcher.Start())
	}
	
	return tea.Batch(cmds...)
}

func (c *commandDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle search mode
		if c.showSearch {
			switch {
			case key.Matches(msg, commandKeys.Escape):
				if c.searchInput.Value() != "" {
					// Clear search
					c.searchInput.SetValue("")
					c.searchQuery = ""
					c.updateFilteredCommands()
				} else {
					// Close search mode
					c.showSearch = false
					c.searchInput.Blur()
				}
				return c, nil
			case key.Matches(msg, commandKeys.Enter):
				// Execute search or select command
				selectedItem, idx := c.listView.GetSelectedItem()
				if idx != -1 {
					return c, util.CmdHandler(CommandSelectedMsg{
						Command: selectedItem,
					})
				}
			default:
				// Update search input
				var cmd tea.Cmd
				c.searchInput, cmd = c.searchInput.Update(msg)
				cmds = append(cmds, cmd)
				
				// Update search query if changed
				newQuery := c.searchInput.Value()
				if newQuery != c.searchQuery {
					c.searchQuery = newQuery
					c.updateFilteredCommands()
				}
			}
		} else {
			// Handle normal mode
			switch {
			case key.Matches(msg, commandKeys.Enter):
				selectedItem, idx := c.listView.GetSelectedItem()
				if idx != -1 {
					return c, util.CmdHandler(CommandSelectedMsg{
						Command: selectedItem,
					})
				}
			case key.Matches(msg, commandKeys.Escape):
				// Stop watcher before closing
				if c.watcher != nil {
					c.watcher.Stop()
				}
				return c, util.CmdHandler(CloseCommandDialogMsg{})
			case key.Matches(msg, commandKeys.Search):
				// Enter search mode
				c.showSearch = true
				c.searchInput.Focus()
				return c, nil
			case key.Matches(msg, commandKeys.ClearSearch):
				// Clear search and filters
				c.searchInput.SetValue("")
				c.searchQuery = ""
				c.showSearch = false
				c.updateFilteredCommands()
				return c, nil
			}
		}
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		
	case CommandsReloadedMsg:
		// Handle command reload
		if msg.Error != nil {
			// Log error but continue with existing commands
			return c, nil
		}
		
		// Convert custom commands to command.Command format
		customCommands := make([]command.Command, 0, len(msg.Commands))
		for _, cmd := range msg.Commands {
			customCommands = append(customCommands, command.Command(cmd))
		}
		
		// Merge with built-in commands
		allCommands := append(c.getBuiltInCommands(), customCommands...)
		c.SetCommands(allCommands)
		
		// Refresh the view
		c.updateFilteredCommands()
		return c, nil
	}

	// Update list view only if not in search mode
	if !c.showSearch {
		u, cmd := c.listView.Update(msg)
		c.listView = u.(utilComponents.SimpleList[Command])
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

// updateFilteredCommands filters commands based on search query
func (c *commandDialogCmp) updateFilteredCommands() {
	if c.searchQuery == "" {
		// Show all commands grouped by scope
		c.filteredCommands = c.commandsToDialogCommands(c.commands)
	} else {
		// Filter commands by search query
		var filtered []command.Command
		for _, cmd := range c.commands {
			if cmd.MatchesSearch(c.searchQuery) {
				filtered = append(filtered, cmd)
			}
		}
		c.filteredCommands = c.commandsToDialogCommands(filtered)
	}
	
	c.listView.SetItems(c.filteredCommands)
}

// commandsToDialogCommands converts and groups commands for display
func (c *commandDialogCmp) commandsToDialogCommands(commands []command.Command) []Command {
	// Group commands by scope
	groups := make(map[command.CommandScope][]command.Command)
	for _, cmd := range commands {
		groups[cmd.Scope] = append(groups[cmd.Scope], cmd)
	}
	
	// Create ordered groups (built-in, user, project)
	var result []Command
	scopes := []command.CommandScope{command.BuiltinScope, command.UserScope, command.ProjectScope}
	
	for _, scope := range scopes {
		if scopeCommands, exists := groups[scope]; exists {
			// Sort commands within scope
			sort.Slice(scopeCommands, func(i, j int) bool {
				return scopeCommands[i].Title < scopeCommands[j].Title
			})
			
			// Add commands to result
			for _, cmd := range scopeCommands {
				result = append(result, Command(cmd))
			}
		}
	}
	
	return result
}

func (c *commandDialogCmp) View() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	maxWidth := 80  // Start with a reasonable width for "Command - Description" format

	// Calculate max width based on commands
	commands := c.listView.GetItems()
	for _, cmd := range commands {
		// Calculate full line width: "emoji Command - Description (aliases)"
		lineWidth := 3 // emoji + spaces
		lineWidth += len(cmd.Title)
		if cmd.Description != "" {
			lineWidth += 3 + len(cmd.Description) // " - " + description
		}
		if len(command.Command(cmd).Aliases) > 0 {
			lineWidth += len(fmt.Sprintf(" (%s)", strings.Join(command.Command(cmd).Aliases, ", ")))
		}
		if lineWidth > maxWidth-6 {
			maxWidth = lineWidth + 6
			if maxWidth > 120 {
				maxWidth = 120 // Cap at 120 chars
			}
		}
	}

	c.listView.SetMaxWidth(maxWidth)

	// Create title with search info
	titleText := "Commands"
	if c.searchQuery != "" {
		titleText = fmt.Sprintf("Commands (filtered: '%s')", c.searchQuery)
	}
	
	title := baseStyle.
		Foreground(t.Primary()).
		Bold(true).
		Width(maxWidth).
		Padding(0, 1).
		Render(titleText)

	// Add search input if in search mode
	var searchView string
	if c.showSearch {
		searchStyle := baseStyle.
			Width(maxWidth).
			Padding(0, 1).
			Foreground(t.Text()).
			Background(t.Background())
		
		searchView = searchStyle.Render("🔍 " + c.searchInput.View())
	}

	// Create grouped view with section headers
	var contentParts []string
	if c.searchQuery == "" && !c.showSearch {
		// Show grouped view
		contentParts = append(contentParts, c.renderGroupedCommands(maxWidth))
	} else {
		// Show filtered/search results
		contentParts = append(contentParts, baseStyle.Width(maxWidth).Render(c.listView.View()))
	}

	// Help text
	helpText := "↑↓ navigate • Enter select • / search • Esc close"
	if c.showSearch {
		helpText = "Type to search • Enter select • Esc clear/close"
	}
	
	helpView := baseStyle.
		Width(maxWidth).
		Padding(0, 1).
		Foreground(t.TextMuted()).
		Render(helpText)

	// Combine all parts
	var allParts []string
	allParts = append(allParts, title)
	if searchView != "" {
		allParts = append(allParts, searchView)
	}
	allParts = append(allParts, baseStyle.Width(maxWidth).Render(""))
	allParts = append(allParts, contentParts...)
	allParts = append(allParts, baseStyle.Width(maxWidth).Render(""))
	allParts = append(allParts, helpView)

	content := lipgloss.JoinVertical(lipgloss.Left, allParts...)

	return baseStyle.Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderBackground(t.Background()).
		BorderForeground(t.TextMuted()).
		Width(lipgloss.Width(content) + 4).
		Render(content)
}

// renderGroupedCommands renders commands grouped by scope with headers
func (c *commandDialogCmp) renderGroupedCommands(maxWidth int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()
	
	// Group commands by scope
	groups := make(map[command.CommandScope][]Command)
	commands := c.listView.GetItems()
	for _, cmd := range commands {
		scope := command.Command(cmd).Scope
		groups[scope] = append(groups[scope], cmd)
	}
	
	var sections []string
	scopes := []command.CommandScope{command.BuiltinScope, command.UserScope, command.ProjectScope}
	
	for _, scope := range scopes {
		if scopeCommands, exists := groups[scope]; exists && len(scopeCommands) > 0 {
			// Create scope header
			cmd := command.Command(scopeCommands[0])
			headerText := fmt.Sprintf("%s %s (%d)", cmd.GetIcon(), cmd.GetScopeDisplayName(), len(scopeCommands))
			
			header := baseStyle.
				Width(maxWidth).
				Padding(0, 1).
				Foreground(t.Primary()).
				Bold(true).
				Render(headerText)
			
			sections = append(sections, header)
			
			// Add commands in this group
			for _, scopeCmd := range scopeCommands {
				// Check if this command is selected
				_, selectedIdx := c.listView.GetSelectedItem()
				cmdIdx := c.findCommandIndex(scopeCmd)
				isSelected := cmdIdx == selectedIdx
				
				sections = append(sections, scopeCmd.Render(isSelected, maxWidth))
			}
			
			// Add spacing between groups
			sections = append(sections, baseStyle.Width(maxWidth).Render(""))
		}
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// findCommandIndex finds the index of a command in the list
func (c *commandDialogCmp) findCommandIndex(target Command) int {
	commands := c.listView.GetItems()
	for i, cmd := range commands {
		if cmd.ID == target.ID {
			return i
		}
	}
	return -1
}

func (c *commandDialogCmp) BindingKeys() []key.Binding {
	return layout.KeyMapToSlice(commandKeys)
}

func (c *commandDialogCmp) SetCommands(commands []command.Command) {
	c.commands = commands
	c.updateFilteredCommands()
}

// getBuiltInCommands retrieves built-in commands from the current command list
func (c *commandDialogCmp) getBuiltInCommands() []command.Command {
	var builtInCommands []command.Command
	for _, cmd := range c.commands {
		if cmd.Scope == command.BuiltinScope {
			builtInCommands = append(builtInCommands, cmd)
		}
	}
	return builtInCommands
}

// NewCommandDialogCmp creates a new command selection dialog
func NewCommandDialogCmp() CommandDialog {
	listView := utilComponents.NewSimpleList[Command](
		[]Command{},
		20, // Height for single-line command display
		"No commands available",
		true,
	)
	
	// Initialize search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search commands..."
	searchInput.CharLimit = 50
	
	return &commandDialogCmp{
		listView:    listView,
		searchInput: searchInput,
		commands:    []command.Command{},
		showSearch:  false,
		searchQuery: "",
	}
}
