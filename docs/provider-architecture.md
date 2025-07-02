# Provider Architecture Documentation

## Overview

The OpenCode LLM provider system has been refactored to provide a modular, scalable, and easily extensible architecture. This document outlines the new provider system and provides guidelines for implementing new providers.

## Architecture Components

### 1. Provider Interfaces

The provider system uses interface segregation to define different capabilities:

```go
// Base interface that all providers must implement
type Provider interface {
    SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error)
    Model() models.Model
}

// Optional interfaces for additional capabilities
type StreamProvider interface {
    Provider
    StreamResponse(ctx context.Context, messages []message.Message, tools []tools.BaseTool) <-chan ProviderEvent
}

type ToolCallingProvider interface {
    Provider
    SupportsToolCalling() bool
}

type ReasoningProvider interface {
    Provider
    SupportsReasoning() bool
    SetReasoningEffort(effort string) error
}

type CachingProvider interface {
    Provider
    SupportsCaching() bool
    SetCacheEnabled(enabled bool)
}

type AttachmentProvider interface {
    Provider
    SupportsAttachments() bool
    GetSupportedMimeTypes() []string
}
```

### 2. Provider Registry

The provider registry is a thread-safe system for managing provider factories:

```go
// Register a provider
RegisterProvider(models.ProviderOpenAI, NewOpenAIProvider, ProviderInfo{
    Name:        models.ProviderOpenAI,
    Description: "OpenAI provider supporting GPT models",
    Capabilities: []string{"streaming", "tool_calling", "reasoning", "caching", "attachments"},
})

// Create a provider instance
provider, err := NewProvider(models.ProviderOpenAI, config)

// List registered providers
providers := ListRegisteredProviders()
```

### 3. Configuration System

Each provider has its own configuration struct with validation:

```go
type OpenAIConfig struct {
    BaseProviderConfig
    BaseURL         string            `json:"baseUrl,omitempty"`
    ExtraHeaders    map[string]string `json:"extraHeaders,omitempty"`
    DisableCache    bool              `json:"disableCache,omitempty"`
    ReasoningEffort string            `json:"reasoningEffort,omitempty"`
}

func (c *OpenAIConfig) Validate() error {
    // Provider-specific validation logic
}
```

### 4. Configuration Loader

The configuration loader centralizes provider configuration from multiple sources:

```go
loader := provider.NewProviderConfigLoader()

// Load configuration for a provider from environment variables or config files
config, err := loader.LoadProviderConfig(models.ProviderOpenAI, model)

// Get available providers based on current configuration
available := loader.GetAvailableProviders()
```

## Implementation Guidelines

### Adding a New Provider

1. **Define Provider Configuration**:
   ```go
   type MyProviderConfig struct {
       BaseProviderConfig
       CustomField string `json:"customField"`
   }
   
   func (c *MyProviderConfig) Validate() error {
       if err := c.BaseProviderConfig.Validate(); err != nil {
           return fmt.Errorf("MyProvider config validation failed: %w", err)
       }
       // Add custom validation
       return nil
   }
   ```

2. **Implement Provider**:
   ```go
   type MyProvider struct {
       config MyProviderConfig
       client *MyClient
   }
   
   // Implement required Provider interface
   func (p *MyProvider) SendMessages(ctx context.Context, messages []message.Message, tools []tools.BaseTool) (*ProviderResponse, error) {
       // Implementation
   }
   
   func (p *MyProvider) Model() models.Model {
       return p.config.GetModel()
   }
   
   // Implement optional interfaces as needed
   func (p *MyProvider) SupportsToolCalling() bool {
       return true // if provider supports tool calling
   }
   ```

3. **Create Factory Function**:
   ```go
   func NewMyProvider(config ProviderConfig) (Provider, error) {
       myConfig, ok := config.(*MyProviderConfig)
       if !ok {
           return nil, fmt.Errorf("MyProvider requires MyProviderConfig, got %T", config)
       }
       
       if err := myConfig.Validate(); err != nil {
           return nil, fmt.Errorf("invalid MyProvider configuration: %w", err)
       }
       
       return &MyProvider{
           config: *myConfig,
           client: newMyClient(myConfig),
       }, nil
   }
   ```

4. **Register Provider**:
   ```go
   func init() {
       RegisterProvider(models.ProviderMyProvider, NewMyProvider, ProviderInfo{
           Name:        models.ProviderMyProvider,
           Description: "My custom provider",
           Capabilities: []string{"streaming", "tool_calling"},
       })
   }
   ```

### Using Type Assertions for Optional Capabilities

```go
// Check if provider supports streaming
if streamProvider, ok := provider.(StreamProvider); ok {
    eventChan := streamProvider.StreamResponse(ctx, messages, tools)
    // Handle streaming
}

// Check if provider supports tool calling
if toolProvider, ok := provider.(ToolCallingProvider); ok {
    if toolProvider.SupportsToolCalling() {
        // Use tools
    }
}
```

### Testing with Mock Provider

```go
func TestMyFeature(t *testing.T) {
    config := &MockConfig{
        BaseProviderConfig: BaseProviderConfig{
            Model:     models.Model{ID: "test-model"},
            MaxTokens: 1000,
        },
        Responses:     []string{"Test response"},
        ToolSupport:   true,
        StreamSupport: true,
    }
    
    provider, err := NewProvider(models.ProviderMock, config)
    if err != nil {
        t.Fatalf("Failed to create mock provider: %v", err)
    }
    
    // Test your feature with the mock provider
}
```

## Error Handling

All errors in the provider system are descriptive and actionable:

- **Configuration Errors**: Include specific validation failures
- **Provider Errors**: Include provider name and operation context
- **Registry Errors**: Include available alternatives when possible

```go
// Example error messages:
"provider 'openai' is not registered. Available providers: [anthropic, gemini, mock]"
"OpenAI config validation failed: API key is required"
"failed to create provider 'anthropic': invalid Anthropic provider configuration: ANTHROPIC_API_KEY environment variable is required"
```

## Backward Compatibility

The refactored system maintains backward compatibility:

- Legacy `NewLegacyProvider` function still works
- Existing code can gradually migrate to the new registry system
- Configuration can be loaded from existing sources

## Benefits

1. **Modularity**: Each provider is self-contained with its own configuration
2. **Scalability**: Easy to add new providers without modifying core code
3. **Extensibility**: Interface segregation allows providers to advertise their capabilities
4. **Testability**: Mock provider supports all interfaces for comprehensive testing
5. **Configuration**: Centralized loading from multiple sources with validation
6. **Thread Safety**: Registry operations are thread-safe for concurrent use

## Migration Guide

### From Legacy System

1. Replace `NewLegacyProvider` calls with `NewProvider`
2. Use provider-specific configuration structs instead of options
3. Use type assertions to check for optional capabilities
4. Update tests to use the mock provider

### Example Migration

**Before**:
```go
provider, err := NewLegacyProvider(models.ProviderOpenAI, 
    WithAPIKey("key"),
    WithModel(model),
    WithMaxTokens(4000),
)
```

**After**:
```go
config := &OpenAIConfig{
    BaseProviderConfig: BaseProviderConfig{
        APIKey:    "key",
        Model:     model,
        MaxTokens: 4000,
    },
}

provider, err := NewProvider(models.ProviderOpenAI, config)
```