package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOAuth2DialogCmp(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	assert.NotNil(t, dialog)
	
	// Cast to concrete type to test internal state
	impl := dialog.(*oauthDialogCmp)
	assert.Equal(t, OAuth2StateIdle, impl.state)
	assert.NotNil(t, impl.spinner)
}

func TestOAuth2Dialog_Init(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	cmd := dialog.Init()
	assert.NotNil(t, cmd)
	
	// Should return spinner tick command
	msg := cmd()
	assert.NotNil(t, msg)
}

func TestOAuth2Dialog_WindowSize(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	windowMsg := tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	}
	
	model, cmd := dialog.Update(windowMsg)
	assert.NotNil(t, model)
	assert.Nil(t, cmd)
	
	// Cast to check size was set
	impl := model.(*oauthDialogCmp)
	assert.Equal(t, 60, impl.width)  // 60% of 100
	assert.Equal(t, 20, impl.height) // 40% of 50
}

func TestOAuth2Dialog_SetSize(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	impl := dialog.(*oauthDialogCmp)
	impl.windowSize = tea.WindowSizeMsg{Width: 200, Height: 100}
	impl.SetSize()
	
	assert.Equal(t, 120, impl.width)  // 60% of 200
	assert.Equal(t, 40, impl.height)  // 40% of 100
	
	// Test minimum size constraints
	impl.windowSize = tea.WindowSizeMsg{Width: 50, Height: 20}
	impl.SetSize()
	
	assert.Equal(t, 50, impl.width)  // Minimum width
	assert.Equal(t, 12, impl.height) // Minimum height
}

func TestOAuth2Dialog_StartOAuth2(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	cmd := dialog.StartOAuth2(OAuth2ProviderGemini)
	assert.NotNil(t, cmd)
	
	// Cast to check state was set
	impl := dialog.(*oauthDialogCmp)
	assert.Equal(t, OAuth2StateAuthenticating, impl.state)
	assert.Equal(t, OAuth2ProviderGemini, impl.provider)
	assert.Equal(t, "Starting OAuth2 authentication...", impl.message)
	assert.Nil(t, impl.error)
}

func TestOAuth2Dialog_OAuth2DialogMsg(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	// Test success message
	successMsg := OAuth2DialogMsg{
		Provider: OAuth2ProviderGemini,
		State:    OAuth2StateSuccess,
		Message:  "Authentication successful!",
		Error:    nil,
	}
	
	model, cmd := dialog.Update(successMsg)
	assert.NotNil(t, model)
	assert.Nil(t, cmd)
	
	impl := model.(*oauthDialogCmp)
	assert.Equal(t, OAuth2StateSuccess, impl.state)
	assert.Equal(t, OAuth2ProviderGemini, impl.provider)
	assert.Equal(t, "Authentication successful!", impl.message)
	assert.Nil(t, impl.error)
	
	// Test error message
	errorMsg := OAuth2DialogMsg{
		Provider: OAuth2ProviderGemini,
		State:    OAuth2StateError,
		Message:  "Authentication failed",
		Error:    assert.AnError,
	}
	
	model, cmd = dialog.Update(errorMsg)
	assert.NotNil(t, model)
	assert.Nil(t, cmd)
	
	impl = model.(*oauthDialogCmp)
	assert.Equal(t, OAuth2StateError, impl.state)
	assert.Equal(t, "Authentication failed", impl.message)
	assert.Equal(t, assert.AnError, impl.error)
}

func TestOAuth2Dialog_KeyHandling(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	// Test close key when not authenticating
	impl := dialog.(*oauthDialogCmp)
	impl.state = OAuth2StateIdle
	
	escapeKey := tea.KeyMsg{Type: tea.KeyEsc}
	model, cmd := dialog.Update(escapeKey)
	assert.NotNil(t, model)
	assert.NotNil(t, cmd)
	
	// Should send close message
	msg := cmd()
	assert.Equal(t, "close_oauth2_dialog", msg)
	
	// Test close key blocked when authenticating
	impl.state = OAuth2StateAuthenticating
	model, cmd = dialog.Update(escapeKey)
	assert.NotNil(t, model)
	assert.Nil(t, cmd)
	
	// Test retry key when in error state
	impl.state = OAuth2StateError
	retryKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	model, cmd = dialog.Update(retryKey)
	assert.NotNil(t, model)
	assert.NotNil(t, cmd)
}

func TestOAuth2Dialog_View(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	// Set up dialog
	impl := dialog.(*oauthDialogCmp)
	impl.width = 80
	impl.height = 20
	impl.provider = OAuth2ProviderGemini
	
	// Test idle state
	impl.state = OAuth2StateIdle
	view := dialog.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Gemini OAuth2 Authentication")
	assert.Contains(t, view, "Ready to authenticate")
	
	// Test authenticating state
	impl.state = OAuth2StateAuthenticating
	impl.message = "Starting authentication..."
	view = dialog.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Starting authentication...")
	assert.Contains(t, view, "browser should open")
	
	// Test success state
	impl.state = OAuth2StateSuccess
	impl.message = "Authentication successful!"
	view = dialog.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Authentication successful!")
	assert.Contains(t, view, "✓")
	
	// Test error state
	impl.state = OAuth2StateError
	impl.message = "Authentication failed"
	impl.error = assert.AnError
	view = dialog.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Authentication failed")
	assert.Contains(t, view, "✗")
	assert.Contains(t, view, "Press 'r' to retry")
}

