package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opencode-ai/opencode/internal/app"
)

func TestContextAwareCompletions(t *testing.T) {
	t.Run("provider completions", func(t *testing.T) {
		testApp := &app.App{}
		
		// Test auth provider completions
		completions := GetDynamicCompletions("auth", "login", []string{}, testApp)
		
		assert.Len(t, completions, 3, "Should have 3 provider completions")
		
		// Check for expected providers
		providers := make(map[string]bool)
		for _, comp := range completions {
			providers[comp.Value] = true
		}
		
		assert.True(t, providers["gemini"])
		assert.True(t, providers["anthropic"])
		assert.True(t, providers["openai"])
	})
	
	t.Run("file completions", func(t *testing.T) {
		testApp := &app.App{}
		
		// Test file completions for project add-dir
		completions := GetDynamicCompletions("project", "add-dir", []string{}, testApp)
		
		// Should return files/directories from current directory
		assert.NotNil(t, completions)
		// Number of completions depends on current directory contents
	})
	
	t.Run("no completions for unknown commands", func(t *testing.T) {
		testApp := &app.App{}
		
		// Test unknown topic
		completions := GetDynamicCompletions("unknown", "action", []string{}, testApp)
		assert.Nil(t, completions)
		
		// Test unknown verb
		completions = GetDynamicCompletions("session", "unknown", []string{}, testApp)
		assert.Nil(t, completions)
	})
}

func TestCommandParserWithApp(t *testing.T) {
	t.Run("parser works without app", func(t *testing.T) {
		// Create parser without app
		registry := NewCommandRegistry()
		parser := NewCommandParser(registry)
		
		// Parse a command
		parsed := parser.Parse("/session switch ")
		completions := parser.GetCompletions(parsed)
		
		// Should have option completions even without app
		assert.NotNil(t, completions)
		// Check that we get at least the --verbose option from session topic
		hasVerbose := false
		for _, c := range completions {
			if c.Value == "--verbose" {
				hasVerbose = true
				break
			}
		}
		assert.True(t, hasVerbose, "Should have --verbose option")
	})
}