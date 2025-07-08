package dialog

import (
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/session"
	"github.com/stretchr/testify/assert"
)

func TestSessionDialogVerboseMode(t *testing.T) {
	dialog := NewSessionDialogCmp()
	
	// Create test sessions
	now := time.Now().Unix()
	sessions := []session.Session{
		{
			ID:           "session1",
			Title:        "Test Session 1",
			MessageCount: 10,
			UpdatedAt:    now - 3600, // 1 hour ago
		},
		{
			ID:           "session2",
			Title:        "Test Session 2",
			MessageCount: 25,
			UpdatedAt:    now - 7200, // 2 hours ago
		},
	}
	
	dialog.SetSessions(sessions)
	
	// Test non-verbose mode
	dialog.SetVerbose(false)
	view := dialog.View()
	assert.Contains(t, view, "Test Session 1")
	assert.NotContains(t, view, "messages") // Should not show message count
	assert.NotContains(t, view, "Updated:") // Should not show update time
	
	// Test verbose mode
	dialog.SetVerbose(true)
	view = dialog.View()
	assert.Contains(t, view, "Test Session 1")
	assert.Contains(t, view, "10") // Should show message count
	assert.Contains(t, view, "messages") // Should show messages word
	assert.Contains(t, view, "Updated:") // Should show update time
	assert.Contains(t, view, "1 hour ago") // Should show relative time
}

func TestFormatTimeAgo(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "just now",
			time:     now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "1 minute ago",
			time:     now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "5 minutes ago",
			time:     now.Add(-5 * time.Minute),
			expected: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			time:     now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "3 hours ago",
			time:     now.Add(-3 * time.Hour),
			expected: "3 hours ago",
		},
		{
			name:     "1 day ago",
			time:     now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "3 days ago",
			time:     now.Add(-72 * time.Hour),
			expected: "3 days ago",
		},
		{
			name:     "old date",
			time:     now.Add(-30 * 24 * time.Hour),
			expected: now.Add(-30 * 24 * time.Hour).Format("Jan 2, 2006"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeAgo(tt.time.Unix())
			assert.Equal(t, tt.expected, result)
		})
	}
}