package dialog

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/tui/layout"
	"github.com/opencode-ai/opencode/internal/tui/styles"
	"github.com/opencode-ai/opencode/internal/tui/theme"
	"github.com/opencode-ai/opencode/internal/tui/util"
)

// OAuth2State represents the current state of OAuth2 authentication
type OAuth2State string

const (
	OAuth2StateIdle        OAuth2State = "idle"
	OAuth2StateAuthenticating OAuth2State = "authenticating"
	OAuth2StateSuccess     OAuth2State = "success"
	OAuth2StateError       OAuth2State = "error"
)

// OAuth2Provider represents the OAuth2 provider type
type OAuth2Provider string

const (
	OAuth2ProviderGemini OAuth2Provider = "gemini"
)

// OAuth2DialogMsg represents messages for OAuth2 dialog
type OAuth2DialogMsg struct {
	Provider OAuth2Provider
	State    OAuth2State
	Message  string
	Error    error
}

// ShowOAuth2DialogMsg represents a message to show the OAuth2 dialog
type ShowOAuth2DialogMsg struct {
	Provider OAuth2Provider
}

// OAuth2DialogCmp interface for OAuth2 dialog component
type OAuth2DialogCmp interface {
	tea.Model
	layout.Bindings
	StartOAuth2(provider OAuth2Provider) tea.Cmd
	SetState(state OAuth2State, message string, err error) tea.Cmd
}

type oauthMapping struct {
	Close key.Binding
	Retry key.Binding
}

var oauthKeys = oauthMapping{
	Close: key.NewBinding(
		key.WithKeys("esc", "enter"),
		key.WithHelp("esc/enter", "close"),
	),
	Retry: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "retry"),
	),
}

// oauthDialogCmp is the implementation of OAuth2DialogCmp
type oauthDialogCmp struct {
	width      int
	height     int
	windowSize tea.WindowSizeMsg
	
	state    OAuth2State
	provider OAuth2Provider
	message  string
	error    error
	
	spinner spinner.Model
}

func (o *oauthDialogCmp) Init() tea.Cmd {
	return o.spinner.Tick
}

func (o *oauthDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		o.windowSize = msg
		o.SetSize()
		
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, oauthKeys.Close):
			if o.state != OAuth2StateAuthenticating {
				return o, util.CmdHandler("close_oauth2_dialog")
			}
			
		case key.Matches(msg, oauthKeys.Retry):
			if o.state == OAuth2StateError {
				return o, o.StartOAuth2(o.provider)
			}
		}
		
	case OAuth2DialogMsg:
		o.state = msg.State
		o.message = msg.Message
		o.error = msg.Error
		o.provider = msg.Provider
		
	case spinner.TickMsg:
		if o.state == OAuth2StateAuthenticating {
			var cmd tea.Cmd
			o.spinner, cmd = o.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	
	return o, tea.Batch(cmds...)
}

func (o *oauthDialogCmp) StartOAuth2(provider OAuth2Provider) tea.Cmd {
	o.state = OAuth2StateAuthenticating
	o.provider = provider
	o.message = "Starting OAuth2 authentication..."
	o.error = nil
	
	return tea.Batch(
		o.spinner.Tick,
		func() tea.Msg {
			ctx := context.Background()
			var err error
			
			switch provider {
			case OAuth2ProviderGemini:
				err = config.LoginWithGeminiOAuth2(ctx)
			default:
				err = fmt.Errorf("unsupported OAuth2 provider: %s", provider)
			}
			
			if err != nil {
				return OAuth2DialogMsg{
					Provider: provider,
					State:    OAuth2StateError,
					Message:  "OAuth2 authentication failed",
					Error:    err,
				}
			}
			
			return OAuth2DialogMsg{
				Provider: provider,
				State:    OAuth2StateSuccess,
				Message:  "OAuth2 authentication successful!",
				Error:    nil,
			}
		},
	)
}

func (o *oauthDialogCmp) SetState(state OAuth2State, message string, err error) tea.Cmd {
	return func() tea.Msg {
		return OAuth2DialogMsg{
			Provider: o.provider,
			State:    state,
			Message:  message,
			Error:    err,
		}
	}
}

