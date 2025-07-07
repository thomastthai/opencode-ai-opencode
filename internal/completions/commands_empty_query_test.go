package completions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandCompletionProvider_EmptyQuery(t *testing.T) {
	provider := NewCommandCompletionProvider()
	
	// Test with empty query (what happens when user has just "/" in editor)
	items, err := provider.GetChildEntries("")
	assert.NoError(t, err)
	
	// Should return all top-level commands
	assert.Greater(t, len(items), 0, "Empty query should return all available commands")
	
	// Verify we get actual commands
	for _, item := range items {
		assert.NotEmpty(t, item.GetValue(), "Command value should not be empty")
		assert.True(t, len(item.GetValue()) > 1, "Command should be more than just /")
		assert.Equal(t, byte('/'), item.GetValue()[0], "Command should start with /")
	}
}

func TestCommandCompletionProvider_PartialQuery(t *testing.T) {
	provider := NewCommandCompletionProvider()
	
	// Test with "lis" query (user typed /lis)
	items, err := provider.GetChildEntries("lis")
	assert.NoError(t, err)
	
	// Should return commands starting with "lis"
	for _, item := range items {
		assert.Contains(t, item.GetValue(), "lis", "Filtered commands should contain 'lis'")
	}
}