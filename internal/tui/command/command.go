package command

import tea "github.com/charmbracelet/bubbletea"

// Command represents a command that can be executed
type Command struct {
	ID          string
	Title       string
	Description string
	Handler     func(cmd Command) tea.Cmd
}
