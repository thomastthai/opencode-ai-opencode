package page

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/opencode-ai/opencode/internal/app"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/tui/components/dialog"
	"github.com/stretchr/testify/assert"
)

// createTestApp creates a minimal app for testing
func createTestApp(t *testing.T) *app.App {
	// Create a temporary directory and a minimal config file
	tmpDir, err := os.MkdirTemp("", "opencode-test-")
	assert.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	configContent := `{
		"agents": {
			"coder": { "model": "test-model" }
		},
		"mcpServers": {}
	}`
	configPath := filepath.Join(tmpDir, ".opencode.json")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Load the configuration
	_, err = config.Load(tmpDir, false)
	assert.NoError(t, err)

	// Create the app
	db, _ := sql.Open("sqlite3", ":memory:")
	app, err := app.New(context.Background(), db, true)
	assert.NoError(t, err)
	
	return app
}

// TestCompleteCommandsCloseDialog verifies that commands without arguments close the dialog after tab completion
func TestCompleteCommandsCloseDialog(t *testing.T) {
	tests := []struct {
		name           string
		editorValue    string
		expectDialogOpen bool
		description    string
	}{
		{
			name:           "system help with space closes dialog",
			editorValue:    "/system help ",
			expectDialogOpen: false,
			description:    "Complete command without arguments should close dialog",
		},
		{
			name:           "system exit with space closes dialog",
			editorValue:    "/system exit ",
			expectDialogOpen: false,
			description:    "Complete command without arguments should close dialog",
		},
		{
			name:           "session clear with space closes dialog",
			editorValue:    "/session clear ",
			expectDialogOpen: false,
			description:    "Complete command without arguments should close dialog",
		},
		{
			name:           "session list with space closes dialog",
			editorValue:    "/session list ",
			expectDialogOpen: false,
			description:    "Complete command without arguments should close dialog",
		},
		{
			name:           "help with space closes dialog",
			editorValue:    "/help ",
			expectDialogOpen: false,
			description:    "Complete command without arguments should close dialog",
		},
		{
			name:           "session new with space keeps dialog open",
			editorValue:    "/session new ",
			expectDialogOpen: true,
			description:    "Command that accepts arguments should keep dialog open",
		},
		{
			name:           "auth login with space keeps dialog open",
			editorValue:    "/auth login ",
			expectDialogOpen: true,
			description:    "Command that requires arguments should keep dialog open",
		},
		{
			name:           "config model with space keeps dialog open",
			editorValue:    "/config model ",
			expectDialogOpen: true,
			description:    "Command that accepts arguments should keep dialog open",
		},
		{
			name:           "incomplete command keeps dialog open",
			editorValue:    "/system h",
			expectDialogOpen: true,
			description:    "Incomplete command should keep dialog open for completion",
		},
		{
			name:           "just slash keeps dialog open",
			editorValue:    "/",
			expectDialogOpen: true,
			description:    "Just slash should show command list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal test app with required components
			testApp := createTestApp(t)
			
			// Create chat page
			page := NewChatPage(testApp).(*chatPage)
			
			// Initialize the page
			page.Init()
			
			// Set editor value
			page.editor.SetValue(tt.editorValue)
			
			// Trigger an update to process the editor value
			_, _ = page.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			
			// Check dialog state
			assert.Equal(t, tt.expectDialogOpen, page.showCompletionDialog, tt.description)
		})
	}
}

