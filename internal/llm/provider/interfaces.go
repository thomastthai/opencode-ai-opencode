package provider

import (
	"context"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/llm/tools"
	"github.com/opencode-ai/opencode/internal/message"
)

// Provider is the base interface that all providers must implement.
// It provides basic functionality for sending messages and getting model information.
type Provider interface {
	// SendMessages sends a batch of messages to the provider and returns a response.
	SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error)
	
	// Model returns the model configuration used by this provider.
	Model() models.Model
}

// StreamProvider extends Provider to support streaming responses.
// Providers that support real-time streaming should implement this interface.
type StreamProvider interface {
	Provider
	
	// StreamResponse sends messages and returns a channel for streaming events.
	StreamResponse(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent
}

// ToolCallingProvider extends Provider to indicate support for tool/function calling.
// Providers that can invoke external tools should implement this interface.
type ToolCallingProvider interface {
	Provider
	
	// SupportsToolCalling returns true if the provider supports tool calling.
	SupportsToolCalling() bool
}

// ReasoningProvider extends Provider to indicate support for reasoning models.
// Providers that support models with reasoning capabilities should implement this interface.
type ReasoningProvider interface {
	Provider
	
	// SupportsReasoning returns true if the provider supports reasoning models.
	SupportsReasoning() bool
	
	// SetReasoningEffort sets the reasoning effort level (e.g., "low", "medium", "high").
	SetReasoningEffort(effort string) error
}

// CachingProvider extends Provider to indicate support for prompt caching.
// Providers that support caching previous context should implement this interface.
type CachingProvider interface {
	Provider
	
	// SupportsCaching returns true if the provider supports prompt caching.
	SupportsCaching() bool
	
	// SetCacheEnabled enables or disables caching for this provider.
	SetCacheEnabled(enabled bool)
}

// AttachmentProvider extends Provider to indicate support for file attachments.
// Providers that can process images, documents, or other files should implement this interface.
type AttachmentProvider interface {
	Provider
	
	// SupportsAttachments returns true if the provider supports file attachments.
	SupportsAttachments() bool
	
	// GetSupportedMimeTypes returns the MIME types supported for attachments.
	GetSupportedMimeTypes() []string
}