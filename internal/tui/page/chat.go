package page

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opencode-ai/opencode/internal/app"
	"github.com/opencode-ai/opencode/internal/commands"
	"github.com/opencode-ai/opencode/internal/completions"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/message"
	"github.com/opencode-ai/opencode/internal/session"
	"github.com/opencode-ai/opencode/internal/tui/components/chat"
	"github.com/opencode-ai/opencode/internal/tui/components/dialog"
	recoveryDialog "github.com/opencode-ai/opencode/internal/tui/components/recovery"
	"github.com/opencode-ai/opencode/internal/tui/layout"
	"github.com/opencode-ai/opencode/internal/tui/util"
)

var ChatPage PageID = "chat"

// CompactSessionRequestMsg is sent to trigger session compacting
type CompactSessionRequestMsg struct{}

type chatPage struct {
	app                     *app.App
	editor               *chat.EditorCmp
	messages                layout.Container
	layout                  layout.SplitPaneLayout
	session                 session.Session
	completionDialog        dialog.CompletionDialog
	showCompletionDialog    bool
	recoveryDialog          *recoveryDialog.RecoveryDialog
	fileCompletionProvider  dialog.CompletionProvider
	commandCompletionProvider dialog.CompletionProvider
}

type ChatKeyMap struct {
	NewSession key.Binding
	Cancel     key.Binding
}

