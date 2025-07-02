package provider

import (
	"os"
	"testing"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/llm/provider"
)

func TestEnvironmentConfigSource(t *testing.T) {
	source := &EnvironmentConfigSource{}
	
	t.Run("OpenAI", func(t *testing.T) {
		// Clean environment
		os.Unsetenv("OPENAI_API_KEY")
		
		// Test unavailable
		if source.IsAvailable(models.ProviderOpenAI) {
			t.Error("Should not be available without API key")
		}
		
		// Set environment variable
		os.Setenv("OPENAI_API_KEY", "test-key")
		defer os.Unsetenv("OPENAI_API_KEY")
		
		// Test available
		if !source.IsAvailable(models.ProviderOpenAI) {
			t.Error("Should be available with API key")
		}
		
		// Test config loading
		config, err := source.GetProviderConfig(models.ProviderOpenAI)
		if err != nil {
			t.Fatalf("Failed to get config: %v", err)
		}
		
		openaiConfig, ok := config.(*provider.OpenAIConfig)
		if !ok {
			t.Fatalf("Expected OpenAIConfig, got %T", config)
		}
		
		if openaiConfig.GetAPIKey() != "test-key" {
			t.Errorf("Expected API key 'test-key', got '%s'", openaiConfig.GetAPIKey())
		}
	})
	
	t.Run("Anthropic", func(t *testing.T) {
		// Clean environment
		os.Unsetenv("ANTHROPIC_API_KEY")
		
		// Test unavailable
		if source.IsAvailable(models.ProviderAnthropic) {
			t.Error("Should not be available without API key")
		}
		
		// Set environment variable
		os.Setenv("ANTHROPIC_API_KEY", "test-key")
		defer os.Unsetenv("ANTHROPIC_API_KEY")
		
		// Test available
		if !source.IsAvailable(models.ProviderAnthropic) {
			t.Error("Should be available with API key")
		}
		
		// Test config loading
		config, err := source.GetProviderConfig(models.ProviderAnthropic)
		if err != nil {
			t.Fatalf("Failed to get config: %v", err)
		}
		
		anthropicConfig, ok := config.(*provider.AnthropicConfig)
		if !ok {
			t.Fatalf("Expected AnthropicConfig, got %T", config)
		}
		
		if anthropicConfig.GetAPIKey() != "test-key" {
			t.Errorf("Expected API key 'test-key', got '%s'", anthropicConfig.GetAPIKey())
		}
	})
	
	t.Run("Bedrock", func(t *testing.T) {
		// Clean environment
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_PROFILE")
		
		// Test unavailable
		if source.IsAvailable(models.ProviderBedrock) {
			t.Error("Should not be available without AWS credentials")
		}
		
		// Set environment variable
		os.Setenv("AWS_ACCESS_KEY_ID", "test-key")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
		defer func() {
			os.Unsetenv("AWS_ACCESS_KEY_ID")
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		}()
		
		// Test available
		if !source.IsAvailable(models.ProviderBedrock) {
			t.Error("Should be available with AWS credentials")
		}
		
		// Test config loading
		config, err := source.GetProviderConfig(models.ProviderBedrock)
		if err != nil {
			t.Fatalf("Failed to get config: %v", err)
		}
		
		bedrockConfig, ok := config.(*provider.BedrockConfig)
		if !ok {
			t.Fatalf("Expected BedrockConfig, got %T", config)
		}
		
		if bedrockConfig.AccessKeyID != "test-key" {
			t.Errorf("Expected access key 'test-key', got '%s'", bedrockConfig.AccessKeyID)
		}
	})
}

func TestProviderConfigLoader(t *testing.T) {
	loader := NewProviderConfigLoader()
	
	t.Run("LoadProviderConfig", func(t *testing.T) {
		// Set up environment for OpenAI
		os.Setenv("OPENAI_API_KEY", "test-key")
		defer os.Unsetenv("OPENAI_API_KEY")
		
		model := models.Model{
			ID:               "gpt-4",
			DefaultMaxTokens: 4096,
		}
		
		config, err := loader.LoadProviderConfig(models.ProviderOpenAI, model)
		if err != nil {
			t.Fatalf("Failed to load provider config: %v", err)
		}
		
		if config.GetAPIKey() != "test-key" {
			t.Errorf("Expected API key 'test-key', got '%s'", config.GetAPIKey())
		}
		
		if config.GetModel().ID != "gpt-4" {
			t.Errorf("Expected model ID 'gpt-4', got '%s'", config.GetModel().ID)
		}
		
		if config.GetMaxTokens() != 4096 {
			t.Errorf("Expected max tokens 4096, got %d", config.GetMaxTokens())
		}
	})
	
	t.Run("GetAvailableProviders", func(t *testing.T) {
		// Clean environment
		for _, env := range []string{
			"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY",
			"AWS_ACCESS_KEY_ID", "GITHUB_TOKEN", "AZURE_OPENAI_ENDPOINT",
			"VERTEX_AI_PROJECT_ID",
		} {
			os.Unsetenv(env)
		}
		
		// Should have no available providers
		available := loader.GetAvailableProviders()
		if len(available) != 0 {
			t.Errorf("Expected no available providers, got %d", len(available))
		}
		
		// Set OpenAI key
		os.Setenv("OPENAI_API_KEY", "test-key")
		defer os.Unsetenv("OPENAI_API_KEY")
		
		available = loader.GetAvailableProviders()
		if len(available) != 1 {
			t.Errorf("Expected 1 available provider, got %d", len(available))
		}
		
		if available[0] != models.ProviderOpenAI {
			t.Errorf("Expected OpenAI provider, got %s", available[0])
		}
		
		// Add Anthropic key
		os.Setenv("ANTHROPIC_API_KEY", "test-key")
		defer os.Unsetenv("ANTHROPIC_API_KEY")
		
		available = loader.GetAvailableProviders()
		if len(available) != 2 {
			t.Errorf("Expected 2 available providers, got %d", len(available))
		}
	})
	
	t.Run("InvalidProvider", func(t *testing.T) {
		model := models.Model{ID: "test-model"}
		
		_, err := loader.LoadProviderConfig("invalid-provider", model)
		if err == nil {
			t.Error("Expected error for invalid provider")
		}
	})
}