// TestTabCompletionForCompleteCommands verifies tab completion behavior for complete commands
func TestTabCompletionForCompleteCommands(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedCompleted string
		expectedKeepOpen  bool
		description       string
	}{
		{
			name:              "system help tab completion",
			input:             "/system h",
			expectedCompleted: "/system help ",
			expectedKeepOpen:  false,
			description:       "Tab completing to a complete command should close dialog",
		},
		{
			name:              "session new tab completion",
			input:             "/session n",
			expectedCompleted: "/session new ",
			expectedKeepOpen:  true,
			description:       "Tab completing to command with arguments should keep dialog open",
		},
		{
			name:              "auth login tab completion",
			input:             "/auth l",
			expectedCompleted: "/auth login ",
			expectedKeepOpen:  true,
			description:       "Tab completing to command requiring arguments should keep dialog open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal test app with required components
			testApp := createTestApp(t)
			
			// Create chat page
			page := NewChatPage(testApp).(*chatPage)
			
			// Initialize the page
			page.Init()
			
			// Set editor value to trigger dialog
			page.editor.SetValue(tt.input)
			_, _ = page.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			
			// Simulate tab completion
			msg := dialog.SlashCommandCompleteMsg{
				OriginalValue: tt.input,
				NewValue:      tt.expectedCompleted,
				CursorPos:     len(tt.expectedCompleted),
				KeepOpen:      tt.expectedKeepOpen,
			}
			
			_, _ = page.Update(msg)
			
			// Verify editor was updated
			assert.Equal(t, tt.expectedCompleted, page.editor.GetValue())
			
			// Verify dialog state matches expected
			assert.Equal(t, tt.expectedKeepOpen, page.showCompletionDialog, tt.description)
		})
	}
}

// TestHelpCommandExecution verifies help command shows output
func TestHelpCommandExecution(t *testing.T) {
	// Create test app
	testApp := createTestApp(t)
	
	// Create chat page
	page := NewChatPage(testApp).(*chatPage)
	
	// Initialize the page
	page.Init()
	
	// Test help requested message
	helpMsg := dialog.HelpRequestedMsg{Topic: ""}
	model, cmd := page.Update(helpMsg)
	
	// Verify model is returned
	assert.NotNil(t, model)
	
	// Verify command is returned (ReportInfo with help text)
	assert.NotNil(t, cmd, "Help command should return a command to display help text")
	
	// Execute the command to get the message
	if cmd != nil {
		msg := cmd()
		// The ReportInfo command should produce a message
		assert.NotNil(t, msg, "Help command should produce a message")
	}
}

// TestRaceConditionPrevention verifies that rapid tab+enter doesn't execute incomplete commands
func TestRaceConditionPrevention(t *testing.T) {
	// Create test app
	testApp := createTestApp(t)
	
	// Create chat page
	page := NewChatPage(testApp).(*chatPage)
	
	// Initialize the page
	page.Init()
	
	// Set initial command
	page.editor.SetValue("/system h")
	_, _ = page.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	
	// Start tab completion
	tabStartMsg := dialog.TabCompletionStartedMsg{
		CurrentValue: "/system h",
		TargetValue:  "/system help ",
	}
	_, _ = page.Update(tabStartMsg)
	
	// Verify completion is pending
	assert.Equal(t, CompletionStatePending, page.completionState)
	assert.Equal(t, "/system help ", page.pendingCompletion)
	
	// Try to execute command while completion is pending
	executeMsg := dialog.SlashCommandExecuteMsg{
		Raw: "/system h",
	}
	_, cmd := page.Update(executeMsg)
	
	// Command should be blocked (no command returned)
	assert.Nil(t, cmd, "Command execution should be blocked during pending completion")
	
	// Complete the tab completion
	completeMsg := dialog.SlashCommandCompleteMsg{
		OriginalValue: "/system h",
		NewValue:      "/system help ",
		CursorPos:     len("/system help "),
		KeepOpen:      false,
	}
	_, _ = page.Update(completeMsg)
	
	// Verify completion state is reset
	assert.Equal(t, CompletionStateIdle, page.completionState)
	assert.Equal(t, "", page.pendingCompletion)
	
	// Now execution should work
	executeMsg2 := dialog.SlashCommandExecuteMsg{
		Raw: "/system help",
	}
	_, cmd2 := page.Update(executeMsg2)
	
	// Command should execute
	assert.NotNil(t, cmd2, "Command should execute after completion is done")
}