var keyMap = ChatKeyMap{
	NewSession: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "new session"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

// GetCurrentSessionID returns the current session ID
func (p *chatPage) GetCurrentSessionID() string {
	return p.session.ID
}

func (p *chatPage) Init() tea.Cmd {
	cmds := []tea.Cmd{
		p.layout.Init(),
		p.completionDialog.Init(),
		p.recoveryDialog.Init(),
	}
	return tea.Batch(cmds...)
}

func (p *chatPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Update recovery dialog first (highest priority)
	recoveryUpdated, recoveryCmd := p.recoveryDialog.Update(msg)
	p.recoveryDialog = recoveryUpdated.(*recoveryDialog.RecoveryDialog)
	if recoveryCmd != nil {
		cmds = append(cmds, recoveryCmd)
	}

	// Block most input during recovery
	if p.app.Recovery.IsRecovering() && p.recoveryDialog.IsVisible() {
		switch msg.(type) {
		case tea.KeyMsg:
			// Only allow recovery dialog keys and window resize
			return p, tea.Batch(cmds...)
		case tea.WindowSizeMsg:
			// Allow window resize events
		default:
			// Block other input during recovery
			return p, tea.Batch(cmds...)
		}
	}

	// Route messages to the layout first, which handles the editor.
	// This ensures the editor's value is up-to-date for our logic below.
	updatedLayout, cmd := p.layout.Update(msg)
	p.layout = updatedLayout.(layout.SplitPaneLayout)
	cmds = append(cmds, cmd)

	// Check the editor's value to determine completion state.
	editorValue := p.editor.GetValue()
	if strings.HasPrefix(editorValue, "/") {
		// Only set provider if dialog is not already showing or provider needs to change
		if !p.showCompletionDialog || p.completionDialog.GetId() != "slash-commands" {
			p.completionDialog.SetProvider(p.commandCompletionProvider)
		}
		p.showCompletionDialog = true
	} else if strings.HasPrefix(editorValue, "@") {
		// Only set provider if dialog is not already showing or provider needs to change
		if !p.showCompletionDialog || p.completionDialog.GetId() != "files" {
			p.completionDialog.SetProvider(p.fileCompletionProvider)
		}
		p.showCompletionDialog = true
	} else {
		p.showCompletionDialog = false
	}

	// Now, handle all other message types.
	switch msg := msg.(type) {
	case dialog.CompletionDialogCloseMsg:
		p.showCompletionDialog = false
	case dialog.SlashCommandCompleteMsg:
		// Forward to editor to update text
		updated, cmd := p.editor.Update(msg)
		p.editor = updated.(*chat.EditorCmp)
		cmds = append(cmds, cmd)
		
		// Keep dialog open if needed
		if !msg.KeepOpen {
			p.showCompletionDialog = false
		}
	case dialog.SlashCommandExecuteMsg:
		// Close completion dialog
		p.showCompletionDialog = false
		
		// Create parser and execute command
		parser := commands.NewCommandParserWithApp(commands.GetGlobalRegistry(), p.app)
		cmd := dialog.ExecuteSlashCommand(parser, msg.Raw)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case chat.SendMsg:
		// When a message is sent, clear the completion dialog.
		p.showCompletionDialog = false
		cmd := p.sendMessage(msg.Text, msg.Attachments)
		if cmd != nil {
			return p, cmd
		}
	case dialog.CommandRunCustomMsg:
		if p.app.CoderAgent.IsBusy() {
			return p, util.ReportWarn("Agent is busy, please wait before executing a command...")
		}
		content := msg.Content
		if msg.Args != nil {
			for name, value := range msg.Args {
				placeholder := "$" + name
				content = strings.ReplaceAll(content, placeholder, value)
			}
		}
		cmd := p.sendMessage(content, nil)
		if cmd != nil {
			return p, cmd
		}
	case dialog.SessionListRequestedMsg:
		// Show the session dialog (same as Ctrl+S)
		return p, func() tea.Msg {
			return tea.KeyMsg{
				Type: tea.KeyRunes,
				Runes: []rune{19}, // Ctrl+S ASCII code
			}
		}
	case dialog.SessionClearRequestedMsg:
		// Clear the current session (same as Ctrl+N)
		p.session = session.Session{}
		return p, tea.Batch(
			p.clearSidebar(),
			util.CmdHandler(chat.SessionClearedMsg{}),
		)
	case dialog.SessionCompactRequestedMsg:
		// Trigger session compacting through the proper channel
		// Send a message upward to trigger the compact process
		return p, func() tea.Msg {
			return CompactSessionRequestMsg{}
		}
	case dialog.ProjectInitRequestedMsg:
		// Send the init prompt to create CLAUDE.md
		prompt := `Please analyze this codebase and create a CLAUDE.md file containing:
1. Build/lint/test commands - especially for running a single test
2. Code style guidelines including imports, formatting, types, naming conventions, error handling, etc.

The file you create will be given to agentic coding agents (such as yourself) that operate in this repository. Make it about 20 lines long.
If there's already a CLAUDE.md, improve it.
If there are Cursor rules (in .cursor/rules/ or .cursorrules) or Copilot rules (in .github/copilot-instructions.md), make sure to include them.`
		cmd := p.sendMessage(prompt, nil)
		if cmd != nil {
			return p, cmd
		}
	case dialog.SessionNewRequestedMsg:
		// Create a new session with optional name
		sessionName := msg.Name
		if sessionName == "" {
			sessionName = "New Session"
		}
		newSession, err := p.app.Sessions.Create(context.Background(), sessionName)
		if err != nil {
			return p, util.ReportError(err)
		}
		p.session = newSession
		return p, tea.Batch(
			p.setSidebar(),
			util.CmdHandler(chat.SessionSelectedMsg(newSession)),
			util.ReportInfo(fmt.Sprintf("Created new session: %s", sessionName)),
		)
	case dialog.ConfigShowRequestedMsg:
		// Show current configuration
		cfg := config.Get()
		if cfg == nil {
			return p, util.ReportError(fmt.Errorf("unable to load configuration"))
		}
		
		// Format configuration info
		var configInfo strings.Builder
		configInfo.WriteString("Current Configuration:\n\n")
		
		// Show agent models
		configInfo.WriteString("Agent Models:\n")
		for agentName, agent := range cfg.Agents {
			model, ok := models.SupportedModels[agent.Model]
			if ok {
				configInfo.WriteString(fmt.Sprintf("  %s: %s (%s)\n", agentName, model.Name, model.ID))
			} else {
				configInfo.WriteString(fmt.Sprintf("  %s: %s\n", agentName, agent.Model))
			}
		}
		
		// Show enabled providers
		configInfo.WriteString("\nEnabled Providers:\n")
		for provider, providerCfg := range cfg.Providers {
			if !providerCfg.Disabled {
				configInfo.WriteString(fmt.Sprintf("  - %s\n", provider))
			}
		}
		
		// Show working directory
		configInfo.WriteString(fmt.Sprintf("\nWorking Directory: %s\n", cfg.WorkingDir))
		
		// Show context paths
		if len(cfg.ContextPaths) > 0 {
			configInfo.WriteString("\nContext Paths:\n")
			for _, path := range cfg.ContextPaths {
				configInfo.WriteString(fmt.Sprintf("  - %s\n", path))
			}
		}
		
		return p, util.ReportInfo(configInfo.String())
	case dialog.ConfigModelRequestedMsg:
		// TODO: Implement model configuration
		return p, util.ReportWarn("Model configuration not yet implemented")
	case dialog.AuthLoginRequestedMsg:
		// Handle OAuth2 login
		switch msg.Provider {
		case "gemini":
			return p, util.CmdHandler(dialog.ShowOAuth2DialogMsg{
				Provider: dialog.OAuth2ProviderGemini,
			})
		default:
			return p, util.ReportWarn(fmt.Sprintf("Unknown provider: %s", msg.Provider))
		}
	case dialog.AuthLogoutRequestedMsg:
		// TODO: Implement auth logout
		return p, util.ReportWarn("Auth logout not yet implemented")
	case dialog.AuthStatusRequestedMsg:
		// TODO: Implement auth status
		return p, util.ReportWarn("Auth status not yet implemented")
	case dialog.HelpRequestedMsg:
		// Show help (toggle help dialog)
		return p, func() tea.Msg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}} }
	case dialog.SessionSelectedMsg:
		// Convert dialog.SessionSelectedMsg to chat.SessionSelectedMsg
		return p, func() tea.Msg {
			return chat.SessionSelectedMsg(msg.Session)
		}
	case chat.SessionSelectedMsg:
		if p.session.ID == "" {
			cmd := p.setSidebar()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		p.session = msg
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keyMap.NewSession):
			p.session = session.Session{}
			return p, tea.Batch(
				p.clearSidebar(),
				util.CmdHandler(chat.SessionClearedMsg{}),
			)
		case key.Matches(msg, keyMap.Cancel):
			if p.showCompletionDialog {
				p.showCompletionDialog = false
			} else if p.session.ID != "" {
				p.app.CoderAgent.Cancel(p.session.ID)
			}
			return p, nil
		}
	}

	// Update the completion dialog if it's visible
	if p.showCompletionDialog {
		context, contextCmd := p.completionDialog.Update(msg)
		p.completionDialog = context.(dialog.CompletionDialog)
		cmds = append(cmds, contextCmd)
	}

	return p, tea.Batch(cmds...)
}

