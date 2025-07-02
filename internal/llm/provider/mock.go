package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/llm/tools"
	"github.com/opencode-ai/opencode/internal/message"
)

// MockProvider is a provider implementation for testing purposes.
// It implements all optional provider interfaces and can be configured
// to return specific responses or errors.
type MockProvider struct {
	config MockConfig
	callCount int
}

// Ensure MockProvider implements all interfaces
var _ Provider = (*MockProvider)(nil)
var _ StreamProvider = (*MockProvider)(nil)
var _ ToolCallingProvider = (*MockProvider)(nil)
var _ ReasoningProvider = (*MockProvider)(nil)
var _ CachingProvider = (*MockProvider)(nil)
var _ AttachmentProvider = (*MockProvider)(nil)

// NewMockProvider creates a new mock provider with the given configuration.
func NewMockProvider(config ProviderConfig) (Provider, error) {
	mockConfig, ok := config.(*MockConfig)
	if !ok {
		return nil, fmt.Errorf("mock provider requires MockConfig, got %T", config)
	}
	
	if err := mockConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid mock provider configuration: %w", err)
	}
	
	return &MockProvider{
		config: *mockConfig,
		callCount: 0,
	}, nil
}

// SendMessages implements Provider interface.
func (m *MockProvider) SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error) {
	m.callCount++
	
	// Return error if configured to do so
	if m.config.ErrorToReturn != "" {
		return nil, fmt.Errorf("%s", m.config.ErrorToReturn)
	}
	
	// Return configured response or default
	var content string
	if len(m.config.Responses) > 0 {
		responseIndex := (m.callCount - 1) % len(m.config.Responses)
		content = m.config.Responses[responseIndex]
	} else {
		content = fmt.Sprintf("Mock response %d for messages: %d", m.callCount, len(messages))
	}
	
	// Simulate tool calls if tools are provided and tool support is enabled
	var toolCalls []message.ToolCall
	if m.config.ToolSupport && len(tools) > 0 {
		toolInfo := tools[0].Info()
		toolCalls = []message.ToolCall{
			{
				ID:    "mock_tool_call_1",
				Name:  toolInfo.Name,
				Input: `{"test": "mock_argument"}`,
				Type:  "function",
			},
		}
	}
	
	return &ProviderResponse{
		Content:   content,
		ToolCalls: toolCalls,
		Usage: TokenUsage{
			InputTokens:  int64(len(fmt.Sprintf("%v", messages)) / 4), // Rough estimation
			OutputTokens: int64(len(content) / 4),
		},
		FinishReason: message.FinishReasonEndTurn,
	}, nil
}

// StreamResponse implements StreamProvider interface.
func (m *MockProvider) StreamResponse(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent {
	resultChan := make(chan ProviderEvent, 10)
	
	if !m.config.StreamSupport {
		// If streaming not supported, just send the regular response as a complete event
		go func() {
			defer close(resultChan)
			
			response, err := m.SendMessages(ctx, messages, tools)
			if err != nil {
				resultChan <- ProviderEvent{
					Type:  EventError,
					Error: err,
				}
				return
			}
			
			resultChan <- ProviderEvent{
				Type:     EventComplete,
				Response: response,
			}
		}()
		
		return resultChan
	}
	
	// Send configured stream events or generate default ones
	go func() {
		defer close(resultChan)
		
		if len(m.config.StreamEvents) > 0 {
			// Send configured events
			for _, event := range m.config.StreamEvents {
				select {
				case resultChan <- event:
				case <-ctx.Done():
					return
				}
				time.Sleep(10 * time.Millisecond) // Small delay to simulate streaming
			}
		} else {
			// Generate default streaming events
			response, err := m.SendMessages(ctx, messages, tools)
			if err != nil {
				resultChan <- ProviderEvent{
					Type:  EventError,
					Error: err,
				}
				return
			}
			
			// Simulate streaming by sending content in chunks
			content := response.Content
			chunkSize := 10
			for i := 0; i < len(content); i += chunkSize {
				end := i + chunkSize
				if end > len(content) {
					end = len(content)
				}
				
				chunk := content[i:end]
				
				if i == 0 {
					resultChan <- ProviderEvent{Type: EventContentStart}
				}
				
				resultChan <- ProviderEvent{
					Type:    EventContentDelta,
					Content: chunk,
				}
				
				select {
				case <-ctx.Done():
					return
				default:
					time.Sleep(10 * time.Millisecond)
				}
			}
			
			resultChan <- ProviderEvent{Type: EventContentStop}
			resultChan <- ProviderEvent{
				Type:     EventComplete,
				Response: response,
			}
		}
	}()
	
	return resultChan
}

// Model implements Provider interface.
func (m *MockProvider) Model() models.Model {
	return m.config.GetModel()
}

// SupportsToolCalling implements ToolCallingProvider interface.
func (m *MockProvider) SupportsToolCalling() bool {
	return m.config.ToolSupport
}

// SupportsReasoning implements ReasoningProvider interface.
func (m *MockProvider) SupportsReasoning() bool {
	return true // Mock provider supports all features
}

// SetReasoningEffort implements ReasoningProvider interface.
func (m *MockProvider) SetReasoningEffort(effort string) error {
	validEfforts := []string{"low", "medium", "high"}
	for _, valid := range validEfforts {
		if effort == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid reasoning effort '%s', must be one of %v", effort, validEfforts)
}

// SupportsCaching implements CachingProvider interface.
func (m *MockProvider) SupportsCaching() bool {
	return true // Mock provider supports all features
}

// SetCacheEnabled implements CachingProvider interface.
func (m *MockProvider) SetCacheEnabled(enabled bool) {
	// Mock implementation - just store the setting
}

// SupportsAttachments implements AttachmentProvider interface.
func (m *MockProvider) SupportsAttachments() bool {
	return true // Mock provider supports all features
}

// GetSupportedMimeTypes implements AttachmentProvider interface.
func (m *MockProvider) GetSupportedMimeTypes() []string {
	return []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"text/plain",
		"application/pdf",
	}
}

// GetCallCount returns the number of times SendMessages has been called.
// This is useful for testing to verify the number of API calls.
func (m *MockProvider) GetCallCount() int {
	return m.callCount
}

// Reset resets the call count and other internal state.
// This is useful for testing to reset the provider between test cases.
func (m *MockProvider) Reset() {
	m.callCount = 0
}

func init() {
	// Register the mock provider
	RegisterProvider(models.ProviderMock, NewMockProvider, ProviderInfo{
		Name:        models.ProviderMock,
		Description: "Mock provider for testing purposes",
		Capabilities: []string{
			"streaming",
			"tool_calling", 
			"reasoning",
			"caching",
			"attachments",
		},
	})
}