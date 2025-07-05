package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_LoadToken(t *testing.T) {
	service := NewService()

	// Save original environment variables
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	originalToken := os.Getenv("GEMINI_TOKEN")

	// Clean up environment variables
	os.Unsetenv("GEMINI_TOKEN")

	// Create temporary directory structure
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")

	// Restore original environment variables after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		}
		if originalToken != "" {
			os.Setenv("GEMINI_TOKEN", originalToken)
		}
	}()

	// Test case: No token exists
	_, _, err := service.LoadToken()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid token found")

	// Test case: Valid token exists
	configDir := filepath.Join(tmpDir, ".config", "gemini")
	err = os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	tokenFile := filepath.Join(configDir, "oauth_creds.json")
	data, err := json.Marshal(token)
	require.NoError(t, err)
	err = os.WriteFile(tokenFile, data, 0600)
	require.NoError(t, err)

	loadedToken, path, err := service.LoadToken()
	assert.NoError(t, err)
	assert.Equal(t, tokenFile, path)
	if assert.NotNil(t, loadedToken) {
		assert.Equal(t, token.AccessToken, loadedToken.AccessToken)
		assert.Equal(t, token.RefreshToken, loadedToken.RefreshToken)
	}
}

func TestService_SaveToken(t *testing.T) {
	service := NewService()

	// Save original environment variables
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")

	// Create temporary directory structure
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")

	// Restore original environment variables after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		}
	}()

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}

	err := service.SaveToken(token)
	assert.NoError(t, err)

	// Verify token was saved
	expectedPath := filepath.Join(tmpDir, ".config", "gemini", "oauth_creds.json")
	assert.FileExists(t, expectedPath)

	// Verify token content
	data, err := os.ReadFile(expectedPath)
	require.NoError(t, err)

	var savedToken oauth2.Token
	err = json.Unmarshal(data, &savedToken)
	require.NoError(t, err)

	assert.Equal(t, token.AccessToken, savedToken.AccessToken)
	assert.Equal(t, token.RefreshToken, savedToken.RefreshToken)
}

func TestService_ClearToken(t *testing.T) {
	service := NewService()

	// Save original environment variables
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")

	// Create temporary directory structure
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")

	// Restore original environment variables after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		}
	}()

	// Create a token file
	configDir := filepath.Join(tmpDir, ".config", "gemini")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	tokenFile := filepath.Join(configDir, "oauth_creds.json")
	token := &oauth2.Token{AccessToken: "test"}
	data, err := json.Marshal(token)
	require.NoError(t, err)
	err = os.WriteFile(tokenFile, data, 0600)
	require.NoError(t, err)

	// Clear token
	err = service.ClearToken()
	assert.NoError(t, err)

	// Verify token was removed
	assert.NoFileExists(t, tokenFile)
}

func TestLoadGeminiToken(t *testing.T) {
	// Test environment variable
	os.Setenv("GEMINI_TOKEN", "env-token")
	defer os.Unsetenv("GEMINI_TOKEN")

	token, err := LoadGeminiToken()
	assert.NoError(t, err)
	assert.Equal(t, "env-token", token)
}

func TestLoadGeminiToken_FromFile(t *testing.T) {
	// Save original environment variables
	originalToken := os.Getenv("GEMINI_TOKEN")
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")

	// Clean up environment variables
	os.Unsetenv("GEMINI_TOKEN")

	// Create temporary directory structure
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")

	// Create the OAuth credentials file
	configDir := filepath.Join(tmpDir, ".config", "gemini")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	oauthFile := filepath.Join(configDir, "oauth_creds.json")
	oauthData := `{
		"access_token": "oauth-access-token",
		"refresh_token": "oauth-refresh-token",
		"token_type": "Bearer",
		"expiry": "2024-12-31T23:59:59Z"
	}`
	err = os.WriteFile(oauthFile, []byte(oauthData), 0600)
	require.NoError(t, err)

	// Restore original environment variables after test
	defer func() {
		if originalToken != "" {
			os.Setenv("GEMINI_TOKEN", originalToken)
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		}
	}()

	token, err := LoadGeminiToken()
	assert.NoError(t, err)
	assert.Equal(t, "oauth-access-token", token)
}

