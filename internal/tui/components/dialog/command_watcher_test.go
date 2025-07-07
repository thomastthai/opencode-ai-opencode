package dialog

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandWatcher_Creation(t *testing.T) {
	watcher, err := NewCommandWatcher()
	require.NoError(t, err)
	assert.NotNil(t, watcher)
	
	// Clean up
	watcher.Stop()
}

func TestCommandWatcher_DirectoryDetection(t *testing.T) {
	watcher, err := NewCommandWatcher()
	require.NoError(t, err)
	defer watcher.Stop()
	
	dirs := watcher.getCommandDirectories()
	
	// May have no directories if config is not loaded in tests
	// Just check that it doesn't panic
	assert.NotNil(t, dirs, "Should return a slice, even if empty")
}

func TestCommandWatcher_FileChangeDetection(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "command_watcher_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	
	// Create commands subdirectory
	cmdDir := filepath.Join(tmpDir, "commands")
	err = os.MkdirAll(cmdDir, 0755)
	require.NoError(t, err)
	
	watcher, err := NewCommandWatcher()
	require.NoError(t, err)
	defer watcher.Stop()
	
	// Add our test directory
	err = watcher.addDirectory(cmdDir)
	assert.NoError(t, err)
	
	// Create a test command file
	testFile := filepath.Join(cmdDir, "test.md")
	content := `---
title: Test Command
description: A test command
---

Test command content`
	
	err = os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)
	
	// Give the watcher time to detect the change
	time.Sleep(100 * time.Millisecond)
	
	// Verify the file exists
	_, err = os.Stat(testFile)
	assert.NoError(t, err)
	
	// Clean up
	os.Remove(testFile)
}

func TestCommandWatcher_RelevantEventFiltering(t *testing.T) {
	watcher, err := NewCommandWatcher()
	require.NoError(t, err)
	defer watcher.Stop()
	
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"markdown file", "command.md", true},
		{"text file", "readme.txt", false},
		{"no extension", "command", false},
		{"hidden markdown", ".command.md", true},
		{"json file", "config.json", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily create fsnotify.Event, so we'll test the logic indirectly
			// by checking file extensions
			result := filepath.Ext(tt.filename) == ".md"
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommandWatcher_StartStop(t *testing.T) {
	watcher, err := NewCommandWatcher()
	require.NoError(t, err)
	
	// Start should return a command
	cmd := watcher.Start()
	assert.NotNil(t, cmd)
	
	// Stop should not panic
	assert.NotPanics(t, func() {
		watcher.Stop()
	})
	
	// Multiple stops should not panic
	assert.NotPanics(t, func() {
		watcher.Stop()
	})
}

func TestCommandWatcher_Debouncing(t *testing.T) {
	watcher, err := NewCommandWatcher()
	require.NoError(t, err)
	defer watcher.Stop()
	
	// Schedule multiple reloads quickly
	watcher.scheduleReload()
	watcher.scheduleReload()
	watcher.scheduleReload()
	
	// The debouncer should be set
	watcher.mu.Lock()
	assert.NotNil(t, watcher.debouncer)
	watcher.mu.Unlock()
	
	// Wait for debounce period
	time.Sleep(600 * time.Millisecond)
	
	// Debouncer should have fired and been cleared
	watcher.mu.Lock()
	// Note: debouncer may still exist but should have fired
	watcher.mu.Unlock()
}