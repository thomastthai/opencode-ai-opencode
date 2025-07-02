package provider_test

import (
	"context"
	"os"
	"testing"

	configprovider "github.com/opencode-ai/opencode/internal/config/provider"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/llm/provider"
	"github.com/opencode-ai/opencode/internal/llm/tools"
	"github.com/opencode-ai/opencode/internal/message"
)

// TestProviderIntegration demonstrates the complete provider architecture workflow.
func TestProviderIntegration(t *testing.T) {
	t.Run("FullWorkflow", func(t *testing.T) {
		// 1. Check available providers
		loader := configprovider.NewProviderConfigLoader()
		
		// Set up environment for testing
		os.Setenv("OPENAI_API_KEY", "test-key")
		defer os.Unsetenv("OPENAI_API_KEY")
		
		availableProviders := loader.GetAvailableProviders()
		t.Logf("Available providers: %v", availableProviders)
		
		// Should include OpenAI
		found := false
		for _, p := range availableProviders {
			if p == models.ProviderOpenAI {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("OpenAI provider should be available")
		}
		
		// 2. Load provider configuration
		model := models.Model{
			ID:                  "gpt-4",
			DefaultMaxTokens:    4096,
			SupportsAttachments: true,
		}
		
		config, err := loader.LoadProviderConfig(models.ProviderOpenAI, model)
		if err != nil {
			t.Fatalf("Failed to load provider config: %v", err)
		}
		
		// 3. Create provider instance using registry
		providerInstance, err := provider.NewProvider(models.ProviderOpenAI, config)
		if err != nil {
			t.Fatalf("Failed to create provider: %v", err)
		}
		
		// 4. Test basic Provider interface
		if providerInstance.Model().ID != "gpt-4" {
			t.Errorf("Expected model ID 'gpt-4', got '%s'", providerInstance.Model().ID)
		}
		
		// 5. Test optional interfaces using type assertions
		// Note: We don't actually call the methods in this test since
		// we're using incomplete configuration for demonstration
		
		// Test StreamProvider interface
		if streamProvider, ok := providerInstance.(provider.StreamProvider); ok {
			t.Log("Provider supports streaming")
			_ = streamProvider // Just verify the interface is implemented
		} else {
			t.Log("Provider does not support streaming")
		}
		
		// Test ToolCallingProvider interface
		if toolProvider, ok := providerInstance.(provider.ToolCallingProvider); ok {
			t.Log("Provider supports tool calling")
			
			if !toolProvider.SupportsToolCalling() {
				t.Error("Provider should support tool calling")
			}
		} else {
			t.Log("Provider does not support tool calling")
		}
		
		// Test ReasoningProvider interface
		if reasoningProvider, ok := providerInstance.(provider.ReasoningProvider); ok {
			t.Log("Provider supports reasoning")
			
			if !reasoningProvider.SupportsReasoning() {
				t.Log("Provider model does not support reasoning")
			} else {
				err := reasoningProvider.SetReasoningEffort("high")
				if err != nil {
					t.Errorf("Failed to set reasoning effort: %v", err)
				}
			}
		} else {
			t.Log("Provider does not support reasoning")
		}
		
		// Test CachingProvider interface
		if cachingProvider, ok := providerInstance.(provider.CachingProvider); ok {
			t.Log("Provider supports caching")
			
			if !cachingProvider.SupportsCaching() {
				t.Error("Provider should support caching")
			}
			
			cachingProvider.SetCacheEnabled(true)
		} else {
			t.Log("Provider does not support caching")
		}
		
		// Test AttachmentProvider interface
		if attachmentProvider, ok := providerInstance.(provider.AttachmentProvider); ok {
			t.Log("Provider supports attachments")
			
			if !attachmentProvider.SupportsAttachments() {
				t.Error("Provider should support attachments")
			}
			
			mimeTypes := attachmentProvider.GetSupportedMimeTypes()
			if len(mimeTypes) == 0 {
				t.Error("Provider should return supported MIME types")
			}
			t.Logf("Supported MIME types: %v", mimeTypes)
		} else {
			t.Log("Provider does not support attachments")
		}
		
		// 6. Test provider registry functions
		registeredProviders := provider.ListRegisteredProviders()
		t.Logf("Registered providers: %d", len(registeredProviders))
		
		// Should include at least Mock, OpenAI, and Anthropic
		expectedProviders := []models.ModelProvider{
			models.ProviderMock,
			models.ProviderOpenAI,
			models.ProviderAnthropic,
		}
		
		for _, expected := range expectedProviders {
			found := false
			for _, registered := range registeredProviders {
				if registered.Name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected provider '%s' to be registered", expected)
			}
		}
		
		// Test getting provider info
		info, err := provider.GetProviderInfo(models.ProviderOpenAI)
		if err != nil {
			t.Fatalf("Failed to get provider info: %v", err)
		}
		
		if info.Name != models.ProviderOpenAI {
			t.Errorf("Expected provider name '%s', got '%s'", models.ProviderOpenAI, info.Name)
		}
		
		if len(info.Capabilities) == 0 {
			t.Error("Provider should have capabilities listed")
		}
		
		t.Logf("Provider info: %+v", info)
	})
	
	t.Run("MockProviderWorkflow", func(t *testing.T) {
		// Test using mock provider for testing scenarios
		model := models.Model{
			ID:               "mock-model",
			DefaultMaxTokens: 1000,
		}
		
		config := &provider.MockConfig{
			BaseProviderConfig: provider.BaseProviderConfig{
				APIKey:    "test-key", // Mock doesn't validate API key
				Model:     model,
				MaxTokens: 1000,
			},
			Responses:     []string{"Hello, world!", "How can I help you?"},
			ToolSupport:   true,
			StreamSupport: true,
		}
		
		providerInstance, err := provider.NewProvider(models.ProviderMock, config)
		if err != nil {
			t.Fatalf("Failed to create mock provider: %v", err)
		}
		
		// Test all interfaces are implemented
		interfaces := []string{}
		
		if _, ok := providerInstance.(provider.Provider); ok {
			interfaces = append(interfaces, "Provider")
		}
		if _, ok := providerInstance.(provider.StreamProvider); ok {
			interfaces = append(interfaces, "StreamProvider")
		}
		if _, ok := providerInstance.(provider.ToolCallingProvider); ok {
			interfaces = append(interfaces, "ToolCallingProvider")
		}
		if _, ok := providerInstance.(provider.ReasoningProvider); ok {
			interfaces = append(interfaces, "ReasoningProvider")
		}
		if _, ok := providerInstance.(provider.CachingProvider); ok {
			interfaces = append(interfaces, "CachingProvider")
		}
		if _, ok := providerInstance.(provider.AttachmentProvider); ok {
			interfaces = append(interfaces, "AttachmentProvider")
		}
		
		expectedInterfaces := 6 // All interfaces should be implemented
		if len(interfaces) != expectedInterfaces {
			t.Errorf("Expected mock provider to implement %d interfaces, got %d: %v", 
				expectedInterfaces, len(interfaces), interfaces)
		}
		
		t.Logf("Mock provider implements: %v", interfaces)
		
		// Test basic functionality
		ctx := context.Background()
		messages := []message.Message{}
		tools := []tools.BaseTool{}
		
		response, err := providerInstance.SendMessages(ctx, messages, tools)
		if err != nil {
			t.Fatalf("Mock provider SendMessages failed: %v", err)
		}
		
		if response.Content != "Hello, world!" {
			t.Errorf("Expected 'Hello, world!', got '%s'", response.Content)
		}
		
		// Test second call
		response, err = providerInstance.SendMessages(ctx, messages, tools)
		if err != nil {
			t.Fatalf("Mock provider SendMessages failed: %v", err)
		}
		
		if response.Content != "How can I help you?" {
			t.Errorf("Expected 'How can I help you?', got '%s'", response.Content)
		}
	})
	
	t.Run("ErrorHandling", func(t *testing.T) {
		// Test error handling in the new architecture
		
		// 1. Test invalid provider type
		_, err := provider.NewProvider("invalid-provider", &provider.MockConfig{})
		if err == nil {
			t.Error("Expected error for invalid provider type")
		}
		
		// 2. Test invalid config type
		_, err = provider.NewProvider(models.ProviderOpenAI, &provider.MockConfig{})
		if err == nil {
			t.Error("Expected error for wrong config type")
		}
		
		// 3. Test config validation
		invalidConfig := &provider.OpenAIConfig{
			BaseProviderConfig: provider.BaseProviderConfig{
				// Missing required fields
			},
		}
		
		_, err = provider.NewProvider(models.ProviderOpenAI, invalidConfig)
		if err == nil {
			t.Error("Expected error for invalid config")
		}
		
		// 4. Test error from config loader
		loader := configprovider.NewProviderConfigLoader()
		
		// Clean environment to ensure no providers are available
		originalEnv := make(map[string]string)
		envVars := []string{
			"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY",
			"AWS_ACCESS_KEY_ID", "GITHUB_TOKEN", "AZURE_OPENAI_ENDPOINT",
		}
		
		for _, env := range envVars {
			originalEnv[env] = os.Getenv(env)
			os.Unsetenv(env)
		}
		
		defer func() {
			for env, value := range originalEnv {
				if value != "" {
					os.Setenv(env, value)
				}
			}
		}()
		
		model := models.Model{ID: "test-model"}
		_, err = loader.LoadProviderConfig(models.ProviderOpenAI, model)
		if err == nil {
			t.Error("Expected error when no configuration is available")
		}
	})
}