package tui

import (
	"context"
	"strings"
	"time"
	
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opencode-ai/opencode/internal/app"
	"github.com/opencode-ai/opencode/internal/commands"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/llm/agent"
	"github.com/opencode-ai/opencode/internal/tui/command"
	"github.com/opencode-ai/opencode/internal/logging"
	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/opencode-ai/opencode/internal/session"
	"github.com/opencode-ai/opencode/internal/tui/components/chat"
	"github.com/opencode-ai/opencode/internal/tui/components/core"
	"github.com/opencode-ai/opencode/internal/tui/components/dialog"
	"github.com/opencode-ai/opencode/internal/tui/layout"
	"github.com/opencode-ai/opencode/internal/tui/page"
	"github.com/opencode-ai/opencode/internal/tui/theme"
	"github.com/opencode-ai/opencode/internal/tui/util"
)

type keyMap struct {
	Logs          key.Binding
	Quit          key.Binding
	Help          key.Binding
	SwitchSession key.Binding
	Commands      key.Binding
	Filepicker    key.Binding
	Models        key.Binding
	SwitchTheme   key.Binding
}

type startCompactSessionMsg struct{}

const (
	quitKey = "q"
)

var keys = keyMap{
	Logs: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "logs"),
	),

	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("ctrl+_", "ctrl+h"),
		key.WithHelp("ctrl+?", "toggle help"),
	),

	SwitchSession: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "switch session"),
	),

	Commands: key.NewBinding(
		key.WithKeys("ctrl+k"),
		key.WithHelp("ctrl+k", "commands"),
	),
	Filepicker: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("ctrl+f", "select files to upload"),
	),
	Models: key.NewBinding(
		key.WithKeys("ctrl+o"),
		key.WithHelp("ctrl+o", "model selection"),
	),

	SwitchTheme: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("ctrl+t", "switch theme"),
	),
}

var helpEsc = key.NewBinding(
	key.WithKeys("?"),
	key.WithHelp("?", "toggle help"),
)

var returnKey = key.NewBinding(
	key.WithKeys("esc"),
	key.WithHelp("esc", "close"),
)

var logsKeyReturnKey = key.NewBinding(
	key.WithKeys("esc", "backspace", quitKey),
	key.WithHelp("esc/q", "go back"),
)

type appModel struct {
	width, height   int
	currentPage     page.PageID
	previousPage    page.PageID
	pages           map[page.PageID]tea.Model
	loadedPages     map[page.PageID]bool
	status          core.StatusCmp
	app             *app.App
	selectedSession session.Session

	showPermissions bool
	permissions     dialog.PermissionDialogCmp

	showHelp bool
	help     dialog.HelpCmp

	showQuit bool
	quit     dialog.QuitDialog

	showSessionDialog bool
	sessionDialog     dialog.SessionDialog

	showCommandDialog bool
	commandDialog     dialog.CommandDialog
	commands          []command.Command

	showModelDialog bool
	modelDialog     dialog.ModelDialog

	showInitDialog bool
	initDialog     dialog.InitDialogCmp

	showFilepicker bool
	filepicker     dialog.FilepickerCmp

	showThemeDialog bool
	themeDialog     dialog.ThemeDialog

	showMultiArgumentsDialog bool
	multiArgumentsDialog     dialog.MultiArgumentsDialogCmp

	showOAuth2Dialog bool
	oauth2Dialog     dialog.OAuth2DialogCmp

	isCompacting      bool
	compactingMessage string
}

func (a appModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmd := a.pages[a.currentPage].Init()
	a.loadedPages[a.currentPage] = true
	cmds = append(cmds, cmd)
	cmd = a.status.Init()
	cmds = append(cmds, cmd)
	cmd = a.quit.Init()
	cmds = append(cmds, cmd)
	cmd = a.help.Init()
	cmds = append(cmds, cmd)
	cmd = a.sessionDialog.Init()
	cmds = append(cmds, cmd)
	cmd = a.commandDialog.Init()
	cmds = append(cmds, cmd)
	cmd = a.modelDialog.Init()
	cmds = append(cmds, cmd)
	cmd = a.initDialog.Init()
	cmds = append(cmds, cmd)
	cmd = a.filepicker.Init()
	cmds = append(cmds, cmd)
	cmd = a.themeDialog.Init()
	cmds = append(cmds, cmd)
	cmd = a.oauth2Dialog.Init()
	cmds = append(cmds, cmd)

	// Check if we should show the init dialog
	cmds = append(cmds, func() tea.Msg {
		shouldShow, err := config.ShouldShowInitDialog()
		if err != nil {
			return util.InfoMsg{
				Type: util.InfoTypeError,
				Msg:  "Failed to check init status: " + err.Error(),
			}
		}
		return dialog.ShowInitDialogMsg{Show: shouldShow}
	})

	return tea.Batch(cmds...)
}

