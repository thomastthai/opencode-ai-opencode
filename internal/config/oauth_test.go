package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/llm/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOAuth2Service(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()

	// Save original environment variables
	originalClientID := os.Getenv("GEMINI_OAUTH_CLIENT_ID")
	originalClientSecret := os.Getenv("GEMINI_OAUTH_CLIENT_SECRET")

	// Test with environment variables
	os.Setenv("GEMINI_OAUTH_CLIENT_ID", "test-client-id")
	os.Setenv("GEMINI_OAUTH_CLIENT_SECRET", "test-client-secret")

	service := GetOAuth2Service()
	assert.NotNil(t, service)
	assert.True(t, service.HasValidCredentials())

	// Restore environment variables
	if originalClientID != "" {
		os.Setenv("GEMINI_OAUTH_CLIENT_ID", originalClientID)
	} else {
		os.Unsetenv("GEMINI_OAUTH_CLIENT_ID")
	}
	if originalClientSecret != "" {
		os.Setenv("GEMINI_OAUTH_CLIENT_SECRET", originalClientSecret)
	} else {
		os.Unsetenv("GEMINI_OAUTH_CLIENT_SECRET")
	}
}

func TestGetOAuth2Service_FromConfig(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()

	// Save original environment variables
	originalClientID := os.Getenv("GEMINI_OAUTH_CLIENT_ID")
	originalClientSecret := os.Getenv("GEMINI_OAUTH_CLIENT_SECRET")

	// Clear environment variables to test config file
	os.Unsetenv("GEMINI_OAUTH_CLIENT_ID")
	os.Unsetenv("GEMINI_OAUTH_CLIENT_SECRET")

	// Create test config directory
	tmpDir := t.TempDir()

	// Create config file with OAuth2 settings
	configData := map[string]interface{}{
		"providers": map[string]interface{}{
			"gemini": map[string]interface{}{
				"oauth2": map[string]interface{}{
					"clientId":     "config-client-id",
					"clientSecret": "config-client-secret",
				},
			},
		},
	}

	configFile := filepath.Join(tmpDir, ".opencode.json")
	data, err := json.Marshal(configData)
	require.NoError(t, err)
	err = os.WriteFile(configFile, data, 0600)
	require.NoError(t, err)

	// Load config
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	cfg = config

	service := GetOAuth2Service()
	assert.NotNil(t, service)
	assert.True(t, service.HasValidCredentials())

	// Restore environment variables
	if originalClientID != "" {
		os.Setenv("GEMINI_OAUTH_CLIENT_ID", originalClientID)
	}
	if originalClientSecret != "" {
		os.Setenv("GEMINI_OAUTH_CLIENT_SECRET", originalClientSecret)
	}
}

