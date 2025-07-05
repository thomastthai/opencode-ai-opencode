package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSessionPrefix(t *testing.T) {
	tests := []struct {
		sessionID string
		expected  string
	}{
		{"abcdefghijklmnop", "abcdefgh"},
		{"12345678", "12345678"},
		{"1234567", "1234567"}, // Less than 8 chars
		{"", ""},              // Empty string
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("sessionID_%s", tt.sessionID), func(t *testing.T) {
			result := GetSessionPrefix(tt.sessionID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppendToSessionLogFile_EmptyParams(t *testing.T) {
	// Test with empty MessageDir
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	MessageDir = ""
	result := AppendToSessionLogFile("session123", "test.log", "content")
	assert.Equal(t, "", result)

	// Test with empty sessionID
	MessageDir = "/tmp/test"
	result = AppendToSessionLogFile("", "test.log", "content")
	assert.Equal(t, "", result)
}

func TestAppendToSessionLogFile_Success(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	MessageDir = tmpDir
	sessionID := "testsession12345"
	filename := "test.log"
	content := "test content\n"

	result := AppendToSessionLogFile(sessionID, filename, content)

	// Verify file was created and content written
	expectedPrefix := GetSessionPrefix(sessionID)
	expectedPath := filepath.Join(tmpDir, expectedPrefix, filename)
	assert.Equal(t, expectedPath, result)

	// Verify content
	data, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestAppendToSessionLogFile_MultipleWrites(t *testing.T) {
	tmpDir := t.TempDir()
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	MessageDir = tmpDir
	sessionID := "testsession12345"
	filename := "test.log"

	// Write multiple times
	content1 := "line 1\n"
	content2 := "line 2\n"

	result1 := AppendToSessionLogFile(sessionID, filename, content1)
	result2 := AppendToSessionLogFile(sessionID, filename, content2)

	assert.Equal(t, result1, result2) // Same file path

	// Verify both contents are appended
	data, err := os.ReadFile(result1)
	require.NoError(t, err)
	assert.Equal(t, content1+content2, string(data))
}

func TestWriteRequestMessage_EmptyParams(t *testing.T) {
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	// Test various empty parameter combinations
	MessageDir = ""
	result := WriteRequestMessage("session123", 1, "content")
	assert.Equal(t, "", result)

	MessageDir = "/tmp/test"
	result = WriteRequestMessage("", 1, "content")
	assert.Equal(t, "", result)

	result = WriteRequestMessage("session123", 0, "content")
	assert.Equal(t, "", result)

	result = WriteRequestMessage("session123", -1, "content")
	assert.Equal(t, "", result)
}

func TestWriteRequestMessage_Success(t *testing.T) {
	tmpDir := t.TempDir()
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	MessageDir = tmpDir
	sessionID := "testsession12345"
	requestSeqID := 5
	message := `{"request": "test"}`

	result := WriteRequestMessage(sessionID, requestSeqID, message)

	// Verify file path
	expectedPrefix := GetSessionPrefix(sessionID)
	expectedPath := filepath.Join(tmpDir, expectedPrefix, "5_request.json")
	assert.Equal(t, expectedPath, result)

	// Verify content
	data, err := os.ReadFile(result)
	require.NoError(t, err)
	assert.Equal(t, message, string(data))
}

func TestWriteRequestMessageJson_Success(t *testing.T) {
	tmpDir := t.TempDir()
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	MessageDir = tmpDir
	sessionID := "testsession12345"
	requestSeqID := 3

	testMessage := map[string]interface{}{
		"type":    "request",
		"content": "test message",
		"id":      float64(123),
	}

	result := WriteRequestMessageJson(sessionID, requestSeqID, testMessage)

	// Verify file was created
	assert.NotEmpty(t, result)

	// Verify content is valid JSON
	data, err := os.ReadFile(result)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, testMessage, unmarshaled)
}

func TestAppendToStreamSessionLog_Success(t *testing.T) {
	tmpDir := t.TempDir()
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	MessageDir = tmpDir
	sessionID := "testsession12345"
	requestSeqID := 2
	chunk := "stream chunk data"

	result := AppendToStreamSessionLog(sessionID, requestSeqID, chunk)

	// Verify file path
	expectedPrefix := GetSessionPrefix(sessionID)
	expectedPath := filepath.Join(tmpDir, expectedPrefix, "2_response_stream.log")
	assert.Equal(t, expectedPath, result)

	// Verify content
	data, err := os.ReadFile(result)
	require.NoError(t, err)
	assert.Equal(t, chunk, string(data))
}

func TestAppendToStreamSessionLogJson_Success(t *testing.T) {
	tmpDir := t.TempDir()
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	MessageDir = tmpDir
	sessionID := "testsession12345"
	requestSeqID := 2

	testChunk := map[string]string{
		"type": "stream",
		"data": "chunk data",
	}

	result := AppendToStreamSessionLogJson(sessionID, requestSeqID, testChunk)

	assert.NotEmpty(t, result)

	// Verify content is valid JSON
	data, err := os.ReadFile(result)
	require.NoError(t, err)

	var unmarshaled map[string]string
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, testChunk, unmarshaled)
}

func TestWriteChatResponseJson_Success(t *testing.T) {
	tmpDir := t.TempDir()
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	MessageDir = tmpDir
	sessionID := "testsession12345"
	requestSeqID := 4

	testResponse := map[string]interface{}{
		"response": "chat response",
		"status":   "completed",
		"tokens":   float64(100),
	}

	result := WriteChatResponseJson(sessionID, requestSeqID, testResponse)

	// Verify file path
	expectedPrefix := GetSessionPrefix(sessionID)
	expectedPath := filepath.Join(tmpDir, expectedPrefix, "4_response.json")
	assert.Equal(t, expectedPath, result)

	// Verify content
	data, err := os.ReadFile(result)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, testResponse, unmarshaled)
}

func TestWriteToolResultsJson_Success(t *testing.T) {
	tmpDir := t.TempDir()
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	MessageDir = tmpDir
	sessionID := "testsession12345"
	requestSeqID := 6

	testToolResults := []map[string]interface{}{
		{
			"tool_id": "tool1",
			"result":  "success",
			"output":  "tool output",
		},
		{
			"tool_id": "tool2",
			"result":  "error",
			"error":   "tool failed",
		},
	}

	result := WriteToolResultsJson(sessionID, requestSeqID, testToolResults)

	// Verify file path
	expectedPrefix := GetSessionPrefix(sessionID)
	expectedPath := filepath.Join(tmpDir, expectedPrefix, "6_tool_results.json")
	assert.Equal(t, expectedPath, result)

	// Verify content
	data, err := os.ReadFile(result)
	require.NoError(t, err)

	var unmarshaled []map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, testToolResults, unmarshaled)
}

func TestRecoverPanic_WithPanic(t *testing.T) {
	// Setup temporary directory for panic logs
	originalDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	cleanupCalled := false
	cleanup := func() {
		cleanupCalled = true
	}

	// Function that panics
	testFunc := func() {
		defer RecoverPanic("test-component", cleanup)
		panic("test panic message")
	}

	// This should not panic, but should handle the panic gracefully
	testFunc()

	// Verify cleanup was called
	assert.True(t, cleanupCalled)

	// Verify panic log file was created
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	panicLogFound := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "opencode-panic-test-component-") && strings.HasSuffix(file.Name(), ".log") {
			panicLogFound = true

			// Verify log content
			content, err := os.ReadFile(filepath.Join(tmpDir, file.Name()))
			require.NoError(t, err)

			contentStr := string(content)
			assert.Contains(t, contentStr, "Panic in test-component: test panic message")
			assert.Contains(t, contentStr, "Stack Trace:")
			break
		}
	}
	assert.True(t, panicLogFound, "Panic log file should be created")
}

