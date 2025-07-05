package message

import (
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Content(t *testing.T) {
	message := Message{
		Parts: []ContentPart{
			ReasoningContent{Thinking: "Let me think..."},
			TextContent{Text: "Hello, world!"},
		},
	}

	content := message.Content()
	assert.Equal(t, "Hello, world!", content.Text)
}

func TestMessage_ReasoningContent(t *testing.T) {
	message := Message{
		Parts: []ContentPart{
			ReasoningContent{Thinking: "Let me think..."},
			TextContent{Text: "Hello"},
		},
	}

	reasoning := message.ReasoningContent()
	assert.Equal(t, "Let me think...", reasoning.Thinking)
}

func TestMessage_ToolCalls(t *testing.T) {
	message := Message{
		Parts: []ContentPart{
			TextContent{Text: "I'll use a tool"},
			ToolCall{ID: "tool-1", Name: "calculator", Input: `{"op": "add"}`},
			ToolCall{ID: "tool-2", Name: "weather", Input: `{"city": "NYC"}`},
		},
	}

	toolCalls := message.ToolCalls()
	assert.Len(t, toolCalls, 2)
	assert.Equal(t, "tool-1", toolCalls[0].ID)
	assert.Equal(t, "calculator", toolCalls[0].Name)
	assert.Equal(t, "tool-2", toolCalls[1].ID)
}

func TestMessage_IsFinished(t *testing.T) {
	tests := []struct {
		name     string
		parts    []ContentPart
		expected bool
	}{
		{
			name: "message with finish part",
			parts: []ContentPart{
				TextContent{Text: "Hello"},
				Finish{Reason: FinishReasonEndTurn, Time: time.Now().Unix()},
			},
			expected: true,
		},
		{
			name: "message without finish part",
			parts: []ContentPart{
				TextContent{Text: "Hello"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := Message{Parts: tt.parts}
			assert.Equal(t, tt.expected, message.IsFinished())
		})
	}
}

func TestMessage_AppendContent(t *testing.T) {
	message := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
		},
	}

	message.AppendContent(" world!")
	
	content := message.Content()
	assert.Equal(t, "Hello world!", content.Text)
}

func TestMessage_AppendContent_NewContent(t *testing.T) {
	message := Message{
		Parts: []ContentPart{
			ReasoningContent{Thinking: "thinking..."},
		},
	}

	message.AppendContent("Hello")
	
	content := message.Content()
	assert.Equal(t, "Hello", content.Text)
	assert.Len(t, message.Parts, 2) // reasoning + text
}

func TestMessage_AddFinish(t *testing.T) {
	message := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
		},
	}

	message.AddFinish(FinishReasonEndTurn)
	
	assert.True(t, message.IsFinished())
	assert.Equal(t, FinishReasonEndTurn, message.FinishReason())
}

func TestMessage_AddFinish_ReplaceExisting(t *testing.T) {
	message := Message{
		Parts: []ContentPart{
			TextContent{Text: "Hello"},
			Finish{Reason: FinishReasonMaxTokens, Time: 123},
		},
	}

	message.AddFinish(FinishReasonEndTurn)
	
	assert.True(t, message.IsFinished())
	assert.Equal(t, FinishReasonEndTurn, message.FinishReason())
	// Should still only have 2 parts (text + new finish)
	assert.Len(t, message.Parts, 2)
}

func TestBinaryContent_String(t *testing.T) {
	bc := BinaryContent{
		MIMEType: "image/png",
		Data:     []byte("test data"),
	}

	// Test OpenAI format
	openaiResult := bc.String(models.ProviderOpenAI)
	assert.Contains(t, openaiResult, "data:image/png;base64,")

	// Test other provider format
	otherResult := bc.String(models.ProviderAnthropic)
	assert.NotContains(t, otherResult, "data:")
}

func TestMessage_IsThinking(t *testing.T) {
	tests := []struct {
		name     string
		parts    []ContentPart
		expected bool
	}{
		{
			name: "thinking with no text or finish",
			parts: []ContentPart{
				ReasoningContent{Thinking: "Let me think..."},
			},
			expected: true,
		},
		{
			name: "thinking with text",
			parts: []ContentPart{
				ReasoningContent{Thinking: "Let me think..."},
				TextContent{Text: "Hello"},
			},
			expected: false,
		},
		{
			name: "thinking with finish",
			parts: []ContentPart{
				ReasoningContent{Thinking: "Let me think..."},
				Finish{Reason: FinishReasonEndTurn},
			},
			expected: false,
		},
		{
			name: "no thinking",
			parts: []ContentPart{
				TextContent{Text: "Hello"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := Message{Parts: tt.parts}
			assert.Equal(t, tt.expected, message.IsThinking())
		})
	}
}