func TestOAuth2Dialog_BindingKeys(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	keys := dialog.BindingKeys()
	assert.NotEmpty(t, keys)
	
	// Should include close and retry keys
	keyStrings := make([]string, len(keys))
	for i, key := range keys {
		keyStrings[i] = key.Help().Key
	}
	
	assert.Contains(t, keyStrings, "esc/enter")
	assert.Contains(t, keyStrings, "r")
}

func TestOAuth2Dialog_SetState(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	impl := dialog.(*oauthDialogCmp)
	impl.provider = OAuth2ProviderGemini
	
	cmd := dialog.SetState(OAuth2StateSuccess, "Test message", nil)
	assert.NotNil(t, cmd)
	
	msg := cmd()
	require.IsType(t, OAuth2DialogMsg{}, msg)
	
	dialogMsg := msg.(OAuth2DialogMsg)
	assert.Equal(t, OAuth2ProviderGemini, dialogMsg.Provider)
	assert.Equal(t, OAuth2StateSuccess, dialogMsg.State)
	assert.Equal(t, "Test message", dialogMsg.Message)
	assert.Nil(t, dialogMsg.Error)
}

func TestOAuth2Dialog_ProviderName(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	impl := dialog.(*oauthDialogCmp)
	impl.provider = OAuth2ProviderGemini
	impl.width = 80
	impl.height = 20
	
	view := dialog.View()
	
	// Should capitalize provider name properly
	assert.Contains(t, view, "Gemini OAuth2 Authentication")
}

func TestOAuth2Dialog_ErrorHandling(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	impl := dialog.(*oauthDialogCmp)
	impl.state = OAuth2StateError
	impl.provider = OAuth2ProviderGemini
	impl.message = "Test error"
	impl.error = assert.AnError
	impl.width = 80
	impl.height = 20
	
	view := dialog.View()
	
	assert.Contains(t, view, "✗")
	assert.Contains(t, view, "Test error")
	assert.Contains(t, view, "Error:")
	assert.Contains(t, view, "Press 'r' to retry")
}

func TestOAuth2Dialog_SpinnerUpdate(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	impl := dialog.(*oauthDialogCmp)
	impl.state = OAuth2StateAuthenticating
	
	// Create a mock spinner tick message
	// Note: This is a simplified test as spinner.TickMsg is not easily mockable
	type mockTickMsg struct{}
	
	// The dialog should handle unknown message types gracefully
	model, cmd := dialog.Update(mockTickMsg{})
	assert.NotNil(t, model)
	assert.Nil(t, cmd)
}

func TestShowOAuth2DialogMsg(t *testing.T) {
	msg := ShowOAuth2DialogMsg{
		Provider: OAuth2ProviderGemini,
	}
	
	assert.Equal(t, OAuth2ProviderGemini, msg.Provider)
}

func TestOAuth2States(t *testing.T) {
	// Test that OAuth2 state constants are properly defined
	assert.Equal(t, OAuth2State("idle"), OAuth2StateIdle)
	assert.Equal(t, OAuth2State("authenticating"), OAuth2StateAuthenticating)
	assert.Equal(t, OAuth2State("success"), OAuth2StateSuccess)
	assert.Equal(t, OAuth2State("error"), OAuth2StateError)
}

func TestOAuth2Providers(t *testing.T) {
	// Test that OAuth2 provider constants are properly defined
	assert.Equal(t, OAuth2Provider("gemini"), OAuth2ProviderGemini)
}

// Integration test for complete dialog flow
func TestOAuth2Dialog_CompleteFlow(t *testing.T) {
	dialog := NewOAuth2DialogCmp()
	
	// 1. Initialize dialog
	cmd := dialog.Init()
	assert.NotNil(t, cmd)
	
	// 2. Set window size
	windowMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	model, cmd := dialog.Update(windowMsg)
	dialog = model.(OAuth2DialogCmp)
	
	// 3. Start OAuth2 flow
	cmd = dialog.StartOAuth2(OAuth2ProviderGemini)
	assert.NotNil(t, cmd)
	
	// 4. Simulate success
	successMsg := OAuth2DialogMsg{
		Provider: OAuth2ProviderGemini,
		State:    OAuth2StateSuccess,
		Message:  "Authentication successful!",
		Error:    nil,
	}
	model, cmd = dialog.Update(successMsg)
	dialog = model.(OAuth2DialogCmp)
	
	// 5. Generate view
	view := dialog.View()
	assert.Contains(t, view, "Authentication successful!")
	assert.Contains(t, view, "✓")
	
	// 6. Close dialog
	escapeKey := tea.KeyMsg{Type: tea.KeyEsc}
	model, cmd = dialog.Update(escapeKey)
	assert.NotNil(t, cmd)
	
	msg := cmd()
	assert.Equal(t, "close_oauth2_dialog", msg)
}