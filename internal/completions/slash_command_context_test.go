package completions

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opencode-ai/opencode/internal/app"
)

func TestSlashCommandProvider_ContextAware(t *testing.T) {
	t.Run("with app context", func(t *testing.T) {
		// Create app instance
		testApp := &app.App{}
		
		// Create provider with app context
		provider := NewSlashCommandProviderWithApp(testApp)
		
		// Verify provider is created
		assert.NotNil(t, provider)
		assert.Equal(t, "slash-commands", provider.GetId())
		
		// Test auth login completions
		items, err := provider.GetChildEntries("/auth login ")
		assert.NoError(t, err)
		
		// Should have provider completions
		assert.Greater(t, len(items), 0, "Should have auth provider completions")
		
		// Check that we get expected providers
		values := make(map[string]bool)
		for _, item := range items {
			values[item.GetValue()] = true
		}
		
		assert.True(t, values["gemini"], "Should have gemini provider")
		assert.True(t, values["anthropic"], "Should have anthropic provider")
		assert.True(t, values["openai"], "Should have openai provider")
	})
	
	t.Run("without app context", func(t *testing.T) {
		// Create provider without app context
		provider := NewSlashCommandProvider()
		
		// Test auth login completions
		items, err := provider.GetChildEntries("/auth login ")
		assert.NoError(t, err)
		
		// Should have option completions (--force, --no-browser) even without app context
		assert.Greater(t, len(items), 0, "Should have option completions")
		
		// Verify we have option completions
		hasOptions := false
		for _, item := range items {
			if strings.HasPrefix(item.GetValue(), "--") {
				hasOptions = true
				break
			}
		}
		assert.True(t, hasOptions, "Should have option completions")
	})
	
	t.Run("config model completions", func(t *testing.T) {
		// Create app instance
		testApp := &app.App{}
		
		// Create provider with app context
		provider := NewSlashCommandProviderWithApp(testApp)
		
		// Test config model completions
		items, err := provider.GetChildEntries("/config model ")
		assert.NoError(t, err)
		
		// Should have model completions if config is available
		// Number depends on config, but we should at least not crash
		assert.NotNil(t, items)
	})
}