func TestRecoverPanic_NoPanic(t *testing.T) {
	cleanupCalled := false
	cleanup := func() {
		cleanupCalled = true
	}

	// Function that doesn't panic
	testFunc := func() {
		defer RecoverPanic("test-component", cleanup)
		// Normal execution, no panic
	}

	testFunc()

	// Cleanup should not be called when there's no panic
	assert.False(t, cleanupCalled)
}

func TestRecoverPanic_NilCleanup(t *testing.T) {
	// Setup temporary directory for panic logs
	originalDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Function that panics with nil cleanup
	testFunc := func() {
		defer RecoverPanic("test-component", nil)
		panic("test panic with nil cleanup")
	}

	// Should not panic even with nil cleanup
	testFunc()

	// Verify panic log file was still created
	files, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	panicLogFound := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "opencode-panic-test-component-") {
			panicLogFound = true
			break
		}
	}
	assert.True(t, panicLogFound)
}

func TestConcurrentSessionLogging(t *testing.T) {
	tmpDir := t.TempDir()
	originalMessageDir := MessageDir
	defer func() { MessageDir = originalMessageDir }()

	MessageDir = tmpDir
	sessionID := "concurrent12345"

	// Write to the same session from multiple goroutines
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			filename := fmt.Sprintf("test_%d.log", id)
			content := fmt.Sprintf("content from goroutine %d\n", id)
			result := AppendToSessionLogFile(sessionID, filename, content)
			assert.NotEmpty(t, result)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all files were created
	sessionPrefix := GetSessionPrefix(sessionID)
	sessionDir := filepath.Join(tmpDir, sessionPrefix)

	files, err := os.ReadDir(sessionDir)
	require.NoError(t, err)
	assert.Len(t, files, numGoroutines)
}