func TestHasGeminiOAuthToken(t *testing.T) {
	// Save original values
	originalToken := os.Getenv("GEMINI_TOKEN")
	originalHome := os.Getenv("HOME")
	
	defer func() {
		if originalToken != "" {
			os.Setenv("GEMINI_TOKEN", originalToken)
		} else {
			os.Unsetenv("GEMINI_TOKEN")
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	// Test with environment variable
	os.Setenv("GEMINI_TOKEN", "test-token")
	assert.True(t, HasGeminiOAuthToken())

	// Test without token - clear env var and set empty HOME
	os.Unsetenv("GEMINI_TOKEN")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	assert.False(t, HasGeminiOAuthToken())
}

func TestTokenInfo_Conversion(t *testing.T) {
	expiry := time.Now().Add(time.Hour)
	
	// Test OAuth2 to TokenInfo conversion
	oauth2Token := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       expiry,
	}

	tokenInfo := NewTokenInfoFromOAuth2(oauth2Token)
	assert.Equal(t, "access", tokenInfo.AccessToken)
	assert.Equal(t, "refresh", tokenInfo.RefreshToken)
	assert.Equal(t, "Bearer", tokenInfo.TokenType)
	assert.Equal(t, expiry, tokenInfo.Expiry)

	// Test TokenInfo to OAuth2 conversion
	convertedToken := tokenInfo.ToOAuth2Token()
	assert.Equal(t, oauth2Token.AccessToken, convertedToken.AccessToken)
	assert.Equal(t, oauth2Token.RefreshToken, convertedToken.RefreshToken)
	assert.Equal(t, oauth2Token.TokenType, convertedToken.TokenType)
	assert.Equal(t, oauth2Token.Expiry, convertedToken.Expiry)
}

func TestTokenInfo_Validity(t *testing.T) {
	// Valid token
	validToken := &TokenInfo{
		AccessToken: "test",
		Expiry:      time.Now().Add(time.Hour),
	}
	assert.True(t, validToken.IsValid())
	assert.False(t, validToken.IsExpired())
	assert.False(t, validToken.WillExpireSoon(30*time.Minute))
	assert.True(t, validToken.WillExpireSoon(2*time.Hour))

	// Expired token
	expiredToken := &TokenInfo{
		AccessToken: "test",
		Expiry:      time.Now().Add(-time.Hour),
	}
	assert.False(t, expiredToken.IsValid())
	assert.True(t, expiredToken.IsExpired())
	assert.True(t, expiredToken.WillExpireSoon(time.Hour))
}

// Additional comprehensive tests

func TestNewServiceWithCredentials(t *testing.T) {
	clientID := "test-client-id.apps.googleusercontent.com"
	clientSecret := "test-client-secret"
	
	service := NewServiceWithCredentials(clientID, clientSecret)
	
	assert.NotNil(t, service)
	assert.Equal(t, clientID, service.clientID)
	assert.Equal(t, clientSecret, service.clientSecret)
	assert.NotNil(t, service.config)
	assert.Equal(t, clientID, service.config.ClientID)
	assert.Equal(t, clientSecret, service.config.ClientSecret)
}

func TestService_HasValidCredentials(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		expected     bool
	}{
		{
			name:         "valid credentials",
			clientID:     "test-client-id.apps.googleusercontent.com",
			clientSecret: "test-client-secret",
			expected:     true,
		},
		{
			name:         "empty client ID",
			clientID:     "",
			clientSecret: "test-client-secret",
			expected:     false,
		},
		{
			name:         "empty client secret",
			clientID:     "test-client-id.apps.googleusercontent.com",
			clientSecret: "",
			expected:     false,
		},
		{
			name:         "placeholder client ID",
			clientID:     "your-client-id.apps.googleusercontent.com",
			clientSecret: "test-client-secret",
			expected:     false,
		},
		{
			name:         "placeholder client secret",
			clientID:     "test-client-id.apps.googleusercontent.com",
			clientSecret: "your-client-secret",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewServiceWithCredentials(tt.clientID, tt.clientSecret)
			assert.Equal(t, tt.expected, service.HasValidCredentials())
		})
	}
}

func TestService_XDGPathPriority(t *testing.T) {
	// Save original environment variables
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	
	// Create temporary directory structure
	tmpDir := t.TempDir()
	xdgDir := filepath.Join(tmpDir, "xdg")
	
	// Set up environment
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CONFIG_HOME", xdgDir)
	
	// Restore original environment variables after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()
	
	service := NewService()
	
	// Create tokens in different locations
	homeToken := &oauth2.Token{AccessToken: "home-token", TokenType: "Bearer"}
	xdgToken := &oauth2.Token{AccessToken: "xdg-token", TokenType: "Bearer"}
	
	// Create home config directory and token
	homeConfigDir := filepath.Join(tmpDir, ".config", "gemini")
	err := os.MkdirAll(homeConfigDir, 0700)
	require.NoError(t, err)
	
	homeTokenFile := filepath.Join(homeConfigDir, "oauth_creds.json")
	homeData, err := json.Marshal(homeToken)
	require.NoError(t, err)
	err = os.WriteFile(homeTokenFile, homeData, 0600)
	require.NoError(t, err)
	
	// Create XDG config directory and token
	xdgConfigDir := filepath.Join(xdgDir, "gemini")
	err = os.MkdirAll(xdgConfigDir, 0700)
	require.NoError(t, err)
	
	xdgTokenFile := filepath.Join(xdgConfigDir, "oauth_creds.json")
	xdgData, err := json.Marshal(xdgToken)
	require.NoError(t, err)
	err = os.WriteFile(xdgTokenFile, xdgData, 0600)
	require.NoError(t, err)
	
	// XDG_CONFIG_HOME should take precedence
	loadedToken, path, err := service.LoadToken()
	assert.NoError(t, err)
	assert.Equal(t, xdgTokenFile, path)
	assert.Equal(t, "xdg-token", loadedToken.AccessToken)
}

