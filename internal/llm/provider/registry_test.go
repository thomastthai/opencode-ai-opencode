package provider

import (
	"context"
	"testing"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/llm/tools"
	"github.com/opencode-ai/opencode/internal/message"
)

func TestProviderRegistry(t *testing.T) {
	// Create a test registry
	registry := NewTestRegistry()
	
	t.Run("RegisterProvider", func(t *testing.T) {
		// Test provider registration
		factory := func(config ProviderConfig) (Provider, error) {
			return NewMockProvider(config)
		}
		
		info := ProviderInfo{
			Name:        models.ProviderMock,
			Description: "Test mock provider",
			Capabilities: []string{"streaming", "tool_calling"},
		}
		
		err := registry.RegisterProvider(models.ProviderMock, factory, info)
		if err != nil {
			t.Fatalf("Failed to register provider: %v", err)
		}
		
		// Test duplicate registration
		err = registry.RegisterProvider(models.ProviderMock, factory, info)
		if err == nil {
			t.Fatal("Expected error for duplicate provider registration")
		}
	})
	
	t.Run("NewProvider", func(t *testing.T) {
		// Register a mock provider first
		factory := func(config ProviderConfig) (Provider, error) {
			return NewMockProvider(config)
		}
		
		info := ProviderInfo{
			Name:        models.ProviderMock,
			Description: "Test mock provider",
		}
		
		registry.RegisterProvider(models.ProviderMock, factory, info)
		
		// Create a provider instance
		config := &MockConfig{
			BaseProviderConfig: BaseProviderConfig{
				APIKey:    "test-key",
				Model:     models.Model{ID: "test-model"},
				MaxTokens: 1000,
			},
		}
		
		provider, err := registry.NewProvider(models.ProviderMock, config)
		if err != nil {
			t.Fatalf("Failed to create provider: %v", err)
		}
		
		if provider == nil {
			t.Fatal("Provider is nil")
		}
		
		// Test non-existent provider
		_, err = registry.NewProvider("non-existent", config)
		if err == nil {
			t.Fatal("Expected error for non-existent provider")
		}
	})
	
	t.Run("ListRegisteredProviders", func(t *testing.T) {
		registry := NewTestRegistry()
		
		// Register multiple providers
		providers := []models.ModelProvider{
			models.ProviderMock,
			models.ProviderOpenAI,
		}
		
		for _, provider := range providers {
			factory := func(config ProviderConfig) (Provider, error) {
				return NewMockProvider(config)
			}
			
			info := ProviderInfo{
				Name:        provider,
				Description: "Test provider",
			}
			
			registry.RegisterProvider(provider, factory, info)
		}
		
		list := registry.ListRegisteredProviders()
		if len(list) != len(providers) {
			t.Fatalf("Expected %d providers, got %d", len(providers), len(list))
		}
	})
	
	t.Run("GetProviderInfo", func(t *testing.T) {
		registry := NewTestRegistry()
		
		factory := func(config ProviderConfig) (Provider, error) {
			return NewMockProvider(config)
		}
		
		expectedInfo := ProviderInfo{
			Name:        models.ProviderMock,
			Description: "Test mock provider",
			Capabilities: []string{"streaming", "tool_calling"},
		}
		
		registry.RegisterProvider(models.ProviderMock, factory, expectedInfo)
		
		info, err := registry.GetProviderInfo(models.ProviderMock)
		if err != nil {
			t.Fatalf("Failed to get provider info: %v", err)
		}
		
		if info.Name != expectedInfo.Name {
			t.Errorf("Expected name %s, got %s", expectedInfo.Name, info.Name)
		}
		
		if info.Description != expectedInfo.Description {
			t.Errorf("Expected description %s, got %s", expectedInfo.Description, info.Description)
		}
		
		// Test non-existent provider
		_, err = registry.GetProviderInfo("non-existent")
		if err == nil {
			t.Fatal("Expected error for non-existent provider")
		}
	})
	
	t.Run("IsRegistered", func(t *testing.T) {
		registry := NewTestRegistry()
		
		if registry.IsRegistered(models.ProviderMock) {
			t.Fatal("Provider should not be registered initially")
		}
		
		factory := func(config ProviderConfig) (Provider, error) {
			return NewMockProvider(config)
		}
		
		info := ProviderInfo{Name: models.ProviderMock}
		registry.RegisterProvider(models.ProviderMock, factory, info)
		
		if !registry.IsRegistered(models.ProviderMock) {
			t.Fatal("Provider should be registered")
		}
	})
	
	t.Run("UnregisterProvider", func(t *testing.T) {
		registry := NewTestRegistry()
		
		factory := func(config ProviderConfig) (Provider, error) {
			return NewMockProvider(config)
		}
		
		info := ProviderInfo{Name: models.ProviderMock}
		registry.RegisterProvider(models.ProviderMock, factory, info)
		
		if !registry.IsRegistered(models.ProviderMock) {
			t.Fatal("Provider should be registered")
		}
		
		registry.UnregisterProvider(models.ProviderMock)
		
		if registry.IsRegistered(models.ProviderMock) {
			t.Fatal("Provider should not be registered after unregistering")
		}
	})
	
	t.Run("Clear", func(t *testing.T) {
		registry := NewTestRegistry()
		
		// Register some providers
		factory := func(config ProviderConfig) (Provider, error) {
			return NewMockProvider(config)
		}
		
		info := ProviderInfo{Name: models.ProviderMock}
		registry.RegisterProvider(models.ProviderMock, factory, info)
		registry.RegisterProvider(models.ProviderOpenAI, factory, info)
		
		list := registry.ListRegisteredProviders()
		if len(list) == 0 {
			t.Fatal("Expected providers to be registered")
		}
		
		registry.Clear()
		
		list = registry.ListRegisteredProviders()
		if len(list) != 0 {
			t.Fatal("Expected no providers after clear")
		}
	})
}

