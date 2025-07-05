package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldShowInitDialog_ConfigNotLoaded(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	// Ensure config is not loaded
	cfg = nil
	
	_, err := ShouldShowInitDialog()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config not loaded")
}

func TestShouldShowInitDialog_InitFlagExists(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Load config to set up the data directory
	_, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	// Create the init flag file
	flagPath := filepath.Join(Get().Data.Directory, InitFlagFilename)
	err = os.MkdirAll(filepath.Dir(flagPath), 0755)
	require.NoError(t, err)
	
	file, err := os.Create(flagPath)
	require.NoError(t, err)
	file.Close()
	
	shouldShow, err := ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.False(t, shouldShow, "should not show dialog when init flag exists")
}

func TestShouldShowInitDialog_InitFlagDoesNotExist(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Load config to set up the data directory
	_, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	// Ensure the init flag file does NOT exist
	flagPath := filepath.Join(Get().Data.Directory, InitFlagFilename)
	os.Remove(flagPath)
	
	shouldShow, err := ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.True(t, shouldShow, "should show dialog when init flag does not exist")
}

func TestShouldShowInitDialog_DirectoryDoesNotExist(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Load config
	_, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	// Remove the entire data directory to simulate non-existence
	os.RemoveAll(Get().Data.Directory)
	
	shouldShow, err := ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.True(t, shouldShow, "should show dialog when data directory does not exist")
}

func TestShouldShowInitDialog_StatError(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Load config
	_, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	// Create the data directory first
	err = os.MkdirAll(Get().Data.Directory, 0755)
	require.NoError(t, err)
	
	// Create a directory where the init flag file should be, causing a stat error
	flagPath := filepath.Join(Get().Data.Directory, InitFlagFilename)
	err = os.MkdirAll(flagPath, 0755) // Create as directory instead of file
	require.NoError(t, err)
	
	// This should return true (show dialog) but not error in this case
	// The stat succeeds but the path is a directory, not a file
	shouldShow, err := ShouldShowInitDialog()
	assert.NoError(t, err) // os.Stat succeeds on directories
	assert.True(t, shouldShow) // Should show dialog because it's not a proper flag file
}

func TestMarkProjectInitialized_ConfigNotLoaded(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	// Ensure config is not loaded
	cfg = nil
	
	err := MarkProjectInitialized()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config not loaded")
}

func TestMarkProjectInitialized_Success(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Load config
	_, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	// Ensure data directory exists
	err = os.MkdirAll(Get().Data.Directory, 0755)
	require.NoError(t, err)
	
	// Mark project as initialized
	err = MarkProjectInitialized()
	assert.NoError(t, err)
	
	// Verify the flag file was created
	flagPath := filepath.Join(Get().Data.Directory, InitFlagFilename)
	_, err = os.Stat(flagPath)
	assert.NoError(t, err, "init flag file should exist")
	
	// Verify we can read the file
	file, err := os.Open(flagPath)
	assert.NoError(t, err)
	file.Close()
}

func TestMarkProjectInitialized_CreateDirectoryError(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Load config
	_, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	// Create a file where the data directory should be, causing create error
	err = os.WriteFile(Get().Data.Directory, []byte("file content"), 0644)
	require.NoError(t, err)
	
	err = MarkProjectInitialized()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create init flag file")
}

func TestMarkProjectInitialized_Idempotent(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Load config
	_, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	// Ensure data directory exists
	err = os.MkdirAll(Get().Data.Directory, 0755)
	require.NoError(t, err)
	
	// Mark project as initialized multiple times
	err = MarkProjectInitialized()
	assert.NoError(t, err)
	
	err = MarkProjectInitialized()
	assert.NoError(t, err, "should be idempotent")
	
	// Verify the flag file still exists
	flagPath := filepath.Join(Get().Data.Directory, InitFlagFilename)
	_, err = os.Stat(flagPath)
	assert.NoError(t, err, "init flag file should still exist")
}

func TestInitWorkflow_FullCycle(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	
	tmpDir := t.TempDir()
	
	// Load config
	_, err := Load(tmpDir, false)
	require.NoError(t, err)
	
	// Initially should show dialog
	shouldShow, err := ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.True(t, shouldShow, "should initially show dialog")
	
	// Mark as initialized
	err = MarkProjectInitialized()
	assert.NoError(t, err)
	
	// Should no longer show dialog
	shouldShow, err = ShouldShowInitDialog()
	assert.NoError(t, err)
	assert.False(t, shouldShow, "should not show dialog after initialization")
}

func TestInitFilename_Constant(t *testing.T) {
	assert.Equal(t, "init", InitFlagFilename)
}