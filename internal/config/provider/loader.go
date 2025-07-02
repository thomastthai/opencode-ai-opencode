// Package provider provides centralized configuration loading and validation for LLM providers.
package provider

import (
	"fmt"
	"os"
	"strings"

	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/llm/provider"
)

// ProviderConfigLoader handles loading provider configurations from various sources.
type ProviderConfigLoader struct {
	// ConfigSources defines the priority order for configuration sources
	ConfigSources []ConfigSource
}

// ConfigSource represents a source of configuration data.
type ConfigSource interface {
	// GetProviderConfig retrieves configuration for a specific provider
	GetProviderConfig(providerType models.ModelProvider) (provider.ProviderConfig, error)
	
	// IsAvailable checks if this config source has data for the provider
	IsAvailable(providerType models.ModelProvider) bool
}

// EnvironmentConfigSource loads configuration from environment variables.
type EnvironmentConfigSource struct{}

// GetProviderConfig implements ConfigSource for environment variables.
func (e *EnvironmentConfigSource) GetProviderConfig(providerType models.ModelProvider) (provider.ProviderConfig, error) {
	switch providerType {
	case models.ProviderOpenAI:
		return e.getOpenAIConfig()
	case models.ProviderAnthropic:
		return e.getAnthropicConfig()
	case models.ProviderGemini:
		return e.getGeminiConfig()
	case models.ProviderBedrock:
		return e.getBedrockConfig()
	case models.ProviderCopilot:
		return e.getCopilotConfig()
	case models.ProviderAzure:
		return e.getAzureConfig()
	case models.ProviderVertexAI:
		return e.getVertexAIConfig()
	default:
		return nil, fmt.Errorf("environment config not supported for provider '%s'", providerType)
	}
}

// IsAvailable implements ConfigSource for environment variables.
func (e *EnvironmentConfigSource) IsAvailable(providerType models.ModelProvider) bool {
	switch providerType {
	case models.ProviderOpenAI:
		return os.Getenv("OPENAI_API_KEY") != ""
	case models.ProviderAnthropic:
		return os.Getenv("ANTHROPIC_API_KEY") != ""
	case models.ProviderGemini:
		return os.Getenv("GEMINI_API_KEY") != ""
	case models.ProviderBedrock:
		return e.hasAWSCredentials()
	case models.ProviderCopilot:
		return e.hasCopilotCredentials()
	case models.ProviderAzure:
		return os.Getenv("AZURE_OPENAI_ENDPOINT") != ""
	case models.ProviderVertexAI:
		return e.hasVertexAICredentials()
	default:
		return false
	}
}

func (e *EnvironmentConfigSource) getOpenAIConfig() (*provider.OpenAIConfig, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	config := &provider.OpenAIConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			APIKey: apiKey,
		},
		BaseURL:         os.Getenv("OPENAI_BASE_URL"),
		ReasoningEffort: getEnvWithDefault("OPENAI_REASONING_EFFORT", "medium"),
	}

	return config, nil
}

func (e *EnvironmentConfigSource) getAnthropicConfig() (*provider.AnthropicConfig, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}

	config := &provider.AnthropicConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			APIKey: apiKey,
		},
		UseBedrock: strings.ToLower(os.Getenv("ANTHROPIC_USE_BEDROCK")) == "true",
	}

	return config, nil
}

func (e *EnvironmentConfigSource) getGeminiConfig() (*provider.GeminiConfig, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	config := &provider.GeminiConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			APIKey: apiKey,
		},
		ProjectID: os.Getenv("GEMINI_PROJECT_ID"),
		Location:  getEnvWithDefault("GEMINI_LOCATION", "us-central1"),
	}

	return config, nil
}

func (e *EnvironmentConfigSource) getBedrockConfig() (*provider.BedrockConfig, error) {
	config := &provider.BedrockConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			APIKey: "", // Bedrock doesn't use API key directly
		},
		Region:          getEnvWithDefault("AWS_REGION", "us-east-1"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
	}

	return config, nil
}

func (e *EnvironmentConfigSource) getCopilotConfig() (*provider.CopilotConfig, error) {
	apiKey := os.Getenv("GITHUB_TOKEN")
	if apiKey == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required for Copilot")
	}

	config := &provider.CopilotConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			APIKey: apiKey,
		},
		ReasoningEffort: getEnvWithDefault("COPILOT_REASONING_EFFORT", "medium"),
	}

	return config, nil
}

func (e *EnvironmentConfigSource) getAzureConfig() (*provider.AzureConfig, error) {
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("AZURE_OPENAI_ENDPOINT environment variable is required")
	}

	config := &provider.AzureConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			APIKey: os.Getenv("AZURE_OPENAI_API_KEY"), // Can be empty for Azure AD
		},
		Endpoint:   endpoint,
		Deployment: os.Getenv("AZURE_OPENAI_DEPLOYMENT"),
		APIVersion: getEnvWithDefault("AZURE_OPENAI_API_VERSION", "2024-02-01"),
		UseAzureAD: strings.ToLower(os.Getenv("AZURE_USE_AD")) == "true",
	}

	return config, nil
}