func (a appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Always update the status bar
	s, _ := a.status.Update(msg)
	a.status = s.(core.StatusCmp)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		msg.Height -= 1 // Make space for the status bar
		a.width, a.height = msg.Width, msg.Height

		// Propagate the window size message to all components
		a.pages[a.currentPage], cmd = a.pages[a.currentPage].Update(msg)
		cmds = append(cmds, cmd)

		prm, permCmd := a.permissions.Update(msg)
		a.permissions = prm.(dialog.PermissionDialogCmp)
		cmds = append(cmds, permCmd)

		help, helpCmd := a.help.Update(msg)
		a.help = help.(dialog.HelpCmp)
		cmds = append(cmds, helpCmd)

		session, sessionCmd := a.sessionDialog.Update(msg)
		a.sessionDialog = session.(dialog.SessionDialog)
		cmds = append(cmds, sessionCmd)

		command, commandCmd := a.commandDialog.Update(msg)
		a.commandDialog = command.(dialog.CommandDialog)
		cmds = append(cmds, commandCmd)

		filepicker, filepickerCmd := a.filepicker.Update(msg)
		a.filepicker = filepicker.(dialog.FilepickerCmp)
		cmds = append(cmds, filepickerCmd)

		a.initDialog.SetSize(msg.Width, msg.Height)

		if a.showMultiArgumentsDialog {
			a.multiArgumentsDialog.SetSize(msg.Width, msg.Height)
			args, argsCmd := a.multiArgumentsDialog.Update(msg)
			a.multiArgumentsDialog = args.(dialog.MultiArgumentsDialogCmp)
			cmds = append(cmds, argsCmd, a.multiArgumentsDialog.Init())
		}
		return a, tea.Batch(cmds...)

	// Handle command selection - CRITICAL FIX
	case dialog.CommandSelectedMsg:
		a.showCommandDialog = false
		cmd := command.Command(msg.Command)
		
		// Track command usage
		a.updateCommandUsage(cmd.ID)
		
		// Handle built-in commands with handlers
		if cmd.Handler != nil {
			return a, cmd.Handler(cmd)
		}
		
		// Handle custom commands with placeholders
		if cmd.HasPlaceholders() {
			// Show arguments dialog for commands with placeholders
			return a, a.showArgumentsDialog(cmd)
		}
		
		// Handle custom commands without placeholders
		if cmd.Content != "" {
			return a, util.CmdHandler(dialog.CommandRunCustomMsg{
				Content: cmd.Content,
			})
		}
		
		return a, nil

	case dialog.CloseCommandDialogMsg:
		a.showCommandDialog = false
		return a, nil

	case dialog.OAuth2DialogMsg:
		if a.showOAuth2Dialog {
			oauth2, cmd := a.oauth2Dialog.Update(msg)
			a.oauth2Dialog = oauth2.(dialog.OAuth2DialogCmp)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)

	case dialog.ShowOAuth2DialogMsg:
		a.showOAuth2Dialog = true
		cmd = a.oauth2Dialog.StartOAuth2(msg.Provider)
		cmds = append(cmds, cmd)
		return a, tea.Batch(cmds...)

	case string:
		if msg == "close_oauth2_dialog" {
			a.showOAuth2Dialog = false
			return a, nil
		}

	// Handle key messages for global shortcuts
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Commands) && !a.showCommandDialog:
			a.showCommandDialog = true
			a.commandDialog.SetCommands(a.commands)
			return a, nil
		case key.Matches(msg, keys.Help) && !a.showHelp:
			a.showHelp = true
			return a, nil
		case key.Matches(msg, keys.SwitchSession) && !a.showSessionDialog:
			a.showSessionDialog = true
			return a, nil
		case key.Matches(msg, keys.Quit):
			return a, tea.Quit
		case msg.Type == tea.KeyEsc:
			// Close any open dialogs
			if a.showHelp {
				a.showHelp = false
				return a, nil
			}
		}

	// Handle compact session message
	case startCompactSessionMsg:
		// Start compacting the current session
		a.isCompacting = true
		a.compactingMessage = "Starting summarization..."

		// Get the current session ID from the chat page
		chatPage, ok := a.pages[page.ChatPage].(interface{ GetCurrentSessionID() string })
		if !ok {
			a.isCompacting = false
			logging.Error("Chat page doesn't implement GetCurrentSessionID")
			return a, util.ReportWarn("Unable to get current session")
		}
		
		sessionID := chatPage.GetCurrentSessionID()
		if sessionID == "" {
			a.isCompacting = false
			return a, util.ReportWarn("No active session to summarize")
		}
		
		// Start the summarization process
		return a, func() tea.Msg {
			ctx := context.Background()
			if err := a.app.CoderAgent.Summarize(ctx, sessionID); err != nil {
				return util.ReportError(err)
			}
			return nil
		}
		
	// Handle agent events for summarization
	case pubsub.Event[agent.AgentEvent]:
		payload := msg.Payload
		if payload.Type == agent.AgentEventTypeSummarize {
			if payload.Error != nil {
				a.isCompacting = false
				return a, util.ReportError(payload.Error)
			}
			
			a.compactingMessage = payload.Progress
			
			if payload.Done {
				a.isCompacting = false
				return a, util.ReportInfo("Session compacted successfully")
			}
		}
		return a, nil
		
	// Handle compact request from chat page
	case page.CompactSessionRequestMsg:
		// Trigger the compact process
		return a, func() tea.Msg {
			return startCompactSessionMsg{}
		}
	}

	// Handle all other messages
	if a.showFilepicker {
		f, filepickerCmd := a.filepicker.Update(msg)
		a.filepicker = f.(dialog.FilepickerCmp)
		cmds = append(cmds, filepickerCmd)
		if _, ok := msg.(tea.KeyMsg); ok {
			return a, tea.Batch(cmds...)
		}
	}

	if a.showQuit {
		q, quitCmd := a.quit.Update(msg)
		a.quit = q.(dialog.QuitDialog)
		cmds = append(cmds, quitCmd)
		if _, ok := msg.(tea.KeyMsg); ok {
			return a, tea.Batch(cmds...)
		}
	}

	if a.showPermissions {
		d, permissionsCmd := a.permissions.Update(msg)
		a.permissions = d.(dialog.PermissionDialogCmp)
		cmds = append(cmds, permissionsCmd)
		if _, ok := msg.(tea.KeyMsg); ok {
			return a, tea.Batch(cmds...)
		}
	}

	if a.showSessionDialog {
		d, sessionCmd := a.sessionDialog.Update(msg)
		a.sessionDialog = d.(dialog.SessionDialog)
		cmds = append(cmds, sessionCmd)
		if _, ok := msg.(tea.KeyMsg); ok {
			return a, tea.Batch(cmds...)
		}
	}

	if a.showCommandDialog {
		d, commandCmd := a.commandDialog.Update(msg)
		a.commandDialog = d.(dialog.CommandDialog)
		cmds = append(cmds, commandCmd)
		if _, ok := msg.(tea.KeyMsg); ok {
			return a, tea.Batch(cmds...)
		}
	}

	if a.showModelDialog {
		d, modelCmd := a.modelDialog.Update(msg)
		a.modelDialog = d.(dialog.ModelDialog)
		cmds = append(cmds, modelCmd)
		if _, ok := msg.(tea.KeyMsg); ok {
			return a, tea.Batch(cmds...)
		}
	}

	if a.showInitDialog {
		d, initCmd := a.initDialog.Update(msg)
		a.initDialog = d.(dialog.InitDialogCmp)
		cmds = append(cmds, initCmd)
		if _, ok := msg.(tea.KeyMsg); ok {
			return a, tea.Batch(cmds...)
		}
	}

	if a.showThemeDialog {
		d, themeCmd := a.themeDialog.Update(msg)
		a.themeDialog = d.(dialog.ThemeDialog)
		cmds = append(cmds, themeCmd)
		if _, ok := msg.(tea.KeyMsg); ok {
			return a, tea.Batch(cmds...)
		}
	}

	if a.showOAuth2Dialog {
		d, oauth2Cmd := a.oauth2Dialog.Update(msg)
		a.oauth2Dialog = d.(dialog.OAuth2DialogCmp)
		cmds = append(cmds, oauth2Cmd)
		if _, ok := msg.(tea.KeyMsg); ok {
			return a, tea.Batch(cmds...)
		}
	}

	// If no dialog is active, pass the message to the current page
	a.pages[a.currentPage], cmd = a.pages[a.currentPage].Update(msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

// RegisterCommand adds a command to the command dialog
func (a *appModel) RegisterCommand(cmd command.Command) {
	a.commands = append(a.commands, cmd)
}

func (a *appModel) findCommand(id string) (command.Command, bool) {
	for _, cmd := range a.commands {
		if cmd.ID == id {
			return cmd, true
		}
	}
	return command.Command{}, false
}

// showArgumentsDialog shows the arguments dialog for commands with placeholders
func (a *appModel) showArgumentsDialog(cmd command.Command) tea.Cmd {
	// For now, just execute the command without arguments
	// TODO: Implement proper arguments dialog
	if cmd.Content != "" {
		return util.CmdHandler(dialog.CommandRunCustomMsg{
			Content: cmd.Content,
		})
	}
	return nil
}

// convertRegistryCommand converts a registry command to a TUI command
func convertRegistryCommand(regCmd commands.Command) command.Command {
	return command.Command{
		ID:          regCmd.ID(),
		Title:       regCmd.Name(),
		Description: regCmd.Description(),
		Content:     "", // Registry commands don't have content
		Scope:       command.BuiltinScope, // Registry commands are built-in
		Category:    regCmd.Category(),
		Aliases:     regCmd.GetAliases(),
		Handler:     createBuiltinHandler(regCmd),
	}
}

// createBuiltinHandler creates a handler for built-in commands
func createBuiltinHandler(regCmd commands.Command) func(cmd command.Command) tea.Cmd {
	return func(cmd command.Command) tea.Cmd {
		switch regCmd.ID() {
		case "init":
			prompt := `Please analyze this codebase and create a OpenCode.md file containing:
1. Build/lint/test commands - especially for running a single test
2. Code style guidelines including imports, formatting, types, naming conventions, error handling, etc.

The file you create will be given to agentic coding agents (such as yourself) that operate in this repository. Make it about 20 lines long.
If there's already a opencode.md, improve it.
If there are Cursor rules (in .cursor/rules/ or .cursorrules) or Copilot rules (in .github/copilot-instructions.md), make sure to include them.`
			return util.CmdHandler(chat.SendMsg{
				Text: prompt,
			})
		case "compact":
			return func() tea.Msg {
				return startCompactSessionMsg{}
			}
		case "exit", "quit", "q":
			return tea.Quit
		case "clear", "cls", "new":
			return util.CmdHandler(chat.SessionClearedMsg{})
		case "help", "h":
			// Show help dialog or help content
			return nil
		case "list", "ls", "commands":
			// This could show command list in a message
			return nil
		case "login gemini":
			// Handle Gemini OAuth2 login
			return util.CmdHandler(dialog.ShowOAuth2DialogMsg{
				Provider: dialog.OAuth2ProviderGemini,
			})
		default:
			// For other built-in commands, try to execute via registry
			return nil
		}
	}
}

// determineCommandScope determines if a custom command is user or project scope
func determineCommandScope(commandID string) command.CommandScope {
	if strings.HasPrefix(commandID, "user:") {
		return command.UserScope
	} else if strings.HasPrefix(commandID, "project:") {
		return command.ProjectScope
	}
	// Default to user scope for backwards compatibility
	return command.UserScope
}

// updateCommandUsage updates the last used time for a command
func (a *appModel) updateCommandUsage(commandID string) {
	for i, cmd := range a.commands {
		if cmd.ID == commandID {
			a.commands[i].LastUsed = time.Now()
			break
		}
	}
}

// getRecentlyUsedCommands returns commands sorted by recently used
func (a *appModel) getRecentlyUsedCommands() []command.Command {
	var recentCommands []command.Command
	
	// Get commands that have been used (have LastUsed set)
	for _, cmd := range a.commands {
		if !cmd.LastUsed.IsZero() {
			recentCommands = append(recentCommands, cmd)
		}
	}
	
	// Sort by most recently used
	for i := 0; i < len(recentCommands)-1; i++ {
		for j := i + 1; j < len(recentCommands); j++ {
			if recentCommands[i].LastUsed.Before(recentCommands[j].LastUsed) {
				recentCommands[i], recentCommands[j] = recentCommands[j], recentCommands[i]
			}
		}
	}
	
	// Return top 5 most recent
	if len(recentCommands) > 5 {
		recentCommands = recentCommands[:5]
	}
	
	return recentCommands
}

func (a *appModel) moveToPage(pageID page.PageID) tea.Cmd {
	if a.app.CoderAgent.IsBusy() {
		// For now we don't move to any page if the agent is busy
		return util.ReportWarn("Agent is busy, please wait...")
	}

	var cmds []tea.Cmd
	if _, ok := a.loadedPages[pageID]; !ok {
		cmd := a.pages[pageID].Init()
		cmds = append(cmds, cmd)
		a.loadedPages[pageID] = true
	}
	a.previousPage = a.currentPage
	a.currentPage = pageID
	if sizable, ok := a.pages[a.currentPage].(layout.Sizeable); ok {
		cmd := sizable.SetSize(a.width, a.height)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (a appModel) View() string {
	components := []string{
		a.pages[a.currentPage].View(),
	}

	components = append(components, a.status.View())

	appView := lipgloss.JoinVertical(lipgloss.Top, components...)

	if a.showPermissions {
		overlay := a.permissions.View()
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)
	}

	if a.showFilepicker {
		overlay := a.filepicker.View()
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)

	}

	// Show compacting status overlay
	if a.isCompacting {
		t := theme.CurrentTheme()
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.BorderFocused()).
			BorderBackground(t.Background()).
			Padding(1, 2).
			Background(t.Background()).
			Foreground(t.Text())

		overlay := style.Render("Summarizing\n" + a.compactingMessage)
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)
	}

	if a.showHelp {
		bindings := layout.KeyMapToSlice(keys)
		if p, ok := a.pages[a.currentPage].(layout.Bindings); ok {
			bindings = append(bindings, p.BindingKeys()...)
		}
		if a.showPermissions {
			bindings = append(bindings, a.permissions.BindingKeys()...)
		}
		if a.currentPage == page.LogsPage {
			bindings = append(bindings, logsKeyReturnKey)
		}
		if !a.app.CoderAgent.IsBusy() {
			bindings = append(bindings, helpEsc)
		}
		a.help.SetBindings(bindings)

		overlay := a.help.View()
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)
	}

	if a.showQuit {
		overlay := a.quit.View()
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)
	}

	if a.showSessionDialog {
		overlay := a.sessionDialog.View()
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)
	}

	if a.showModelDialog {
		overlay := a.modelDialog.View()
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)
	}

	if a.showCommandDialog {
		overlay := a.commandDialog.View()
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)
	}

	if a.showInitDialog {
		overlay := a.initDialog.View()
		appView = layout.PlaceOverlay(
			a.width/2-lipgloss.Width(overlay)/2,
			a.height/2-lipgloss.Height(overlay)/2,
			overlay,
			appView,
			true,
		)
	}

	if a.showThemeDialog {
		overlay := a.themeDialog.View()
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)
	}

	if a.showMultiArgumentsDialog {
		overlay := a.multiArgumentsDialog.View()
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)
	}

	if a.showOAuth2Dialog {
		overlay := a.oauth2Dialog.View()
		row := lipgloss.Height(appView) / 2
		row -= lipgloss.Height(overlay) / 2
		col := lipgloss.Width(appView) / 2
		col -= lipgloss.Width(overlay) / 2
		appView = layout.PlaceOverlay(
			col,
			row,
			overlay,
			appView,
			true,
		)
	}

	return appView
}

