package page

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opencode-ai/opencode/internal/app"
	"github.com/opencode-ai/opencode/internal/completions"
	"github.com/opencode-ai/opencode/internal/message"
	"github.com/opencode-ai/opencode/internal/session"
	"github.com/opencode-ai/opencode/internal/tui/components/chat"
	"github.com/opencode-ai/opencode/internal/tui/components/dialog"
	"github.com/opencode-ai/opencode/internal/tui/layout"
	"github.com/opencode-ai/opencode/internal/tui/util"
)

var ChatPage PageID = "chat"

type chatPage struct {
	app                     *app.App
	editor               *chat.EditorCmp
	messages                layout.Container
	layout                  layout.SplitPaneLayout
	session                 session.Session
	completionDialog        dialog.CompletionDialog
	showCompletionDialog    bool
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

func (p *chatPage) Init() tea.Cmd {
	cmds := []tea.Cmd{
		p.layout.Init(),
		p.completionDialog.Init(),
	}
	return tea.Batch(cmds...)
}

func (p *chatPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Route messages to the layout first, which handles the editor.
	// This ensures the editor's value is up-to-date for our logic below.
	updatedLayout, cmd := p.layout.Update(msg)
	p.layout = updatedLayout.(layout.SplitPaneLayout)
	cmds = append(cmds, cmd)

	// Check the editor's value to determine completion state.
	editorValue := p.editor.GetValue()
	if strings.HasPrefix(editorValue, "/") {
		p.completionDialog.SetProvider(p.commandCompletionProvider)
		p.showCompletionDialog = true
	} else if strings.HasPrefix(editorValue, "@") {
		p.completionDialog.SetProvider(p.fileCompletionProvider)
		p.showCompletionDialog = true
	} else {
		p.showCompletionDialog = false
	}

	// Now, handle all other message types.
	switch msg := msg.(type) {
	case dialog.CompletionDialogCloseMsg:
		p.showCompletionDialog = false
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
	} else {
		u, cmd := p.layout.Update(msg)
		p.layout = u.(layout.SplitPaneLayout)
		cmds = append(cmds, cmd)
	}

	// Always pass key messages to the layout to update the editor
	if _, ok := msg.(tea.KeyMsg); ok {
		u, cmd := p.layout.Update(msg)
		p.layout = u.(layout.SplitPaneLayout)
		cmds = append(cmds, cmd)
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
	commandCompletionProvider := completions.NewCommandCompletionProvider()

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
		fileCompletionProvider:  fileCompletionProvider,
		commandCompletionProvider: commandCompletionProvider,
		layout: layout.NewSplitPane(
			layout.WithLeftPanel(messagesContainer),
			layout.WithBottomPanel(editorContainer),
		),
	}
}