func TestGetGeminiAuthStatus(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()

	// Save original environment variables
	originalHome := os.Getenv("HOME")
	originalToken := os.Getenv("GEMINI_TOKEN")
	originalAPIKey := os.Getenv("GEMINI_API_KEY")

	// Create temporary directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("GEMINI_TOKEN")
	os.Unsetenv("GEMINI_API_KEY")

	// Test with no authentication
	status, isValid := GetGeminiAuthStatus()
	assert.Contains(t, status, "No authentication")
	assert.False(t, isValid)

	// Test with environment token
	os.Setenv("GEMINI_TOKEN", "test-env-token")
	status, isValid = GetGeminiAuthStatus()
	assert.Contains(t, status, "Environment variable token")
	assert.True(t, isValid)

	// Test with OAuth file
	os.Unsetenv("GEMINI_TOKEN")
	configDir := filepath.Join(tmpDir, ".config", "gemini")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	tokenFile := filepath.Join(configDir, "oauth_creds.json")
	tokenData := oauth.TokenInfo{
		AccessToken: "test-oauth-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	data, err := json.Marshal(tokenData)
	require.NoError(t, err)
	err = os.WriteFile(tokenFile, data, 0600)
	require.NoError(t, err)

	status, isValid = GetGeminiAuthStatus()
	assert.Contains(t, status, "OAuth2 token")
	assert.True(t, isValid)

	// Test with API key
	os.Remove(tokenFile)
	os.Setenv("GEMINI_API_KEY", "test-api-key")
	status, isValid = GetGeminiAuthStatus()
	assert.Contains(t, status, "API key")
	assert.True(t, isValid)

	// Restore environment variables
	if originalHome != "" {
		os.Setenv("HOME", originalHome)
	}
	if originalToken != "" {
		os.Setenv("GEMINI_TOKEN", originalToken)
	}
	if originalAPIKey != "" {
		os.Setenv("GEMINI_API_KEY", originalAPIKey)
	}
}

func TestGetGeminiAuthMethod(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()

	// Test default (auto)
	method := GetGeminiAuthMethod()
	assert.Equal(t, "auto", method)

	// Test with config
	tmpDir := t.TempDir()
	configData := map[string]interface{}{
		"providers": map[string]interface{}{
			"gemini": map[string]interface{}{
				"authMethod": "oauth2",
			},
		},
	}

	configFile := filepath.Join(tmpDir, ".opencode.json")
	data, err := json.Marshal(configData)
	require.NoError(t, err)
	err = os.WriteFile(configFile, data, 0600)
	require.NoError(t, err)

	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	cfg = config

	method = GetGeminiAuthMethod()
	assert.Equal(t, "oauth2", method)
}

func TestLoginWithGeminiOAuth2_InvalidCredentials(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()

	// Save original environment variables
	originalClientID := os.Getenv("GEMINI_OAUTH_CLIENT_ID")
	originalClientSecret := os.Getenv("GEMINI_OAUTH_CLIENT_SECRET")

	// Clear environment variables
	os.Unsetenv("GEMINI_OAUTH_CLIENT_ID")
	os.Unsetenv("GEMINI_OAUTH_CLIENT_SECRET")

	ctx := context.Background()
	err := LoginWithGeminiOAuth2(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth2 credentials not configured")

	// Restore environment variables
	if originalClientID != "" {
		os.Setenv("GEMINI_OAUTH_CLIENT_ID", originalClientID)
	}
	if originalClientSecret != "" {
		os.Setenv("GEMINI_OAUTH_CLIENT_SECRET", originalClientSecret)
	}
}

func TestLogoutGeminiOAuth2(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()

	// Save original environment variables
	originalHome := os.Getenv("HOME")

	// Create temporary directory and token file
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ".config", "gemini")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	tokenFile := filepath.Join(configDir, "oauth_creds.json")
	tokenData := oauth.TokenInfo{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}
	data, err := json.Marshal(tokenData)
	require.NoError(t, err)
	err = os.WriteFile(tokenFile, data, 0600)
	require.NoError(t, err)

	// Verify token exists
	assert.FileExists(t, tokenFile)

	// Logout
	err = LogoutGeminiOAuth2()
	assert.NoError(t, err)

	// Verify token was removed
	assert.NoFileExists(t, tokenFile)

	// Restore environment variables
	if originalHome != "" {
		os.Setenv("HOME", originalHome)
	}
}

func TestAuthMethodPriority(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()

	tests := []struct {
		name       string
		authMethod AuthMethod
		hasOAuth   bool
		hasAPIKey  bool
		expected   string
	}{
		{
			name:       "api_key method with API key",
			authMethod: AuthMethodAPIKey,
			hasAPIKey:  true,
			expected:   "API key",
		},
		{
			name:       "oauth2 method with OAuth token",
			authMethod: AuthMethodOAuth2,
			hasOAuth:   true,
			expected:   "OAuth2 token",
		},
		{
			name:       "auto method with both - OAuth preferred",
			authMethod: AuthMethodAuto,
			hasOAuth:   true,
			hasAPIKey:  true,
			expected:   "OAuth2 token",
		},
		{
			name:       "auto method with API key only",
			authMethod: AuthMethodAuto,
			hasAPIKey:  true,
			expected:   "API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment variables
			originalHome := os.Getenv("HOME")
			originalToken := os.Getenv("GEMINI_TOKEN")
			originalAPIKey := os.Getenv("GEMINI_API_KEY")

			// Create temporary directory
			tmpDir := t.TempDir()
			os.Setenv("HOME", tmpDir)
			os.Unsetenv("GEMINI_TOKEN")
			os.Unsetenv("GEMINI_API_KEY")

			// Set up config with auth method
			configData := map[string]interface{}{
				"providers": map[string]interface{}{
					"gemini": map[string]interface{}{
						"authMethod": string(tt.authMethod),
					},
				},
			}

			configFile := filepath.Join(tmpDir, ".opencode.json")
			data, err := json.Marshal(configData)
			require.NoError(t, err)
			err = os.WriteFile(configFile, data, 0600)
			require.NoError(t, err)

			config, err := Load(tmpDir, false)
			require.NoError(t, err)
			cfg = config

			// Set up OAuth token if needed
			if tt.hasOAuth {
				configDir := filepath.Join(tmpDir, ".config", "gemini")
				err := os.MkdirAll(configDir, 0700)
				require.NoError(t, err)

				tokenFile := filepath.Join(configDir, "oauth_creds.json")
				tokenData := oauth.TokenInfo{
					AccessToken: "test-oauth-token",
					TokenType:   "Bearer",
					Expiry:      time.Now().Add(time.Hour),
				}
				tokenFileData, err := json.Marshal(tokenData)
				require.NoError(t, err)
				err = os.WriteFile(tokenFile, tokenFileData, 0600)
				require.NoError(t, err)
			}

			// Set up API key if needed
			if tt.hasAPIKey {
				os.Setenv("GEMINI_API_KEY", "test-api-key")
			}

			// Test status
			status, isValid := GetGeminiAuthStatus()
			assert.True(t, isValid)
			assert.Contains(t, status, tt.expected)

			// Restore environment variables
			if originalHome != "" {
				os.Setenv("HOME", originalHome)
			}
			if originalToken != "" {
				os.Setenv("GEMINI_TOKEN", originalToken)
			}
			if originalAPIKey != "" {
				os.Setenv("GEMINI_API_KEY", originalAPIKey)
			}
		})
	}
}

