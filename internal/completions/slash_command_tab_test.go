package completions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlashCommandProvider_TabCompletion(t *testing.T) {
	provider := NewSlashCommandProvider()
	slashProvider := provider.(*slashCommandProvider)

	t.Run("tab completion single match", func(t *testing.T) {
		// Type /se and press tab
		completed, options, err := slashProvider.HandleTabCompletion("/se")
		
		assert.NoError(t, err)
		assert.Equal(t, "/session ", completed)
		assert.Nil(t, options, "Should have no options for unique match")
	})

	t.Run("tab completion multiple matches", func(t *testing.T) {
		// Type /s and press tab - matches session and system
		completed, options, err := slashProvider.HandleTabCompletion("/s")
		
		assert.NoError(t, err)
		assert.Equal(t, "/s", completed, "Should return original for ambiguous match")
		assert.Greater(t, len(options), 1, "Should have multiple options")
		
		// Check that options are valid completion items
		for _, opt := range options {
			assert.NotEmpty(t, opt.GetValue())
			assert.NotEmpty(t, opt.DisplayValue())
		}
	})

	t.Run("tab completion for verbs", func(t *testing.T) {
		// Type /session n and press tab
		completed, options, err := slashProvider.HandleTabCompletion("/session n")
		
		assert.NoError(t, err)
		assert.Equal(t, "/session new ", completed)
		assert.Nil(t, options, "Should have no options for unique verb match")
	})

	t.Run("tab completion with no matches", func(t *testing.T) {
		// Type something that doesn't match
		completed, options, err := slashProvider.HandleTabCompletion("/xyz")
		
		assert.NoError(t, err)
		assert.Equal(t, "/xyz", completed, "Should return original when no matches")
		assert.Nil(t, options)
	})

	t.Run("tab completion already complete", func(t *testing.T) {
		// Type /help which is already a complete command
		completed, options, err := slashProvider.HandleTabCompletion("/help")
		
		assert.NoError(t, err)
		assert.Equal(t, "/help ", completed)
		assert.Nil(t, options)
	})
}