func (o *oauthDialogCmp) renderContent() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()
	
	var content []string
	
	// Provider header
	providerName := strings.Title(string(o.provider))
	header := baseStyle.
		Bold(true).
		Foreground(t.Primary()).
		Render(fmt.Sprintf("%s OAuth2 Authentication", providerName))
	content = append(content, header)
	content = append(content, baseStyle.Render(""))
	
	// State-specific content
	switch o.state {
	case OAuth2StateIdle:
		content = append(content, baseStyle.Render("Ready to authenticate"))
		
	case OAuth2StateAuthenticating:
		spinnerStr := o.spinner.View()
		statusLine := baseStyle.Render(fmt.Sprintf("%s %s", spinnerStr, o.message))
		content = append(content, statusLine)
		content = append(content, baseStyle.Render(""))
		
		// Add instruction for browser
		instruction := baseStyle.
			Foreground(t.TextMuted()).
			Render("Your browser should open automatically.")
		content = append(content, instruction)
		
		instruction2 := baseStyle.
			Foreground(t.TextMuted()).
			Render("If it doesn't, check the console for the authentication URL.")
		content = append(content, instruction2)
		
	case OAuth2StateSuccess:
		successIcon := baseStyle.
			Foreground(t.Success()).
			Bold(true).
			Render("✓")
		successMsg := baseStyle.
			Foreground(t.Success()).
			Render(o.message)
		content = append(content, fmt.Sprintf("%s %s", successIcon, successMsg))
		content = append(content, baseStyle.Render(""))
		
		infoMsg := baseStyle.
			Foreground(t.TextMuted()).
			Render("You can now use "+providerName+" models in OpenCode.")
		content = append(content, infoMsg)
		
	case OAuth2StateError:
		errorIcon := baseStyle.
			Foreground(t.Error()).
			Bold(true).
			Render("✗")
		errorMsg := baseStyle.
			Foreground(t.Error()).
			Render(o.message)
		content = append(content, fmt.Sprintf("%s %s", errorIcon, errorMsg))
		content = append(content, baseStyle.Render(""))
		
		if o.error != nil {
			errorDetail := baseStyle.
				Foreground(t.TextMuted()).
				Width(o.width - 8).
				Render(fmt.Sprintf("Error: %s", o.error.Error()))
			content = append(content, errorDetail)
			content = append(content, baseStyle.Render(""))
		}
		
		retryMsg := baseStyle.
			Foreground(t.TextMuted()).
			Render("Press 'r' to retry authentication.")
		content = append(content, retryMsg)
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (o *oauthDialogCmp) renderButtons() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()
	
	var buttons []string
	
	switch o.state {
	case OAuth2StateAuthenticating:
		// No buttons during authentication
		return ""
		
	case OAuth2StateError:
		retryBtn := baseStyle.
			Foreground(t.Primary()).
			Bold(true).
			Render("Retry (r)")
		buttons = append(buttons, retryBtn)
		
		closeBtn := baseStyle.
			Foreground(t.TextMuted()).
			Render("Close (esc)")
		buttons = append(buttons, closeBtn)
		
	default:
		closeBtn := baseStyle.
			Foreground(t.TextMuted()).
			Render("Close (esc)")
		buttons = append(buttons, closeBtn)
	}
	
	if len(buttons) == 0 {
		return ""
	}
	
	buttonRow := lipgloss.JoinHorizontal(lipgloss.Left, buttons...)
	
	// Center the buttons
	remainingWidth := o.width - 4 - lipgloss.Width(buttonRow)
	if remainingWidth > 0 {
		padding := strings.Repeat(" ", remainingWidth/2)
		buttonRow = padding + buttonRow
	}
	
	return buttonRow
}

func (o *oauthDialogCmp) render() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()
	
	// Dialog title
	title := baseStyle.
		Bold(true).
		Width(o.width - 4).
		Foreground(t.Primary()).
		Render("OAuth2 Authentication")
	
	// Main content
	content := o.renderContent()
	
	// Buttons
	buttons := o.renderButtons()
	
	// Join all parts
	var parts []string
	parts = append(parts, title)
	parts = append(parts, baseStyle.Render(""))
	parts = append(parts, content)
	
	if buttons != "" {
		parts = append(parts, baseStyle.Render(""))
		parts = append(parts, buttons)
	}
	
	dialogContent := lipgloss.JoinVertical(lipgloss.Left, parts...)
	
	// Add padding and border
	return baseStyle.
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderBackground(t.Background()).
		BorderForeground(t.TextMuted()).
		Width(o.width).
		Height(o.height).
		Render(dialogContent)
}

func (o *oauthDialogCmp) View() string {
	return o.render()
}

func (o *oauthDialogCmp) BindingKeys() []key.Binding {
	return layout.KeyMapToSlice(oauthKeys)
}

func (o *oauthDialogCmp) SetSize() {
	o.width = int(float64(o.windowSize.Width) * 0.6)
	o.height = int(float64(o.windowSize.Height) * 0.4)
	
	if o.width < 50 {
		o.width = 50
	}
	if o.height < 12 {
		o.height = 12
	}
}

func NewOAuth2DialogCmp() OAuth2DialogCmp {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	return &oauthDialogCmp{
		state:   OAuth2StateIdle,
		spinner: s,
	}
}