func New(app *app.App) tea.Model {
	startPage := page.ChatPage
	model := &appModel{
		currentPage:   startPage,
		loadedPages:   make(map[page.PageID]bool),
		status:        core.NewStatusCmp(app.LSPClients),
		help:          dialog.NewHelpCmp(),
		quit:          dialog.NewQuitCmp(),
		sessionDialog: dialog.NewSessionDialogCmp(),
		commandDialog: dialog.NewCommandDialogCmp(),
		modelDialog:   dialog.NewModelDialogCmp(),
		permissions:   dialog.NewPermissionDialogCmp(),
		initDialog:    dialog.NewInitDialogCmp(),
		themeDialog:   dialog.NewThemeDialogCmp(),
		oauth2Dialog:  dialog.NewOAuth2DialogCmp(),
		app:           app,
		commands:      []command.Command{},
		pages: map[page.PageID]tea.Model{
			page.ChatPage: page.NewChatPage(app),
			page.LogsPage: page.NewLogsPage(),
		},
		filepicker: dialog.NewFilepickerCmp(app),
	}

	// Load commands from registry system
	registry := commands.GetGlobalRegistry()
	registryCommands := registry.List()
	
	// Convert registry commands to TUI commands
	for _, regCmd := range registryCommands {
		tuiCmd := convertRegistryCommand(regCmd)
		model.RegisterCommand(tuiCmd)
	}
	
	// Load custom commands
	customCommands, err := dialog.LoadCustomCommands()
	if err != nil {
		logging.Warn("Failed to load custom commands", "error", err)
	} else {
		for _, cmd := range customCommands {
			// Convert custom commands and set their scope
			tuiCmd := command.Command(cmd)
			tuiCmd.Scope = determineCommandScope(cmd.ID)
			tuiCmd.Source = "custom"
			model.RegisterCommand(tuiCmd)
		}
	}

	return model
}
