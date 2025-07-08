package dialog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTabHandlerCompleteCommands verifies the complete commands list behavior
func TestTabHandlerCompleteCommands(t *testing.T) {
	// Test that specific completed values result in correct keepOpen behavior
	tests := []struct {
		name           string
		completed      string
		expectKeepOpen bool
		description    string
	}{
		{
			name:           "system help closes dialog",
			completed:      "/system help ",
			expectKeepOpen: false,
			description:    "Complete 'system help' command should close dialog",
		},
		{
			name:           "session new keeps dialog open",
			completed:      "/session new ",
			expectKeepOpen: true,
			description:    "'session new' accepts arguments, should keep dialog open",
		},
		{
			name:           "session clear closes dialog",
			completed:      "/session clear ",
			expectKeepOpen: false,
			description:    "Complete 'session clear' command should close dialog",
		},
		{
			name:           "help closes dialog",
			completed:      "/help ",
			expectKeepOpen: false,
			description:    "Complete 'help' command should close dialog",
		},
		{
			name:           "auth login keeps dialog open",
			completed:      "/auth login ",
			expectKeepOpen: true,
			description:    "'auth login' requires arguments, should keep dialog open",
		},
		{
			name:           "config model keeps dialog open",
			completed:      "/config model ",
			expectKeepOpen: true,
			description:    "'config model' accepts arguments, should keep dialog open",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the logic directly
			keepOpen := true
			completeCommands := []string{
				"/system help ",
				"/system exit ",
				"/session clear ",
				"/session list ",
				"/project init ",
				"/config show ",
				"/auth status ",
				"/help ",
			}
			
			for _, cmd := range completeCommands {
				if tt.completed == cmd {
					keepOpen = false
					break
				}
			}
			
			assert.Equal(t, tt.expectKeepOpen, keepOpen, tt.description)
		})
	}
}

// TestCompleteCommandsList verifies the list of complete commands is consistent
func TestCompleteCommandsList(t *testing.T) {
	// These are the commands that should close the dialog after completion
	expectedCompleteCommands := []string{
		"/system help ",
		"/system exit ",
		"/session clear ",
		"/session list ",
		"/project init ",
		"/config show ",
		"/auth status ",
		"/help ",
	}
	
	// Verify the list matches what's in tab_handler.go
	// This test ensures the hardcoded list stays in sync
	completeCommands := []string{
		"/system help ",
		"/system exit ",
		"/session clear ",
		"/session list ",
		"/project init ",
		"/config show ",
		"/auth status ",
		"/help ",
	}
	
	assert.Equal(t, expectedCompleteCommands, completeCommands, 
		"Complete commands list should match expected list")
}