func (e *EnvironmentConfigSource) getVertexAIConfig() (*provider.VertexAIConfig, error) {
	projectID := os.Getenv("VERTEX_AI_PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("VERTEX_AI_PROJECT_ID environment variable is required")
	}

	config := &provider.VertexAIConfig{
		BaseProviderConfig: provider.BaseProviderConfig{
			APIKey: "", // Vertex AI uses service account credentials
		},
		ProjectID:   projectID,
		Location:    getEnvWithDefault("VERTEX_AI_LOCATION", "us-central1"),
		Credentials: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
	}

	return config, nil
}

func (e *EnvironmentConfigSource) hasAWSCredentials() bool {
	return os.Getenv("AWS_ACCESS_KEY_ID") != "" || 
		   os.Getenv("AWS_PROFILE") != "" ||
		   os.Getenv("AWS_ROLE_ARN") != ""
}

func (e *EnvironmentConfigSource) hasCopilotCredentials() bool {
	return os.Getenv("GITHUB_TOKEN") != ""
}

func (e *EnvironmentConfigSource) hasVertexAICredentials() bool {
	return os.Getenv("VERTEX_AI_PROJECT_ID") != "" && 
		   (os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" || 
		   os.Getenv("GOOGLE_CLOUD_PROJECT") != "")
}

// NewProviderConfigLoader creates a new configuration loader with default sources.
func NewProviderConfigLoader() *ProviderConfigLoader {
	return &ProviderConfigLoader{
		ConfigSources: []ConfigSource{
			&EnvironmentConfigSource{},
			// TODO: Add file-based config source
			// TODO: Add viper-based config source
		},
	}
}

// LoadProviderConfig loads configuration for a provider from available sources.
func (l *ProviderConfigLoader) LoadProviderConfig(providerType models.ModelProvider, model models.Model) (provider.ProviderConfig, error) {
	for _, source := range l.ConfigSources {
		if source.IsAvailable(providerType) {
			config, err := source.GetProviderConfig(providerType)
			if err != nil {
				continue // Try next source
			}
			
			// Set model and default values
			if err := l.setModelDefaults(config, model); err != nil {
				return nil, fmt.Errorf("failed to set model defaults: %w", err)
			}
			
			// Validate the configuration
			if err := config.Validate(); err != nil {
				return nil, fmt.Errorf("provider configuration validation failed: %w", err)
			}
			
			return config, nil
		}
	}
	
	return nil, fmt.Errorf("no configuration source available for provider '%s'", providerType)
}

// setModelDefaults sets model-specific defaults on the configuration.
func (l *ProviderConfigLoader) setModelDefaults(config provider.ProviderConfig, model models.Model) error {
	// Use type assertion to set model on the base config
	switch c := config.(type) {
	case *provider.OpenAIConfig:
		c.Model = model
		if c.MaxTokens == 0 {
			c.MaxTokens = model.DefaultMaxTokens
		}
	case *provider.AnthropicConfig:
		c.Model = model
		if c.MaxTokens == 0 {
			c.MaxTokens = model.DefaultMaxTokens
		}
	case *provider.GeminiConfig:
		c.Model = model
		if c.MaxTokens == 0 {
			c.MaxTokens = model.DefaultMaxTokens
		}
	case *provider.BedrockConfig:
		c.Model = model
		if c.MaxTokens == 0 {
			c.MaxTokens = model.DefaultMaxTokens
		}
	case *provider.CopilotConfig:
		c.Model = model
		if c.MaxTokens == 0 {
			c.MaxTokens = model.DefaultMaxTokens
		}
	case *provider.AzureConfig:
		c.Model = model
		if c.MaxTokens == 0 {
			c.MaxTokens = model.DefaultMaxTokens
		}
	case *provider.VertexAIConfig:
		c.Model = model
		if c.MaxTokens == 0 {
			c.MaxTokens = model.DefaultMaxTokens
		}
	default:
		return fmt.Errorf("unsupported config type: %T", config)
	}
	
	return nil
}

// GetAvailableProviders returns a list of providers that have configuration available.
func (l *ProviderConfigLoader) GetAvailableProviders() []models.ModelProvider {
	var available []models.ModelProvider
	
	providers := []models.ModelProvider{
		models.ProviderOpenAI,
		models.ProviderAnthropic,
		models.ProviderGemini,
		models.ProviderBedrock,
		models.ProviderCopilot,
		models.ProviderAzure,
		models.ProviderVertexAI,
	}
	
	for _, providerType := range providers {
		for _, source := range l.ConfigSources {
			if source.IsAvailable(providerType) {
				available = append(available, providerType)
				break
			}
		}
	}
	
	return available
}

// Helper function to get environment variable with default value.
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}