func TestOAuth2ConfigValidation(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()

	tests := []struct {
		name           string
		clientID       string
		clientSecret   string
		expectValid    bool
	}{
		{
			name:         "valid credentials",
			clientID:     "test.apps.googleusercontent.com",
			clientSecret: "secret123",
			expectValid:  true,
		},
		{
			name:         "empty client ID",
			clientID:     "",
			clientSecret: "secret123",
			expectValid:  false,
		},
		{
			name:         "empty client secret",
			clientID:     "test.apps.googleusercontent.com",
			clientSecret: "",
			expectValid:  false,
		},
		{
			name:         "placeholder values",
			clientID:     "your-client-id.apps.googleusercontent.com",
			clientSecret: "your-client-secret",
			expectValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment variables
			originalClientID := os.Getenv("GEMINI_OAUTH_CLIENT_ID")
			originalClientSecret := os.Getenv("GEMINI_OAUTH_CLIENT_SECRET")

			// Set test environment variables
			if tt.clientID != "" {
				os.Setenv("GEMINI_OAUTH_CLIENT_ID", tt.clientID)
			} else {
				os.Unsetenv("GEMINI_OAUTH_CLIENT_ID")
			}

			if tt.clientSecret != "" {
				os.Setenv("GEMINI_OAUTH_CLIENT_SECRET", tt.clientSecret)
			} else {
				os.Unsetenv("GEMINI_OAUTH_CLIENT_SECRET")
			}

			service := GetOAuth2Service()
			assert.Equal(t, tt.expectValid, service.HasValidCredentials())

			// Restore environment variables
			if originalClientID != "" {
				os.Setenv("GEMINI_OAUTH_CLIENT_ID", originalClientID)
			} else {
				os.Unsetenv("GEMINI_OAUTH_CLIENT_ID")
			}
			if originalClientSecret != "" {
				os.Setenv("GEMINI_OAUTH_CLIENT_SECRET", originalClientSecret)
			} else {
				os.Unsetenv("GEMINI_OAUTH_CLIENT_SECRET")
			}
		})
	}
}

func TestGetGeminiCredentials_Priority(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()

	// Save original environment variables
	originalHome := os.Getenv("HOME")
	originalToken := os.Getenv("GEMINI_TOKEN")
	originalAPIKey := os.Getenv("GEMINI_API_KEY")

	// Create temporary directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create config with API key
	configData := map[string]interface{}{
		"providers": map[string]interface{}{
			"gemini": map[string]interface{}{
				"apiKey": "config-api-key",
			},
		},
	}

	configFile := filepath.Join(tmpDir, ".opencode.json")
	data, err := json.Marshal(configData)
	require.NoError(t, err)
	err = os.WriteFile(configFile, data, 0600)
	require.NoError(t, err)

	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	cfg = config

	// Create OAuth token file
	configDir := filepath.Join(tmpDir, ".config", "gemini")
	err = os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	tokenFile := filepath.Join(configDir, "oauth_creds.json")
	tokenData := oauth.TokenInfo{
		AccessToken: "oauth-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	tokenFileData, err := json.Marshal(tokenData)
	require.NoError(t, err)
	err = os.WriteFile(tokenFile, tokenFileData, 0600)
	require.NoError(t, err)

	// Test priority: GEMINI_API_KEY > GEMINI_TOKEN > OAuth file > config API key

	// 1. Only config API key and OAuth file - OAuth should be preferred (auto mode)
	os.Unsetenv("GEMINI_TOKEN")
	os.Unsetenv("GEMINI_API_KEY")
	status, isValid := GetGeminiAuthStatus()
	assert.True(t, isValid)
	assert.Contains(t, status, "OAuth2 token")

	// 2. Environment token should take precedence over OAuth file
	os.Setenv("GEMINI_TOKEN", "env-token")
	status, isValid = GetGeminiAuthStatus()
	assert.True(t, isValid)
	assert.Contains(t, status, "Environment variable token")

	// 3. API key should take precedence over everything
	os.Setenv("GEMINI_API_KEY", "env-api-key")
	status, isValid = GetGeminiAuthStatus()
	assert.True(t, isValid)
	assert.Contains(t, status, "API key")

	// Restore environment variables
	if originalHome != "" {
		os.Setenv("HOME", originalHome)
	}
	if originalToken != "" {
		os.Setenv("GEMINI_TOKEN", originalToken)
	} else {
		os.Unsetenv("GEMINI_TOKEN")
	}
	if originalAPIKey != "" {
		os.Setenv("GEMINI_API_KEY", originalAPIKey)
	} else {
		os.Unsetenv("GEMINI_API_KEY")
	}
}