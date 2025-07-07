package recovery

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/opencode-ai/opencode/internal/pubsub"
	"github.com/opencode-ai/opencode/internal/recovery"
	"github.com/opencode-ai/opencode/internal/tui/layout"
	"github.com/opencode-ai/opencode/internal/tui/styles"
	"github.com/opencode-ai/opencode/internal/tui/theme"
)

type RecoveryDialog struct {
	status   recovery.RecoveryStatus
	visible  bool
	width    int
	height   int
	keyMap   RecoveryKeyMap
}

type RecoveryKeyMap struct {
	Dismiss key.Binding
}

var defaultKeyMap = RecoveryKeyMap{
	Dismiss: key.NewBinding(
		key.WithKeys("esc", "enter"),
		key.WithHelp("esc/enter", "dismiss"),
	),
}

func NewRecoveryDialog() *RecoveryDialog {
	return &RecoveryDialog{
		keyMap:  defaultKeyMap,
		visible: false,
	}
}

func (r *RecoveryDialog) Init() tea.Cmd {
	return nil
}

func (r *RecoveryDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if r.visible {
			switch {
			case key.Matches(msg, r.keyMap.Dismiss):
				if r.status.State == recovery.RecoveryCompleted || r.status.State == recovery.RecoveryFailed {
					r.visible = false
				}
				return r, nil
			}
		}
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height
	case pubsub.Event[recovery.RecoveryStatus]:
		r.status = msg.Payload
		// Show dialog when recovery starts
		if r.status.State == recovery.RecoveryInProgress {
			r.visible = true
		}
		// Auto-hide after 3 seconds if recovery completed successfully
		if r.status.State == recovery.RecoveryCompleted {
			return r, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
				return HideRecoveryMsg{}
			})
		}
	case HideRecoveryMsg:
		if r.status.State == recovery.RecoveryCompleted {
			r.visible = false
		}
	}
	
	return r, nil
}

type HideRecoveryMsg struct{}

func (r *RecoveryDialog) View() string {
	if !r.visible {
		return ""
	}

	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	// Build the content
	var content strings.Builder
	
	// Title
	title := "🔄 System Recovery"
	if r.status.State == recovery.RecoveryCompleted {
		title = "✅ Recovery Complete"
	} else if r.status.State == recovery.RecoveryFailed {
		title = "❌ Recovery Failed"
	}
	
	titleStyle := baseStyle.Bold(true).Foreground(t.Primary())
	content.WriteString(titleStyle.Render(title) + "\n\n")
	
	// Progress
	if len(r.status.Steps) > 0 {
		progress := fmt.Sprintf("Progress: %d/%d", r.status.CompletedSteps, r.status.TotalSteps)
		content.WriteString(baseStyle.Foreground(t.TextMuted()).Render(progress) + "\n\n")
		
		// Steps
		for _, step := range r.status.Steps {
			stepLine := r.renderStep(step)
			content.WriteString(stepLine + "\n")
		}
	}
	
	// Duration
	if !r.status.StartTime.IsZero() {
		var duration time.Duration
		if !r.status.EndTime.IsZero() {
			duration = r.status.EndTime.Sub(r.status.StartTime)
		} else {
			duration = time.Since(r.status.StartTime)
		}
		
		content.WriteString("\n")
		durationText := fmt.Sprintf("Duration: %v", duration.Round(time.Millisecond))
		content.WriteString(baseStyle.Foreground(t.TextMuted()).Render(durationText))
	}
	
	// Help text
	if r.status.State == recovery.RecoveryCompleted || r.status.State == recovery.RecoveryFailed {
		content.WriteString("\n\n")
		helpText := baseStyle.Foreground(t.TextMuted()).Render("Press esc or enter to dismiss")
		content.WriteString(helpText)
	}

	// Calculate dialog dimensions
	lines := strings.Split(content.String(), "\n")
	contentWidth := 0
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth > contentWidth {
			contentWidth = lineWidth
		}
	}
	
	dialogWidth := contentWidth + 4 // padding
	if dialogWidth > r.width-4 {
		dialogWidth = r.width - 4
	}
	if dialogWidth < 40 {
		dialogWidth = 40
	}

	// Style the dialog
	dialogStyle := baseStyle.
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary()).
		Background(t.Background()).
		Padding(1, 2).
		Width(dialogWidth)
		
	dialog := dialogStyle.Render(content.String())
	
	// Center the dialog
	dialogHeight := lipgloss.Height(dialog)
	x := (r.width - dialogWidth) / 2
	y := (r.height - dialogHeight) / 2
	if y < 0 {
		y = 0
	}
	
	return layout.PlaceOverlay(x, y, dialog, "", false)
}

func (r *RecoveryDialog) renderStep(step recovery.RecoveryStep) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()
	
	var icon string
	var color lipgloss.TerminalColor
	
	switch step.Status {
	case recovery.StepPending:
		icon = "⏳"
		color = t.TextMuted()
	case recovery.StepInProgress:
		icon = "🔄"
		color = t.Warning()
	case recovery.StepCompleted:
		icon = "✅"
		color = t.Success()
	case recovery.StepFailed:
		icon = "❌"
		color = t.Error()
	}
	
	stepText := fmt.Sprintf("%s %s", icon, step.Description)
	if step.Error != nil {
		stepText += fmt.Sprintf(" (%v)", step.Error)
	}
	
	return baseStyle.Foreground(color).Render(stepText)
}

func (r *RecoveryDialog) SetSize(width, height int) {
	r.width = width
	r.height = height
}

func (r *RecoveryDialog) IsVisible() bool {
	return r.visible
}

func (r *RecoveryDialog) BindingKeys() []key.Binding {
	if r.visible {
		return layout.KeyMapToSlice(r.keyMap)
	}
	return []key.Binding{}
}