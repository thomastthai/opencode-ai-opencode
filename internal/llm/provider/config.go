package provider

import (
	"fmt"
	"os"

	"github.com/opencode-ai/opencode/internal/llm/models"
)

// ProviderConfig is the base configuration interface for all providers.
type ProviderConfig interface {
	// GetAPIKey returns the API key for authentication.
	GetAPIKey() string
	
	// GetModel returns the model to be used.
	GetModel() models.Model
	
	// GetMaxTokens returns the maximum number of tokens for responses.
	GetMaxTokens() int64
	
	// GetSystemMessage returns the system message to be used.
	GetSystemMessage() string
	
	// Validate checks if the configuration is valid and returns an error if not.
	Validate() error
}

// BaseProviderConfig provides common configuration fields for all providers.
type BaseProviderConfig struct {
	APIKey        string        `json:"apiKey"`
	Model         models.Model  `json:"model"`
	MaxTokens     int64         `json:"maxTokens"`
	SystemMessage string        `json:"systemMessage"`
}

// GetAPIKey returns the API key.
func (c *BaseProviderConfig) GetAPIKey() string {
	return c.APIKey
}

// GetModel returns the model.
func (c *BaseProviderConfig) GetModel() models.Model {
	return c.Model
}

// GetMaxTokens returns the maximum tokens.
func (c *BaseProviderConfig) GetMaxTokens() int64 {
	return c.MaxTokens
}

// GetSystemMessage returns the system message.
func (c *BaseProviderConfig) GetSystemMessage() string {
	return c.SystemMessage
}

// Validate validates the base configuration.
func (c *BaseProviderConfig) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.Model.ID == "" {
		return fmt.Errorf("model ID is required")
	}
	if c.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be greater than 0, got %d", c.MaxTokens)
	}
	return nil
}

// OpenAIConfig contains OpenAI-specific configuration.
type OpenAIConfig struct {
	BaseProviderConfig
	BaseURL         string            `json:"baseUrl,omitempty"`
	ExtraHeaders    map[string]string `json:"extraHeaders,omitempty"`
	DisableCache    bool              `json:"disableCache,omitempty"`
	ReasoningEffort string            `json:"reasoningEffort,omitempty"`
}

