package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageLayoutDefaults(t *testing.T) {
	// Test default values are applied correctly
	setDefaults(false)
	
	assert.Equal(t, MessageLayoutClassic, MessageLayout("classic"))
	assert.Equal(t, MessageLayoutMessaging, MessageLayout("messaging"))
}

func TestMessageLayoutConfig(t *testing.T) {
	cfg := &Config{
		TUI: TUIConfig{
			MessageLayout: MessageLayoutMessaging,
			MessageLayoutConfig: MessageLayoutConfig{
				UserMessageWidth:      0.65,
				AssistantMessageWidth: 0.75,
				UserRightMargin:       10,
				AssistantLeftMargin:   2,
				UseBackgrounds:        true,
				UseRoundedBorders:     true,
			},
		},
	}
	
	assert.Equal(t, MessageLayoutMessaging, cfg.TUI.MessageLayout)
	assert.Equal(t, 0.65, cfg.TUI.MessageLayoutConfig.UserMessageWidth)
	assert.Equal(t, 0.75, cfg.TUI.MessageLayoutConfig.AssistantMessageWidth)
	assert.Equal(t, 10, cfg.TUI.MessageLayoutConfig.UserRightMargin)
	assert.Equal(t, 2, cfg.TUI.MessageLayoutConfig.AssistantLeftMargin)
	assert.True(t, cfg.TUI.MessageLayoutConfig.UseBackgrounds)
	assert.True(t, cfg.TUI.MessageLayoutConfig.UseRoundedBorders)
}

func TestMessageLayoutConstants(t *testing.T) {
	assert.Equal(t, "classic", string(MessageLayoutClassic))
	assert.Equal(t, "messaging", string(MessageLayoutMessaging))
}