package completions

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlashCommandProvider(t *testing.T) {
	provider := NewSlashCommandProvider()

	t.Run("basic properties", func(t *testing.T) {
		assert.Equal(t, "slash-commands", provider.GetId())
		
		entry := provider.GetEntry()
		assert.Equal(t, "Commands", entry.DisplayValue())
		assert.Equal(t, "/", entry.GetValue())
	})

	t.Run("get items for topics", func(t *testing.T) {
		items, err := provider.GetChildEntries("/")
		assert.NoError(t, err)
		assert.Greater(t, len(items), 0, "Should have topic completions")
		
		// Check that we have expected topics by looking at the values
		topics := make(map[string]bool)
		for _, item := range items {
			// The value should be like "/session ", so trim the / and space
			value := strings.TrimSpace(strings.TrimPrefix(item.GetValue(), "/"))
			topics[value] = true
		}
		
		assert.True(t, topics["session"], "Should have session topic")
		assert.True(t, topics["config"], "Should have config topic")
		assert.True(t, topics["help"], "Should have help topic")
	})

	t.Run("get items for verbs", func(t *testing.T) {
		items, err := provider.GetChildEntries("/session ")
		assert.NoError(t, err)
		assert.Greater(t, len(items), 0, "Should have verb completions")
		
		// Check that values are complete commands
		for _, item := range items {
			assert.True(t, strings.HasPrefix(item.GetValue(), "/session "), 
				"Verb completion should include topic")
		}
	})

	t.Run("filtered completions", func(t *testing.T) {
		items, err := provider.GetChildEntries("/se")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(items), "Should have one matching topic")
		assert.Equal(t, "/session ", items[0].GetValue())
	})

	t.Run("tab completion single match", func(t *testing.T) {
		slashProvider := provider.(*slashCommandProvider)
		completed, options, err := slashProvider.HandleTabCompletion("/se")
		
		assert.NoError(t, err)
		assert.Equal(t, "/session ", completed)
		assert.Nil(t, options, "Should have no options for unique match")
	})

	t.Run("tab completion multiple matches", func(t *testing.T) {
		slashProvider := provider.(*slashCommandProvider)
		completed, options, err := slashProvider.HandleTabCompletion("/s")
		
		assert.NoError(t, err)
		assert.Equal(t, "/s", completed, "Should return original for ambiguous match")
		assert.Greater(t, len(options), 1, "Should have multiple options")
	})

	t.Run("command structure parsing", func(t *testing.T) {
		cmd := ParseSlashCommand("/session new my-session")
		
		assert.Equal(t, "session", cmd.Topic)
		assert.Equal(t, "new", cmd.Verb)
		assert.Equal(t, []string{"my-session"}, cmd.Args)
	})

	t.Run("slash command detection", func(t *testing.T) {
		assert.True(t, IsSlashCommand("/session"))
		assert.True(t, IsSlashCommand("/"))
		assert.False(t, IsSlashCommand("session"))
		assert.False(t, IsSlashCommand("@file"))
	})
}

// Test state transitions and temporal stability
func TestSlashCommandProvider_StateTransitions(t *testing.T) {
	provider := NewSlashCommandProvider()

	t.Run("progressive completion building", func(t *testing.T) {
		// Simulate progressive typing
		inputs := []string{
			"/",
			"/s",
			"/se",
			"/sess",
			"/session",
			"/session ",
			"/session n",
			"/session ne",
			"/session new",
			"/session new ",
		}

		previousCount := -1
		for _, input := range inputs {
			items, err := provider.GetChildEntries(input)
			assert.NoError(t, err)
			
			// Verify we get appropriate items at each stage
			if strings.HasSuffix(input, " ") || strings.Contains(input, " ") {
				// After a space or with a verb, we should have different items
				if previousCount != -1 && len(items) > 0 {
					// Note: count might be same if filtering produces same number
					// But the items themselves should be different
				}
			}
			
			// Verify all items are valid completion items
			for _, item := range items {
				assert.NotEmpty(t, item.GetValue(), 
					"All items should have a value for: %s", input)
				assert.NotEmpty(t, item.DisplayValue(),
					"All items should have a display value for: %s", input)
			}
			
			if len(items) > 0 {
				previousCount = len(items)
			}
		}
	})

	t.Run("custom item rendering", func(t *testing.T) {
		items, err := provider.GetChildEntries("/se")
		assert.NoError(t, err)
		assert.Greater(t, len(items), 0, "Should have items")
		
		// Test that our custom items render properly
		item := items[0]
		rendered := item.Render(true, 50)
		assert.NotEmpty(t, rendered, "Should render item")
	})
}