func (p *chatPage) setSidebar() tea.Cmd {
	sidebarContainer := layout.NewContainer(
		chat.NewSidebarCmp(p.session, p.app.History),
		layout.WithPadding(1, 1, 1, 1),
	)
	return tea.Batch(p.layout.SetRightPanel(sidebarContainer), sidebarContainer.Init())
}

func (p *chatPage) clearSidebar() tea.Cmd {
	return p.layout.ClearRightPanel()
}

func (p *chatPage) sendMessage(text string, attachments []message.Attachment) tea.Cmd {
	var cmds []tea.Cmd
	if p.session.ID == "" {
		session, err := p.app.Sessions.Create(context.Background(), "New Session")
		if err != nil {
			return util.ReportError(err)
		}

		p.session = session
		cmd := p.setSidebar()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, util.CmdHandler(chat.SessionSelectedMsg(session)))
	}

	_, err := p.app.CoderAgent.Run(context.Background(), p.session.ID, text, attachments...)
	if err != nil {
		return util.ReportError(err)
	}
	return tea.Batch(cmds...)
}

func (p *chatPage) SetSize(width, height int) tea.Cmd {
	return p.layout.SetSize(width, height)
}

func (p *chatPage) GetSize() (int, int) {
	return p.layout.GetSize()
}

func (p *chatPage) View() string {
	layoutView := p.layout.View()

	if p.showCompletionDialog {
		_, layoutHeight := p.layout.GetSize()
		editorWidth, editorHeight := p.editor.GetSize()

		p.completionDialog.SetWidth(editorWidth)
		overlay := p.completionDialog.View()

		layoutView = layout.PlaceOverlay(
			0,
			layoutHeight-editorHeight-lipgloss.Height(overlay),
			overlay,
			layoutView,
			false,
		)
	}

	// Recovery dialog has highest priority and overlays everything
	if p.recoveryDialog.IsVisible() {
		width, height := p.layout.GetSize()
		p.recoveryDialog.SetSize(width, height)
		recoveryOverlay := p.recoveryDialog.View()
		if recoveryOverlay != "" {
			layoutView = recoveryOverlay
		}
	}

	return layoutView
}

func (p *chatPage) BindingKeys() []key.Binding {
	bindings := layout.KeyMapToSlice(keyMap)
	bindings = append(bindings, p.messages.BindingKeys()...)
	bindings = append(bindings, p.editor.BindingKeys()...)
	return bindings
}

func NewChatPage(app *app.App) tea.Model {
	fileCompletionProvider := completions.NewFileAndFolderContextGroup()
	commandCompletionProvider := completions.NewSlashCommandProviderWithApp(app)

	completionDialog := dialog.NewCompletionDialogCmp(fileCompletionProvider)

	messagesContainer := layout.NewContainer(
		chat.NewMessagesCmp(app),
		layout.WithPadding(1, 1, 0, 1),
	)

	editorCmp := chat.NewEditorCmp(app)

	editorContainer := layout.NewContainer(
		editorCmp,
		layout.WithBorder(true, false, false, false),
	)
	return &chatPage{
		app:                     app,
		editor:                  editorCmp.(*chat.EditorCmp),
		messages:                messagesContainer,
		completionDialog:        completionDialog,
		recoveryDialog:          recoveryDialog.NewRecoveryDialog(),
		fileCompletionProvider:  fileCompletionProvider,
		commandCompletionProvider: commandCompletionProvider,
		layout: layout.NewSplitPane(
			layout.WithLeftPanel(messagesContainer),
			layout.WithBottomPanel(editorContainer),
		),
	}
}