func TestService_MultipleTokenLocations(t *testing.T) {
	// Save original environment variables
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	
	// Create temporary directory structure
	tmpDir := t.TempDir()
	
	// Set up environment (no XDG_CONFIG_HOME)
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")
	
	// Restore original environment variables after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		}
	}()
	
	service := NewService()
	
	// Create tokens in both standard and fallback locations
	standardToken := &oauth2.Token{AccessToken: "standard-token", TokenType: "Bearer"}
	fallbackToken := &oauth2.Token{AccessToken: "fallback-token", TokenType: "Bearer"}
	
	// Create standard config directory and token
	standardConfigDir := filepath.Join(tmpDir, ".config", "gemini")
	err := os.MkdirAll(standardConfigDir, 0700)
	require.NoError(t, err)
	
	standardTokenFile := filepath.Join(standardConfigDir, "oauth_creds.json")
	standardData, err := json.Marshal(standardToken)
	require.NoError(t, err)
	err = os.WriteFile(standardTokenFile, standardData, 0600)
	require.NoError(t, err)
	
	// Create fallback config directory and token
	fallbackConfigDir := filepath.Join(tmpDir, ".gemini")
	err = os.MkdirAll(fallbackConfigDir, 0700)
	require.NoError(t, err)
	
	fallbackTokenFile := filepath.Join(fallbackConfigDir, "oauth_creds.json")
	fallbackData, err := json.Marshal(fallbackToken)
	require.NoError(t, err)
	err = os.WriteFile(fallbackTokenFile, fallbackData, 0600)
	require.NoError(t, err)
	
	// Standard location should take precedence
	loadedToken, path, err := service.LoadToken()
	assert.NoError(t, err)
	assert.Equal(t, standardTokenFile, path)
	assert.Equal(t, "standard-token", loadedToken.AccessToken)
}

func TestService_InvalidCredentialsError(t *testing.T) {
	service := NewService() // Uses placeholder credentials
	ctx := context.Background()
	
	_, err := service.Login(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OAuth2 credentials not configured")
}

func TestService_SaveTokenPermissions(t *testing.T) {
	// Save original environment variables
	originalHome := os.Getenv("HOME")
	
	// Create temporary directory structure
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	
	// Restore original environment variables after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()
	
	service := NewService()
	
	token := &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}
	
	err := service.SaveToken(token)
	assert.NoError(t, err)
	
	// Check file permissions
	tokenFile := filepath.Join(tmpDir, ".config", "gemini", "oauth_creds.json")
	info, err := os.Stat(tokenFile)
	require.NoError(t, err)
	
	// File should be readable/writable by owner only (0600)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestService_ClearMultipleTokens(t *testing.T) {
	// Save original environment variables
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	
	// Create temporary directory structure
	tmpDir := t.TempDir()
	xdgDir := filepath.Join(tmpDir, "xdg")
	
	// Set up environment
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CONFIG_HOME", xdgDir)
	
	// Restore original environment variables after test
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()
	
	service := NewService()
	
	// Create token files in multiple locations
	locations := []string{
		filepath.Join(xdgDir, "gemini", "oauth_creds.json"),
		filepath.Join(tmpDir, ".config", "gemini", "oauth_creds.json"),
		filepath.Join(tmpDir, ".gemini", "oauth_creds.json"),
	}
	
	token := &oauth2.Token{AccessToken: "test", TokenType: "Bearer"}
	tokenData, _ := json.Marshal(token)
	
	for _, location := range locations {
		err := os.MkdirAll(filepath.Dir(location), 0700)
		require.NoError(t, err)
		err = os.WriteFile(location, tokenData, 0600)
		require.NoError(t, err)
	}
	
	// Clear tokens
	err := service.ClearToken()
	assert.NoError(t, err)
	
	// Verify all tokens were removed
	for _, location := range locations {
		assert.NoFileExists(t, location)
	}
}

