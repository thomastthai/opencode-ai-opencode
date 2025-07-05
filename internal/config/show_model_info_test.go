package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowModelInfo_DefaultValue(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Create a minimal config to avoid interference from existing configs
	minimalConfig := `{}`
	configPath := filepath.Join(tmpDir, ".opencode.json")
	err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
	require.NoError(t, err)
	
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	// Default should be true
	assert.True(t, config.TUI.ShowModelInfo, "ShowModelInfo should default to true")
}

func TestShowModelInfo_ExplicitlySetToFalse(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Create config with showModelInfo set to false
	configContent := `{
		"tui": {
			"showModelInfo": false
		}
	}`
	
	configPath := filepath.Join(tmpDir, ".opencode.json")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	assert.False(t, config.TUI.ShowModelInfo, "ShowModelInfo should be false when explicitly set")
}

func TestShowModelInfo_ExplicitlySetToTrue(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Create config with showModelInfo set to true
	configContent := `{
		"tui": {
			"showModelInfo": true
		}
	}`
	
	configPath := filepath.Join(tmpDir, ".opencode.json")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	config, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	assert.True(t, config.TUI.ShowModelInfo, "ShowModelInfo should be true when explicitly set")
}