func TestMockProvider(t *testing.T) {
	t.Run("BasicFunctionality", func(t *testing.T) {
		config := &MockConfig{
			BaseProviderConfig: BaseProviderConfig{
				APIKey:    "test-key",
				Model:     models.Model{ID: "test-model"},
				MaxTokens: 1000,
			},
			Responses: []string{"Hello", "World"},
		}
		
		provider, err := NewMockProvider(config)
		if err != nil {
			t.Fatalf("Failed to create mock provider: %v", err)
		}
		
		mockProvider := provider.(*MockProvider)
		
		// Test SendMessages
		ctx := context.Background()
		messages := []message.Message{}
		tools := []tools.BaseTool{}
		
		response, err := mockProvider.SendMessages(ctx, messages, tools)
		if err != nil {
			t.Fatalf("SendMessages failed: %v", err)
		}
		
		if response.Content != "Hello" {
			t.Errorf("Expected 'Hello', got '%s'", response.Content)
		}
		
		// Test second call
		response, err = mockProvider.SendMessages(ctx, messages, tools)
		if err != nil {
			t.Fatalf("SendMessages failed: %v", err)
		}
		
		if response.Content != "World" {
			t.Errorf("Expected 'World', got '%s'", response.Content)
		}
		
		// Test cycling back to first response
		response, err = mockProvider.SendMessages(ctx, messages, tools)
		if err != nil {
			t.Fatalf("SendMessages failed: %v", err)
		}
		
		if response.Content != "Hello" {
			t.Errorf("Expected 'Hello', got '%s'", response.Content)
		}
	})
	
	t.Run("InterfaceImplementations", func(t *testing.T) {
		config := &MockConfig{
			BaseProviderConfig: BaseProviderConfig{
				Model:     models.Model{ID: "test-model"},
				MaxTokens: 1000,
			},
			ToolSupport:   true,
			StreamSupport: true,
		}
		
		provider, err := NewMockProvider(config)
		if err != nil {
			t.Fatalf("Failed to create mock provider: %v", err)
		}
		
		// Test Provider interface
		if _, ok := provider.(Provider); !ok {
			t.Fatal("Mock provider should implement Provider interface")
		}
		
		// Test StreamProvider interface
		if streamProvider, ok := provider.(StreamProvider); !ok {
			t.Fatal("Mock provider should implement StreamProvider interface")
		} else {
			ctx := context.Background()
			messages := []message.Message{}
			tools := []tools.BaseTool{}
			
			eventChan := streamProvider.StreamResponse(ctx, messages, tools)
			
			// Read at least one event
			event := <-eventChan
			if event.Type == EventError {
				t.Fatalf("Stream response returned error: %v", event.Error)
			}
		}
		
		// Test ToolCallingProvider interface
		if toolProvider, ok := provider.(ToolCallingProvider); !ok {
			t.Fatal("Mock provider should implement ToolCallingProvider interface")
		} else {
			if !toolProvider.SupportsToolCalling() {
				t.Fatal("Mock provider should support tool calling when configured")
			}
		}
		
		// Test ReasoningProvider interface
		if reasoningProvider, ok := provider.(ReasoningProvider); !ok {
			t.Fatal("Mock provider should implement ReasoningProvider interface")
		} else {
			if !reasoningProvider.SupportsReasoning() {
				t.Fatal("Mock provider should support reasoning")
			}
			
			err := reasoningProvider.SetReasoningEffort("high")
			if err != nil {
				t.Fatalf("SetReasoningEffort failed: %v", err)
			}
			
			err = reasoningProvider.SetReasoningEffort("invalid")
			if err == nil {
				t.Fatal("SetReasoningEffort should fail for invalid effort")
			}
		}
		
		// Test CachingProvider interface
		if cachingProvider, ok := provider.(CachingProvider); !ok {
			t.Fatal("Mock provider should implement CachingProvider interface")
		} else {
			if !cachingProvider.SupportsCaching() {
				t.Fatal("Mock provider should support caching")
			}
			
			cachingProvider.SetCacheEnabled(true)
			cachingProvider.SetCacheEnabled(false)
		}
		
		// Test AttachmentProvider interface
		if attachmentProvider, ok := provider.(AttachmentProvider); !ok {
			t.Fatal("Mock provider should implement AttachmentProvider interface")
		} else {
			if !attachmentProvider.SupportsAttachments() {
				t.Fatal("Mock provider should support attachments")
			}
			
			mimeTypes := attachmentProvider.GetSupportedMimeTypes()
			if len(mimeTypes) == 0 {
				t.Fatal("Mock provider should return supported MIME types")
			}
		}
	})
	
	t.Run("ErrorHandling", func(t *testing.T) {
		config := &MockConfig{
			BaseProviderConfig: BaseProviderConfig{
				Model:     models.Model{ID: "test-model"},
				MaxTokens: 1000,
			},
			ErrorToReturn: "test error",
		}
		
		provider, err := NewMockProvider(config)
		if err != nil {
			t.Fatalf("Failed to create mock provider: %v", err)
		}
		
		ctx := context.Background()
		messages := []message.Message{}
		tools := []tools.BaseTool{}
		
		_, err = provider.SendMessages(ctx, messages, tools)
		if err == nil {
			t.Fatal("Expected error from mock provider")
		}
		
		if err.Error() != "test error" {
			t.Errorf("Expected 'test error', got '%s'", err.Error())
		}
	})
}