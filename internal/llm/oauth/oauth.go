// Package oauth provides OAuth2 authentication services for OpenCode
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/pkg/browser"
)

const (
	tokenFileName = "oauth_creds.json"
)

var oauthScopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
}

// Service provides OAuth2 authentication services
type Service struct {
	config   *oauth2.Config
	clientID string
	clientSecret string
}

// NewService creates a new OAuth2 service with default credentials
func NewService() *Service {
	return NewServiceWithCredentials("your-client-id.apps.googleusercontent.com", "your-client-secret")
}

// NewServiceWithCredentials creates a new OAuth2 service with custom credentials
func NewServiceWithCredentials(clientID, clientSecret string) *Service {
	return &Service{
		clientID:     clientID,
		clientSecret: clientSecret,
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       oauthScopes,
			Endpoint:     google.Endpoint,
		},
	}
}

// HasValidCredentials checks if the service has valid OAuth2 credentials configured
func (s *Service) HasValidCredentials() bool {
	return s.clientID != "" && 
		   s.clientSecret != "" && 
		   s.clientID != "your-client-id.apps.googleusercontent.com" &&
		   s.clientSecret != "your-client-secret"
}

// Login performs the OAuth2 login flow
func (s *Service) Login(ctx context.Context) (*oauth2.Token, error) {
	// Check if valid credentials are configured
	if !s.HasValidCredentials() {
		return nil, fmt.Errorf("OAuth2 credentials not configured. Please set GEMINI_OAUTH_CLIENT_ID and GEMINI_OAUTH_CLIENT_SECRET environment variables or configure them in your config file")
	}

	// Check if we already have a valid token
	if token, _, err := s.LoadToken(); err == nil && token.Valid() {
		slog.Info("Using existing valid token")
		return token, nil
	}

	// Start OAuth2 flow
	slog.Info("Starting OAuth2 login flow")
	return s.performOAuthFlow(ctx)
}

// LoadToken loads an existing token from XDG-compliant locations
func (s *Service) LoadToken() (*oauth2.Token, string, error) {
	paths, err := getTokenFilePaths()
	if err != nil {
		return nil, "", err
	}

	for _, path := range paths {
		f, err := os.Open(path)
		if err == nil {
			defer f.Close()
			var token oauth2.Token
			if err := json.NewDecoder(f).Decode(&token); err == nil {
				slog.Debug("Found token", "path", path)
				return &token, path, nil
			}
		}
	}

	return nil, "", fmt.Errorf("no valid token found in known locations")
}

// SaveToken saves a token to the first available XDG-compliant location
func (s *Service) SaveToken(token *oauth2.Token) error {
	paths, err := getTokenFilePaths()
	if err != nil {
		return err
	}

	for _, path := range paths {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0700); err != nil {
			continue
		}

		f, err := os.Create(path)
		if err == nil {
			defer f.Close()
			slog.Info("Saved token", "path", path)
			return json.NewEncoder(f).Encode(token)
		}
	}

	return fmt.Errorf("failed to save token in all known locations")
}

// ClearToken removes stored tokens from all known locations
func (s *Service) ClearToken() error {
	paths, err := getTokenFilePaths()
	if err != nil {
		return err
	}

	var lastErr error
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			lastErr = err
		} else if err == nil {
			slog.Info("Removed token", "path", path)
		}
	}

	return lastErr
}

// GetAuthenticatedClient returns an HTTP client with OAuth2 authentication
func (s *Service) GetAuthenticatedClient(ctx context.Context) (*http.Client, error) {
	token, err := s.Login(ctx)
	if err != nil {
		return nil, err
	}

	return s.config.Client(ctx, token), nil
}

// RefreshToken refreshes an expired token
func (s *Service) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	tokenSource := s.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Save the refreshed token
	if err := s.SaveToken(newToken); err != nil {
		slog.Warn("Failed to save refreshed token", "error", err)
	}

	return newToken, nil
}

// performOAuthFlow executes the OAuth2 web flow
func (s *Service) performOAuthFlow(ctx context.Context) (*oauth2.Token, error) {
	port := randomPort()
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)
	s.config.RedirectURL = redirectURL

	state := randomState()
	codeCh := make(chan string)
	errCh := make(chan error)

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "State parameter mismatch", http.StatusBadRequest)
			errCh <- fmt.Errorf("state mismatch")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No authorization code received", http.StatusBadRequest)
			errCh <- fmt.Errorf("no authorization code received")
			return
		}

		fmt.Fprintln(w, "Authentication successful! You can close this window.")
		codeCh <- code
	})

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", server.Addr, err)
	}
	defer ln.Close()

	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := s.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	slog.Info("Opening browser for OAuth2 authentication", "url", authURL)

	if err := openBrowser(authURL); err != nil {
		slog.Warn("Failed to open browser automatically", "error", err)
		fmt.Printf("Please open the following URL manually in your browser:\n%s\n", authURL)
	}

	var code string
	select {
	case code = <-codeCh:
		slog.Info("Received authorization code from browser callback")
	case err := <-errCh:
		return nil, fmt.Errorf("OAuth error: %w", err)
	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("timeout waiting for OAuth2 callback")
	}

	server.Shutdown(ctx)

	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	// Save the token
	if err := s.SaveToken(token); err != nil {
		slog.Warn("Failed to save token", "error", err)
	}

	slog.Info("OAuth2 login successful")
	return token, nil
}

// getTokenFilePaths returns the ordered list of possible token file paths per XDG + fallback
func getTokenFilePaths() ([]string, error) {
	var paths []string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// XDG_CONFIG_HOME/gemini/oauth_creds.json
	if configDir := os.Getenv("XDG_CONFIG_HOME"); configDir != "" {
		paths = append(paths, filepath.Join(configDir, "gemini", tokenFileName))
	} else {
		// ~/.config/gemini/oauth_creds.json
		paths = append(paths, filepath.Join(homeDir, ".config", "gemini", tokenFileName))
	}

	// ~/.gemini/oauth_creds.json (fallback)
	paths = append(paths, filepath.Join(homeDir, ".gemini", tokenFileName))

	return paths, nil
}

// openBrowser attempts to open the given URL in the user's default browser
func openBrowser(url string) error {
	return browser.OpenURL(url)
}

// randomPort returns a random port in the ephemeral range
func randomPort() int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(65535-49152))
	if err != nil {
		return 54321 // fallback
	}
	return int(nBig.Int64()) + 49152
}

// randomState generates a random state parameter for CSRF protection
func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}