func TestGetTokenFilePaths(t *testing.T) {
	// Save original environment variables
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	
	// Create temporary directory
	tmpDir := t.TempDir()
	xdgDir := filepath.Join(tmpDir, "xdg")
	
	// Test with XDG_CONFIG_HOME set
	os.Setenv("HOME", tmpDir)
	os.Setenv("XDG_CONFIG_HOME", xdgDir)
	
	paths, err := getTokenFilePaths()
	assert.NoError(t, err)
	assert.Len(t, paths, 2)
	assert.Equal(t, filepath.Join(xdgDir, "gemini", "oauth_creds.json"), paths[0])
	assert.Equal(t, filepath.Join(tmpDir, ".gemini", "oauth_creds.json"), paths[1])
	
	// Test without XDG_CONFIG_HOME
	os.Unsetenv("XDG_CONFIG_HOME")
	
	paths, err = getTokenFilePaths()
	assert.NoError(t, err)
	assert.Len(t, paths, 2)
	assert.Equal(t, filepath.Join(tmpDir, ".config", "gemini", "oauth_creds.json"), paths[0])
	assert.Equal(t, filepath.Join(tmpDir, ".gemini", "oauth_creds.json"), paths[1])
	
	// Restore original environment variables
	if originalHome != "" {
		os.Setenv("HOME", originalHome)
	} else {
		os.Unsetenv("HOME")
	}
	if originalXDG != "" {
		os.Setenv("XDG_CONFIG_HOME", originalXDG)
	}
}

func TestLoadGeminiToken_Priority(t *testing.T) {
	// Save original environment variables
	originalToken := os.Getenv("GEMINI_TOKEN")
	originalHome := os.Getenv("HOME")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	
	// Create temporary directory
	tmpDir := t.TempDir()
	
	// Set up environment
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")
	
	// Create OAuth token file
	configDir := filepath.Join(tmpDir, ".config", "gemini")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)
	
	oauthFile := filepath.Join(configDir, "oauth_creds.json")
	oauthData := `{
		"access_token": "file-token",
		"token_type": "Bearer"
	}`
	err = os.WriteFile(oauthFile, []byte(oauthData), 0600)
	require.NoError(t, err)
	
	// Restore original environment variables after test
	defer func() {
		if originalToken != "" {
			os.Setenv("GEMINI_TOKEN", originalToken)
		} else {
			os.Unsetenv("GEMINI_TOKEN")
		}
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		}
	}()
	
	// Test without environment variable - should use file
	os.Unsetenv("GEMINI_TOKEN")
	token, err := LoadGeminiToken()
	assert.NoError(t, err)
	assert.Equal(t, "file-token", token)
	
	// Test with environment variable - should take precedence
	os.Setenv("GEMINI_TOKEN", "env-token")
	token, err = LoadGeminiToken()
	assert.NoError(t, err)
	assert.Equal(t, "env-token", token)
}

func TestRandomPort(t *testing.T) {
	port1 := randomPort()
	port2 := randomPort()
	
	// Ports should be in ephemeral range
	assert.GreaterOrEqual(t, port1, 49152)
	assert.LessOrEqual(t, port1, 65535)
	assert.GreaterOrEqual(t, port2, 49152)
	assert.LessOrEqual(t, port2, 65535)
	
	// Multiple calls should potentially return different ports
	// (though this test might occasionally fail due to randomness)
	// We'll just ensure they're in the valid range
}

func TestRandomState(t *testing.T) {
	state1 := randomState()
	state2 := randomState()
	
	// States should be non-empty hex strings
	assert.NotEmpty(t, state1)
	assert.NotEmpty(t, state2)
	assert.Len(t, state1, 32) // 16 bytes * 2 hex chars
	assert.Len(t, state2, 32)
	
	// States should be different
	assert.NotEqual(t, state1, state2)
	
	// Should only contain hex characters
	for _, r := range state1 {
		assert.True(t, (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'))
	}
}

// Mock OAuth2 server for testing
func setupMockOAuth2Server(t *testing.T) (*httptest.Server, string) {
	mux := http.NewServeMux()
	
	// Authorization endpoint
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		clientID := query.Get("client_id")
		redirectURI := query.Get("redirect_uri")
		state := query.Get("state")
		
		if clientID == "" || redirectURI == "" || state == "" {
			http.Error(w, "Missing required parameters", http.StatusBadRequest)
			return
		}
		
		// Simulate user authorization by redirecting with auth code
		authCode := "test-auth-code"
		redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, authCode, state)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	})
	
	// Token endpoint
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		
		grantType := r.FormValue("grant_type")
		code := r.FormValue("code")
		
		if grantType != "authorization_code" || code != "test-auth-code" {
			http.Error(w, "Invalid grant", http.StatusBadRequest)
			return
		}
		
		// Return mock token response
		response := map[string]interface{}{
			"access_token":  "mock-access-token",
			"refresh_token": "mock-refresh-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	server := httptest.NewServer(mux)
	return server, server.URL
}

// Note: Full OAuth2 flow testing would require more complex setup
// with actual browser automation or mocking the entire flow.
// These tests cover the core components and edge cases.