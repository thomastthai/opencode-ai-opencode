package config

import (
	"testing"
)

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		name     string
	}{
		{"", true, "empty string"},
		{"#ffffff", true, "valid 6-digit hex"},
		{"#000000", true, "valid 6-digit hex black"},
		{"#abc", true, "valid 3-digit hex"},
		{"#ABC", true, "valid 3-digit hex uppercase"},
		{"#123456", true, "valid 6-digit hex with numbers"},
		{"#a1b2c3", true, "valid 6-digit hex mixed"},
		{"ffffff", false, "missing hash"},
		{"#gggggg", false, "invalid characters"},
		{"#12345", false, "invalid length 5"},
		{"#1234567", false, "invalid length 7"},
		{"#zz", false, "invalid length 2"},
		{"red", false, "color name"},
		{"#", false, "just hash"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isValidHexColor(test.input)
			if result != test.expected {
				t.Errorf("isValidHexColor(%q) = %v, want %v", test.input, result, test.expected)
			}
		})
	}
}

func TestValidateMessageLayoutConfig(t *testing.T) {
	cfg := &Config{
		TUI: TUIConfig{
			MessageLayoutConfig: MessageLayoutConfig{
				UserTextColor:            "#ffffff",
				UserBackgroundColor:      "#007acc",
				AssistantTextColor:       "#000000",
				AssistantBackgroundColor: "#f0f0f0",
			},
		},
	}

	err := validateMessageLayoutConfig(cfg)
	if err != nil {
		t.Errorf("validateMessageLayoutConfig() with valid colors failed: %v", err)
	}

	// Test with invalid colors
	cfg.TUI.MessageLayoutConfig.UserTextColor = "invalid"
	cfg.TUI.MessageLayoutConfig.AssistantBackgroundColor = "#gggggg"

	err = validateMessageLayoutConfig(cfg)
	if err != nil {
		t.Errorf("validateMessageLayoutConfig() should not return error for invalid colors, should just reset them: %v", err)
	}

	// Check that invalid colors were reset
	if cfg.TUI.MessageLayoutConfig.UserTextColor != "" {
		t.Errorf("Invalid UserTextColor should have been reset to empty string")
	}
	if cfg.TUI.MessageLayoutConfig.AssistantBackgroundColor != "" {
		t.Errorf("Invalid AssistantBackgroundColor should have been reset to empty string")
	}
}

func TestValidateBorderConfig(t *testing.T) {
	tests := []struct {
		name           string
		border         BorderConfig
		expectedChar   string
		expectedHorizontalChar string
		expectedFgValid bool
		expectedBgValid bool
	}{
		{
			name: "valid config",
			border: BorderConfig{
				Character:       "│",
				HorizontalChar:  "─",
				TopLeftChar:     "┌",
				TopRightChar:    "┐",
				BottomLeftChar:  "└",
				BottomRightChar: "┘",
				ForegroundColor: "#ff0000",
				BackgroundColor: "#00ff00",
			},
			expectedChar:          "│",
			expectedHorizontalChar: "─",
			expectedFgValid:       true,
			expectedBgValid:       true,
		},
		{
			name: "invalid characters too long",
			border: BorderConfig{
				Character:       "abc",
				HorizontalChar:  "def",
				TopLeftChar:     "xyz",
				ForegroundColor: "#ff0000",
				BackgroundColor: "#00ff00",
			},
			expectedChar:          "│", // Should reset to default
			expectedHorizontalChar: "─", // Should reset to default
			expectedFgValid:       true,
			expectedBgValid:       true,
		},
		{
			name: "invalid colors",
			border: BorderConfig{
				Character:       "│",
				HorizontalChar:  "─",
				ForegroundColor: "invalid",
				BackgroundColor: "#gggggg",
			},
			expectedChar:          "│",
			expectedHorizontalChar: "─",
			expectedFgValid:       false, // Should be reset to empty
			expectedBgValid:       false, // Should be reset to empty
		},
		{
			name: "empty values",
			border: BorderConfig{
				Character:       "",
				HorizontalChar:  "",
				ForegroundColor: "",
				BackgroundColor: "",
			},
			expectedChar:          "",
			expectedHorizontalChar: "",
			expectedFgValid:       true, // Empty is valid
			expectedBgValid:       true, // Empty is valid
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			border := test.border
			validateBorderConfig(&border, "test")

			if border.Character != test.expectedChar {
				t.Errorf("Expected character %q, got %q", test.expectedChar, border.Character)
			}

			if border.HorizontalChar != test.expectedHorizontalChar {
				t.Errorf("Expected horizontal character %q, got %q", test.expectedHorizontalChar, border.HorizontalChar)
			}

			if test.expectedFgValid && border.ForegroundColor == "" && test.border.ForegroundColor != "" {
				t.Errorf("Valid foreground color was incorrectly reset")
			}
			if !test.expectedFgValid && border.ForegroundColor != "" {
				t.Errorf("Invalid foreground color was not reset")
			}

			if test.expectedBgValid && border.BackgroundColor == "" && test.border.BackgroundColor != "" {
				t.Errorf("Valid background color was incorrectly reset")
			}
			if !test.expectedBgValid && border.BackgroundColor != "" {
				t.Errorf("Invalid background color was not reset")
			}
		})
	}
}