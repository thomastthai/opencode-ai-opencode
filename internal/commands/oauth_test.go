package commands

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/llm/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuth2Commands_Registration(t *testing.T) {
	// Get the global registry (built-in commands are loaded during init)
	registry := GetGlobalRegistry()

	// Check that OAuth2 commands are registered
	loginCmd, exists := registry.Get("login")
	assert.True(t, exists)
	assert.NotNil(t, loginCmd)
	assert.Equal(t, "Login", loginCmd.Name())
	assert.Equal(t, "Authenticate with OAuth2 providers", loginCmd.Description())

	// Check subcommands
	subCommands := loginCmd.GetSubCommands()
	assert.NotEmpty(t, subCommands)

	var geminiLoginCmd Command
	for _, sub := range subCommands {
		if sub.ID() == "gemini" {
			geminiLoginCmd = sub
			break
		}
	}
	assert.NotNil(t, geminiLoginCmd)
	assert.Equal(t, "Login Gemini", geminiLoginCmd.Name())
	assert.Equal(t, "Login to Gemini with OAuth2", geminiLoginCmd.Description())

	// Check logout commands
	logoutCmd, exists := registry.Get("logout")
	assert.True(t, exists)
	assert.NotNil(t, logoutCmd)
	assert.Equal(t, "Logout", logoutCmd.Name())

	// Check auth commands
	authCmd, exists := registry.Get("auth")
	assert.True(t, exists)
	assert.NotNil(t, authCmd)
	assert.Equal(t, "Auth Status", authCmd.Name())
}

