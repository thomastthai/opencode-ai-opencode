package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTabCompletionDoubleSlash(t *testing.T) {
	// Create parser with test registry
	registry := NewCommandRegistry()
	parser := NewCommandParser(registry)
	
	// Initialize hierarchical commands
	InitializeBuiltinCommands(parser.hierarchicalReg)
	
	tests := []struct {
		name           string
		input          string
		expectedPrefix string
		hasCompletions bool
		description    string
	}{
		{
			name:           "slash with partial topic",
			input:          "/se",
			expectedPrefix: "/session",
			hasCompletions: false, // Single match
			description:    "Should complete /se to /session without double slash",
		},
		{
			name:           "no slash with partial topic",
			input:          "se",  
			expectedPrefix: "se",  // Should return input as-is when no match
			hasCompletions: true,
			description:    "Should not add slash when input doesn't have one",
		},
		{
			name:           "slash with full topic",
			input:          "/session",
			expectedPrefix: "/session",
			hasCompletions: true,
			description:    "Should return input as-is with available commands",
		},
		{
			name:           "slash with ambiguous prefix",
			input:          "/s",
			expectedPrefix: "/s",
			hasCompletions: true,
			description:    "Should return input with completions when ambiguous",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, completions := parser.GetTabCompletion(tt.input)
			
			assert.Equal(t, tt.expectedPrefix, result, tt.description)
			
			if tt.hasCompletions {
				assert.NotNil(t, completions, "Should have completions")
				assert.Greater(t, len(completions), 0, "Should have at least one completion")
			} else {
				assert.Nil(t, completions, "Should not have completions for single match")
			}
			
			// Ensure no double slash in result
			assert.NotContains(t, result, "//", "Result should never contain double slash")
		})
	}
}

func TestGetTabCompletionSequence(t *testing.T) {
	// Test the specific sequence that causes the bug:
	// 1. Type /session<tab>
	// 2. Backspace to erase all
	// 3. Type /se<tab>
	
	registry := NewCommandRegistry()
	parser := NewCommandParser(registry)
	InitializeBuiltinCommands(parser.hierarchicalReg)
	
	// Step 1: Complete /session
	result1, _ := parser.GetTabCompletion("/session")
	assert.Equal(t, "/session", result1)
	assert.NotContains(t, result1, "//")
	
	// Step 2: User backspaces everything (simulated by empty string)
	// This is just to show the state
	
	// Step 3: Type /se and tab complete
	result3, _ := parser.GetTabCompletion("/se")
	assert.Equal(t, "/session ", result3) // Should complete to "/session " not "//session"
	assert.NotContains(t, result3, "//", "Should not have double slash after backspace and retype")
}