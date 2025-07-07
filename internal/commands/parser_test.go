package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandParser_Parse(t *testing.T) {
	parser := NewCommandParser(nil)

	tests := []struct {
		name     string
		input    string
		expected SlashCommand
	}{
		{
			name:  "empty slash",
			input: "/",
			expected: SlashCommand{
				Raw:        "/",
				Topic:      "",
				Verb:       "",
				Args:       nil,
				Incomplete: true,
			},
		},
		{
			name:  "topic only without space",
			input: "/session",
			expected: SlashCommand{
				Raw:        "/session",
				Topic:      "session",
				Verb:       "",
				Args:       nil,
				Incomplete: true,
			},
		},
		{
			name:  "topic only with space",
			input: "/session ",
			expected: SlashCommand{
				Raw:        "/session ",
				Topic:      "session",
				Verb:       "",
				Args:       nil,
				Incomplete: true,
			},
		},
		{
			name:  "topic and verb without space",
			input: "/session new",
			expected: SlashCommand{
				Raw:        "/session new",
				Topic:      "session",
				Verb:       "new",
				Args:       nil,
				Incomplete: true,
			},
		},
		{
			name:  "topic and verb with space",
			input: "/session new ",
			expected: SlashCommand{
				Raw:        "/session new ",
				Topic:      "session",
				Verb:       "new",
				Args:       nil,
				Incomplete: false,
			},
		},
		{
			name:  "complete command with args",
			input: "/session new my-session",
			expected: SlashCommand{
				Raw:        "/session new my-session",
				Topic:      "session",
				Verb:       "new",
				Args:       []string{"my-session"},
				Incomplete: true,
			},
		},
		{
			name:  "complete command with multiple args",
			input: "/session compact focus on errors",
			expected: SlashCommand{
				Raw:        "/session compact focus on errors",
				Topic:      "session",
				Verb:       "compact",
				Args:       []string{"focus", "on", "errors"},
				Incomplete: true,
			},
		},
		{
			name:  "no slash prefix",
			input: "session new",
			expected: SlashCommand{
				Raw:        "session new",
				Topic:      "",
				Verb:       "",
				Args:       nil,
				Incomplete: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommandParser_GetParseState(t *testing.T) {
	parser := NewCommandParser(nil)

	tests := []struct {
		name     string
		input    string
		expected ParseState
	}{
		{
			name:     "empty slash",
			input:    "/",
			expected: ParseStateTopic,
		},
		{
			name:     "typing topic",
			input:    "/sess",
			expected: ParseStateTopic,
		},
		{
			name:     "topic with space",
			input:    "/session ",
			expected: ParseStateVerb,
		},
		{
			name:     "typing verb",
			input:    "/session ne",
			expected: ParseStateVerb,
		},
		{
			name:     "verb with space",
			input:    "/session new ",
			expected: ParseStateArgs,
		},
		{
			name:     "typing args",
			input:    "/session new my-sess",
			expected: ParseStateArgs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := parser.Parse(tt.input)
			state := parser.GetParseState(parsed)
			assert.Equal(t, tt.expected, state, "Parse state should match expected")
		})
	}
}

func TestCommandParser_GetCompletions(t *testing.T) {
	parser := NewCommandParser(nil)

	t.Run("topic completions", func(t *testing.T) {
		parsed := parser.Parse("/")
		completions := parser.GetCompletions(parsed)
		
		assert.Greater(t, len(completions), 0, "Should have topic completions")
		
		// Verify we have expected topics
		topics := make(map[string]bool)
		for _, c := range completions {
			topics[c.Value] = true
		}
		
		assert.True(t, topics["session"], "Should have session topic")
		assert.True(t, topics["config"], "Should have config topic")
		assert.True(t, topics["auth"], "Should have auth topic")
	})

	t.Run("filtered topic completions", func(t *testing.T) {
		parsed := parser.Parse("/se")
		completions := parser.GetCompletions(parsed)
		
		assert.Equal(t, 1, len(completions), "Should have one matching topic")
		assert.Equal(t, "session", completions[0].Value)
		assert.Equal(t, "/session ", completions[0].Complete)
	})

	t.Run("verb completions", func(t *testing.T) {
		parsed := parser.Parse("/session ")
		completions := parser.GetCompletions(parsed)
		
		assert.Greater(t, len(completions), 0, "Should have verb completions")
		
		// Verify we have expected verbs
		verbs := make(map[string]bool)
		for _, c := range completions {
			verbs[c.Value] = true
		}
		
		assert.True(t, verbs["new"], "Should have new verb")
		assert.True(t, verbs["list"], "Should have list verb")
		assert.True(t, verbs["compact"], "Should have compact verb")
	})

	t.Run("filtered verb completions", func(t *testing.T) {
		parsed := parser.Parse("/session ne")
		completions := parser.GetCompletions(parsed)
		
		assert.Equal(t, 1, len(completions), "Should have one matching verb")
		assert.Equal(t, "new", completions[0].Value)
		assert.Equal(t, "/session new ", completions[0].Complete)
	})

	t.Run("no completions for unknown topic", func(t *testing.T) {
		parsed := parser.Parse("/unknown ")
		completions := parser.GetCompletions(parsed)
		
		assert.Equal(t, 0, len(completions), "Should have no completions for unknown topic")
	})
}

func TestCommandParser_TabCompletion(t *testing.T) {
	parser := NewCommandParser(nil)

	tests := []struct {
		name              string
		input             string
		expectedComplete  string
		hasMultipleOptions bool
	}{
		{
			name:              "unique topic match",
			input:             "/se",
			expectedComplete:  "/session ",
			hasMultipleOptions: false,
		},
		{
			name:              "ambiguous topic match",
			input:             "/s",
			expectedComplete:  "/s", // Should return original with options
			hasMultipleOptions: true,
		},
		{
			name:              "unique verb match",
			input:             "/session ne",
			expectedComplete:  "/session new ",
			hasMultipleOptions: false,
		},
		{
			name:              "ambiguous verb match",
			input:             "/session c",
			expectedComplete:  "/session c", // clear, compact, cost
			hasMultipleOptions: true,
		},
		{
			name:              "no matches",
			input:             "/xyz",
			expectedComplete:  "/xyz",
			hasMultipleOptions: false,
		},
		{
			name:              "already complete",
			input:             "/help",
			expectedComplete:  "/help ",
			hasMultipleOptions: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completed, options := parser.GetTabCompletion(tt.input)
			assert.Equal(t, tt.expectedComplete, completed)
			
			if tt.hasMultipleOptions {
				assert.Greater(t, len(options), 1, "Should have multiple options")
			} else if tt.expectedComplete != tt.input {
				assert.Nil(t, options, "Should have no options when unique match")
			}
		})
	}
}

func TestCommandParser_CommonPrefix(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected string
	}{
		{
			name:     "identical strings",
			a:        "session",
			b:        "session",
			expected: "session",
		},
		{
			name:     "common prefix",
			a:        "session",
			b:        "system",
			expected: "s",
		},
		{
			name:     "no common prefix",
			a:        "auth",
			b:        "session",
			expected: "",
		},
		{
			name:     "one empty string",
			a:        "session",
			b:        "",
			expected: "",
		},
		{
			name:     "different lengths",
			a:        "session",
			b:        "sess",
			expected: "sess",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := commonPrefix(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test state transitions and temporal stability
func TestCommandParser_StateTransitions(t *testing.T) {
	parser := NewCommandParser(nil)

	t.Run("progressive command building", func(t *testing.T) {
		// Simulate user typing progressively
		inputs := []string{
			"/",
			"/s",
			"/se",
			"/session",
			"/session ",
			"/session n",
			"/session new",
			"/session new ",
			"/session new my-session",
		}

		expectedStates := []ParseState{
			ParseStateTopic,
			ParseStateTopic,
			ParseStateTopic,
			ParseStateTopic,
			ParseStateVerb,
			ParseStateVerb,
			ParseStateVerb,
			ParseStateArgs,
			ParseStateArgs,
		}

		for i, input := range inputs {
			parsed := parser.Parse(input)
			state := parser.GetParseState(parsed)
			assert.Equal(t, expectedStates[i], state, 
				"State should be correct for input: %s", input)
			
			// Verify completions are appropriate for state
			completions := parser.GetCompletions(parsed)
			if state == ParseStateArgs {
				// Args state might have no completions (free-form)
				continue
			}
			
			// For non-arg states, we should have completions unless it's an unknown command
			if parsed.Topic == "" || (state == ParseStateTopic && len(parsed.Topic) > 0) ||
			   (state == ParseStateVerb && parser.isValidTopic(parsed.Topic)) {
				assert.NotNil(t, completions, 
					"Should have completions for input: %s", input)
			}
		}
	})
}

// Helper method for testing
func (p *CommandParser) isValidTopic(topic string) bool {
	validTopics := []string{"session", "config", "project", "auth", "dev", "system", "help"}
	for _, valid := range validTopics {
		if valid == topic {
			return true
		}
	}
	return false
}