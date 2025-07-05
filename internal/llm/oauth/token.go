// Package oauth provides token management utilities
package oauth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

// TokenInfo represents the OAuth2 token information stored in files
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}

// LoadGeminiToken loads a Gemini OAuth2 token from XDG-compliant locations
// This function is designed to be used by the config package
func LoadGeminiToken() (string, error) {
	// First check environment variable
	if token := os.Getenv("GEMINI_TOKEN"); token != "" {
		return token, nil
	}

	// Get token file paths following XDG Base Directory Specification
	paths, err := getGeminiTokenFilePaths()
	if err != nil {
		return "", err
	}

	// Try each path in order
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var tokenData map[string]interface{}
		if err := json.Unmarshal(data, &tokenData); err != nil {
			continue
		}

		// Extract access token from OAuth2 token structure
		if accessToken, ok := tokenData["access_token"].(string); ok && accessToken != "" {
			return accessToken, nil
		}
	}

	return "", fmt.Errorf("Gemini token not found in standard locations")
}

// getGeminiTokenFilePaths returns the ordered list of token file paths
// This matches the XDG specification used in the OAuth service
func getGeminiTokenFilePaths() ([]string, error) {
	var paths []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// XDG_CONFIG_HOME/gemini/oauth_creds.json
	if configDir := os.Getenv("XDG_CONFIG_HOME"); configDir != "" {
		paths = append(paths, filepath.Join(configDir, "gemini", "oauth_creds.json"))
	} else {
		// ~/.config/gemini/oauth_creds.json
		paths = append(paths, filepath.Join(homeDir, ".config", "gemini", "oauth_creds.json"))
	}

	// ~/.gemini/oauth_creds.json (fallback)
	paths = append(paths, filepath.Join(homeDir, ".gemini", "oauth_creds.json"))

	return paths, nil
}

// HasGeminiOAuthToken checks if a valid OAuth2 token exists
func HasGeminiOAuthToken() bool {
	_, err := LoadGeminiToken()
	return err == nil
}

// ConvertToOAuth2Token converts a TokenInfo to oauth2.Token
func (t *TokenInfo) ToOAuth2Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
		Expiry:       t.Expiry,
	}
}

// NewTokenInfoFromOAuth2 creates a TokenInfo from oauth2.Token
func NewTokenInfoFromOAuth2(token *oauth2.Token) *TokenInfo {
	return &TokenInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}
}

// IsValid checks if the token is still valid (not expired)
func (t *TokenInfo) IsValid() bool {
	return time.Now().Before(t.Expiry)
}

// IsExpired checks if the token has expired
func (t *TokenInfo) IsExpired() bool {
	return !t.IsValid()
}

// WillExpireSoon checks if the token will expire within the given duration
func (t *TokenInfo) WillExpireSoon(d time.Duration) bool {
	return time.Now().Add(d).After(t.Expiry)
}