func TestHandleLoginGemini_NoCredentials(t *testing.T) {
	// Save original environment variables
	originalClientID := os.Getenv("GEMINI_OAUTH_CLIENT_ID")
	originalClientSecret := os.Getenv("GEMINI_OAUTH_CLIENT_SECRET")

	// Clear environment variables
	os.Unsetenv("GEMINI_OAUTH_CLIENT_ID")
	os.Unsetenv("GEMINI_OAUTH_CLIENT_SECRET")

	defer func() {
		if originalClientID != "" {
			os.Setenv("GEMINI_OAUTH_CLIENT_ID", originalClientID)
		}
		if originalClientSecret != "" {
			os.Setenv("GEMINI_OAUTH_CLIENT_SECRET", originalClientSecret)
		}
	}()

	ctx := context.Background()
	err := handleLoginGemini(ctx, map[string]interface{}{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth2 credentials not configured")
}

func TestHandleLoginGemini_WithCredentials(t *testing.T) {
	t.Skip("Skipping test that requires browser interaction")
}

func TestHandleLogoutGemini(t *testing.T) {
	// Call logout handler - it should work even if no token exists
	ctx := context.Background()
	err := handleLogoutGemini(ctx, map[string]interface{}{})
	
	// The logout should succeed even if there's no token to remove
	assert.NoError(t, err)
}

func TestHandleAuthStatus(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalToken := os.Getenv("GEMINI_TOKEN")
	originalAPIKey := os.Getenv("GEMINI_API_KEY")

	os.Setenv("HOME", tmpDir)
	os.Unsetenv("GEMINI_TOKEN")
	os.Unsetenv("GEMINI_API_KEY")

	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		if originalToken != "" {
			os.Setenv("GEMINI_TOKEN", originalToken)
		}
		if originalAPIKey != "" {
			os.Setenv("GEMINI_API_KEY", originalAPIKey)
		}
	}()

	ctx := context.Background()

	// Test with no authentication
	err := handleAuthStatus(ctx, map[string]interface{}{})
	assert.NoError(t, err)

	// Test with environment token
	os.Setenv("GEMINI_TOKEN", "test-token")
	err = handleAuthStatus(ctx, map[string]interface{}{})
	assert.NoError(t, err)

	// Test with OAuth token file
	os.Unsetenv("GEMINI_TOKEN")
	configDir := filepath.Join(tmpDir, ".config", "gemini")
	err = os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	tokenFile := filepath.Join(configDir, "oauth_creds.json")
	tokenData := oauth.TokenInfo{
		AccessToken: "oauth-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	data, err := json.Marshal(tokenData)
	require.NoError(t, err)
	err = os.WriteFile(tokenFile, data, 0600)
	require.NoError(t, err)

	err = handleAuthStatus(ctx, map[string]interface{}{})
	assert.NoError(t, err)
}

func TestHandleAuthMethod(t *testing.T) {
	ctx := context.Background()
	err := handleAuthMethod(ctx, map[string]interface{}{})
	assert.NoError(t, err)
}

func TestOAuth2CommandsHelp(t *testing.T) {
	// Test that OAuth2 commands provide helpful error messages
	
	// Save original environment variables
	originalClientID := os.Getenv("GEMINI_OAUTH_CLIENT_ID")
	originalClientSecret := os.Getenv("GEMINI_OAUTH_CLIENT_SECRET")

	// Clear environment variables
	os.Unsetenv("GEMINI_OAUTH_CLIENT_ID")
	os.Unsetenv("GEMINI_OAUTH_CLIENT_SECRET")

	defer func() {
		if originalClientID != "" {
			os.Setenv("GEMINI_OAUTH_CLIENT_ID", originalClientID)
		}
		if originalClientSecret != "" {
			os.Setenv("GEMINI_OAUTH_CLIENT_SECRET", originalClientSecret)
		}
	}()

	ctx := context.Background()
	err := handleLoginGemini(ctx, map[string]interface{}{})

	assert.Error(t, err)
	
	// The error should contain helpful setup instructions
	errMsg := err.Error()
	assert.Contains(t, errMsg, "OAuth2 credentials not configured")
	
	// We can't easily test the printed output, but we can verify
	// that the function completes and returns the expected error
}

func TestOAuth2CommandsWithConfig(t *testing.T) {
	t.Skip("Skipping test that requires browser interaction")
}

func TestOAuth2Commands_ErrorMessages(t *testing.T) {
	// Test that error messages are user-friendly and informative
	
	tests := []struct {
		name               string
		clientID           string
		clientSecret       string
		expectedErrorParts []string
	}{
		{
			name:         "no credentials",
			clientID:     "",
			clientSecret: "",
			expectedErrorParts: []string{
				"OAuth2 credentials not configured",
			},
		},
		{
			name:         "placeholder credentials",
			clientID:     "your-client-id.apps.googleusercontent.com",
			clientSecret: "your-client-secret",
			expectedErrorParts: []string{
				"OAuth2 credentials not configured",
			},
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

			defer func() {
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
			}()

			ctx := context.Background()
			err := handleLoginGemini(ctx, map[string]interface{}{})

			// With invalid credentials, it should always fail
			if tt.clientID == "" || tt.clientSecret == "" || 
			   tt.clientID == "your-client-id.apps.googleusercontent.com" ||
			   tt.clientSecret == "your-client-secret" {
				assert.Error(t, err)
				if err != nil {
					errMsg := err.Error()
					for _, expectedPart := range tt.expectedErrorParts {
						assert.Contains(t, errMsg, expectedPart)
					}
				}
			} else {
				// With valid test credentials, it might succeed or fail
				if err != nil {
					// If it fails, check that it's not a credentials error
					errMsg := err.Error()
					for _, expectedPart := range tt.expectedErrorParts {
						if expectedPart == "OAuth2 credentials not configured" {
							// This error should not occur with valid credentials
							assert.NotContains(t, errMsg, expectedPart)
						} else {
							assert.Contains(t, errMsg, expectedPart)
						}
					}
				}
			}
		})
	}
}

func TestOAuth2Commands_TokenFileIntegration(t *testing.T) {
	// Test auth status and logout handlers
	ctx := context.Background()

	// Test auth status - should work even with no auth configured
	err := handleAuthStatus(ctx, map[string]interface{}{})
	assert.NoError(t, err)

	// Test logout - should work even if no token exists
	err = handleLogoutGemini(ctx, map[string]interface{}{})
	assert.NoError(t, err)

	// Test auth method
	err = handleAuthMethod(ctx, map[string]interface{}{})
	assert.NoError(t, err)
}

// Helper function to check if config package has the needed functions
// This is to ensure the test requirements are met
func TestConfigIntegrationRequirements(t *testing.T) {
	// These functions should exist in the config package
	// and be accessible from the commands package
	
	service := config.GetOAuth2Service()
	assert.NotNil(t, service)
	
	method := config.GetGeminiAuthMethod()
	assert.NotEmpty(t, method)
	
	status, _ := config.GetGeminiAuthStatus()
	assert.NotEmpty(t, status)
	
	// These functions should exist for command handlers to work
	ctx := context.Background()
	
	// LoginWithGeminiOAuth2 should exist (will fail without credentials)
	err := config.LoginWithGeminiOAuth2(ctx)
	assert.Error(t, err) // Expected to fail without valid credentials
	
	// LogoutGeminiOAuth2 should exist
	err = config.LogoutGeminiOAuth2()
	assert.NoError(t, err) // Should succeed even with no tokens
}