// Validate validates OpenAI configuration.
func (c *OpenAIConfig) Validate() error {
	if err := c.BaseProviderConfig.Validate(); err != nil {
		return fmt.Errorf("OpenAI config validation failed: %w", err)
	}
	
	if c.ReasoningEffort != "" {
		validEfforts := []string{"low", "medium", "high"}
		valid := false
		for _, effort := range validEfforts {
			if c.ReasoningEffort == effort {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("reasoning effort must be one of %v, got '%s'", validEfforts, c.ReasoningEffort)
		}
	}
	
	return nil
}

// AnthropicConfig contains Anthropic-specific configuration.
type AnthropicConfig struct {
	BaseProviderConfig
	UseBedrock   bool                         `json:"useBedrock,omitempty"`
	DisableCache bool                         `json:"disableCache,omitempty"`
	ShouldThink  func(userMessage string) bool `json:"-"` // Not serializable
}

// Validate validates Anthropic configuration.
func (c *AnthropicConfig) Validate() error {
	if err := c.BaseProviderConfig.Validate(); err != nil {
		return fmt.Errorf("Anthropic config validation failed: %w", err)
	}
	return nil
}

// GeminiConfig contains Gemini-specific configuration.
type GeminiConfig struct {
	BaseProviderConfig
	ProjectID string `json:"projectId,omitempty"`
	Location  string `json:"location,omitempty"`
}

// Validate validates Gemini configuration.
func (c *GeminiConfig) Validate() error {
	if err := c.BaseProviderConfig.Validate(); err != nil {
		return fmt.Errorf("Gemini config validation failed: %w", err)
	}
	return nil
}

// BedrockConfig contains AWS Bedrock-specific configuration.
type BedrockConfig struct {
	BaseProviderConfig
	Region          string `json:"region,omitempty"`
	AccessKeyID     string `json:"accessKeyId,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	SessionToken    string `json:"sessionToken,omitempty"`
}

// Validate validates Bedrock configuration.
func (c *BedrockConfig) Validate() error {
	if err := c.BaseProviderConfig.Validate(); err != nil {
		return fmt.Errorf("Bedrock config validation failed: %w", err)
	}
	
	// API key is not required for Bedrock if using IAM roles
	if c.BaseProviderConfig.APIKey == "" && c.AccessKeyID == "" {
		// Check for environment variables
		if os.Getenv("AWS_ACCESS_KEY_ID") == "" && os.Getenv("AWS_PROFILE") == "" {
			return fmt.Errorf("Bedrock requires either API key, access key ID, AWS_ACCESS_KEY_ID environment variable, or AWS_PROFILE")
		}
	}
	
	return nil
}

// CopilotConfig contains GitHub Copilot-specific configuration.
type CopilotConfig struct {
	BaseProviderConfig
	BearerToken     string            `json:"bearerToken,omitempty"`
	ExtraHeaders    map[string]string `json:"extraHeaders,omitempty"`
	ReasoningEffort string            `json:"reasoningEffort,omitempty"`
}

// Validate validates Copilot configuration.
func (c *CopilotConfig) Validate() error {
	if err := c.BaseProviderConfig.Validate(); err != nil {
		return fmt.Errorf("Copilot config validation failed: %w", err)
	}
	return nil
}

// AzureConfig contains Azure OpenAI-specific configuration.
type AzureConfig struct {
	BaseProviderConfig
	Endpoint     string `json:"endpoint"`
	Deployment   string `json:"deployment,omitempty"`
	APIVersion   string `json:"apiVersion,omitempty"`
	UseAzureAD   bool   `json:"useAzureAd,omitempty"`
}

// Validate validates Azure configuration.
func (c *AzureConfig) Validate() error {
	if err := c.BaseProviderConfig.Validate(); err != nil {
		return fmt.Errorf("Azure config validation failed: %w", err)
	}
	
	if c.Endpoint == "" {
		return fmt.Errorf("Azure endpoint is required")
	}
	
	return nil
}

// VertexAIConfig contains Google Vertex AI-specific configuration.
type VertexAIConfig struct {
	BaseProviderConfig
	ProjectID   string `json:"projectId"`
	Location    string `json:"location"`
	Credentials string `json:"credentials,omitempty"` // Path to service account JSON
}

// Validate validates Vertex AI configuration.
func (c *VertexAIConfig) Validate() error {
	if err := c.BaseProviderConfig.Validate(); err != nil {
		return fmt.Errorf("Vertex AI config validation failed: %w", err)
	}
	
	if c.ProjectID == "" {
		return fmt.Errorf("Vertex AI project ID is required")
	}
	
	if c.Location == "" {
		return fmt.Errorf("Vertex AI location is required")
	}
	
	return nil
}

// MockConfig contains mock provider configuration for testing.
type MockConfig struct {
	BaseProviderConfig
	Responses       []string          `json:"responses,omitempty"`
	StreamEvents    []ProviderEvent   `json:"streamEvents,omitempty"`
	StreamEventSets [][]ProviderEvent `json:"streamEventSets,omitempty"`
	ErrorToReturn   string            `json:"errorToReturn,omitempty"`
	ToolSupport     bool              `json:"toolSupport,omitempty"`
	StreamSupport   bool              `json:"streamSupport,omitempty"`
}

// Validate validates mock configuration.
func (c *MockConfig) Validate() error {
	// Mock provider doesn't require API key
	if c.Model.ID == "" {
		return fmt.Errorf("model ID is required")
	}
	if c.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be greater than 0, got %d", c.MaxTokens)
	}
	return nil
}