package dialog

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opencode-ai/opencode/internal/commands"
	"github.com/opencode-ai/opencode/internal/tui/util"
)

// SlashCommandExecuteMsg is sent when a slash command should be executed
type SlashCommandExecuteMsg struct {
	Command commands.SlashCommand
	Raw     string
}

// ExecuteSlashCommand parses and executes a slash command
func ExecuteSlashCommand(parser *commands.CommandParser, input string) tea.Cmd {
	// Parse the command
	cmd := parser.Parse(input)
	
	// Check if it can be executed
	if !parser.CanExecute(cmd) {
		return util.ReportWarn(fmt.Sprintf("Invalid command: %s", input))
	}
	
	// Get the hierarchical registry
	registry := parser.GetHierarchicalRegistry()
	
	// Execute the command
	err := registry.Execute(context.Background(), cmd)
	
	// Handle special error messages that indicate command requests
	if err != nil {
		errMsg := err.Error()
		
		// Map error messages to appropriate TUI messages
		switch {
		case strings.Contains(errMsg, "session_new_requested"):
			return util.CmdHandler(SessionNewRequestedMsg{
				Name: getArgOrDefault(cmd.Args, 0, ""),
			})
			
		case strings.Contains(errMsg, "session_list_requested"):
			verbose := false
			if cmd.Options != nil {
				verbose = cmd.Options.GetBool("verbose")
			}
			return util.CmdHandler(SessionListRequestedMsg{
				Verbose: verbose,
			})
			
		case strings.Contains(errMsg, "clear_session_requested"):
			return util.CmdHandler(SessionClearRequestedMsg{})
			
		case strings.Contains(errMsg, "compact_session_requested"):
			instructions := ""
			if len(cmd.Args) > 0 {
				instructions = strings.Join(cmd.Args, " ")
			}
			return util.CmdHandler(SessionCompactRequestedMsg{
				Instructions: instructions,
			})
			
		case strings.Contains(errMsg, "init_project_requested"):
			return util.CmdHandler(ProjectInitRequestedMsg{})
			
		case strings.Contains(errMsg, "auth_login_requested"):
			provider := getArgOrDefault(cmd.Args, 0, "")
			if provider == "" {
				return util.ReportWarn("Please specify a provider (e.g., /auth login gemini)")
			}
			return util.CmdHandler(AuthLoginRequestedMsg{
				Provider: provider,
			})
			
		case strings.Contains(errMsg, "auth_logout_requested"):
			return util.CmdHandler(AuthLogoutRequestedMsg{
				Provider: getArgOrDefault(cmd.Args, 0, ""),
			})
			
		case strings.Contains(errMsg, "auth_status_requested"):
			return util.CmdHandler(AuthStatusRequestedMsg{})
			
		case strings.Contains(errMsg, "config_show_requested"):
			return util.CmdHandler(ConfigShowRequestedMsg{})
			
		case strings.Contains(errMsg, "config_model_requested"):
			return util.CmdHandler(ConfigModelRequestedMsg{
				Model: getArgOrDefault(cmd.Args, 0, ""),
			})
			
		case strings.Contains(errMsg, "help_requested"):
			return util.CmdHandler(HelpRequestedMsg{
				Topic: getArgOrDefault(cmd.Args, 0, ""),
			})
			
		case strings.Contains(errMsg, "exit_requested"):
			return tea.Quit
			
		default:
			// Unknown error
			return util.ReportError(err)
		}
	}
	
	return nil
}

// Helper function to get argument or default value
func getArgOrDefault(args []string, index int, defaultValue string) string {
	if index < len(args) {
		return args[index]
	}
	return defaultValue
}

// Command execution messages - these will be handled by the appropriate components

type SessionNewRequestedMsg struct {
	Name string
}

type SessionListRequestedMsg struct{
	Verbose bool
}

type SessionClearRequestedMsg struct{}

type SessionCompactRequestedMsg struct {
	Instructions string
}

type ProjectInitRequestedMsg struct{}

type AuthLoginRequestedMsg struct {
	Provider string
}

type AuthLogoutRequestedMsg struct {
	Provider string
}

type AuthStatusRequestedMsg struct{}

type ConfigShowRequestedMsg struct{}

type ConfigModelRequestedMsg struct {
	Model string
}

type HelpRequestedMsg struct {
	Topic string
}