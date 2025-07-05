package dialog

import (
	"os"
	"path/filepath"
	"testing"
	"regexp"
	"strings"
)

func TestNamedArgPattern(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "This is a test with $ARGUMENTS placeholder",
			expected: []string{"ARGUMENTS"},
		},
		{
			input:    "This is a test with $FOO and $BAR placeholders",
			expected: []string{"FOO", "BAR"},
		},
		{
			input:    "This is a test with $FOO_BAR and $BAZ123 placeholders",
			expected: []string{"FOO_BAR", "BAZ123"},
		},
		{
			input:    "This is a test with no placeholders",
			expected: []string{},
		},
		{
			input:    "This is a test with $FOO appearing twice: $FOO",
			expected: []string{"FOO"},
		},
		{
			input:    "This is a test with $1INVALID placeholder",
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		matches := namedArgPattern.FindAllStringSubmatch(tc.input, -1)
		
		// Extract unique argument names
		argNames := make([]string, 0)
		argMap := make(map[string]bool)
		
		for _, match := range matches {
			argName := match[1] // Group 1 is the name without $
			if !argMap[argName] {
				argMap[argName] = true
				argNames = append(argNames, argName)
			}
		}
		
		// Check if we got the expected number of arguments
		if len(argNames) != len(tc.expected) {
			t.Errorf("Expected %d arguments, got %d for input: %s", len(tc.expected), len(argNames), tc.input)
			continue
		}
		
		// Check if we got the expected argument names
		for _, expectedArg := range tc.expected {
			found := false
			for _, actualArg := range argNames {
				if actualArg == expectedArg {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected argument %s not found in %v for input: %s", expectedArg, argNames, tc.input)
			}
		}
	}
}

func TestRegexPattern(t *testing.T) {
	pattern := regexp.MustCompile(`\$([A-Z][A-Z0-9_]*)`)
	
	validMatches := []string{
		"$FOO",
		"$BAR",
		"$FOO_BAR",
		"$BAZ123",
		"$ARGUMENTS",
	}
	
	invalidMatches := []string{
		"$foo",
		"$1BAR",
		"$_FOO",
		"FOO",
		"$",
	}
	
	for _, valid := range validMatches {
		if !pattern.MatchString(valid) {
			t.Errorf("Expected %s to match, but it didn't", valid)
		}
	}
	
	for _, invalid := range invalidMatches {
		if pattern.MatchString(invalid) {
			t.Errorf("Expected %s not to match, but it did", invalid)
		}
	}
}

func TestLoadCustomCommands(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up XDG_CONFIG_HOME
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	
	// Create test commands directory
	commandsDir := filepath.Join(tempDir, "opencode", "commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test commands directory: %v", err)
	}
	
	// Create test command files
	testCommands := map[string]string{
		"test-simple.md": "echo 'Hello World'",
		"test-with-args.md": "echo 'Hello $NAME, welcome to $PROJECT'",
		"test-complex.md": `---
name: "Complex Test"
description: "A complex test command"
aliases: ["complex", "comp"]
---
echo 'Running complex command with $ARG1 and $ARG2'`,
		"subdir/nested.md": "echo 'This is a nested command'",
	}
	
	for filename, content := range testCommands {
		fullPath := filepath.Join(commandsDir, filename)
		dir := filepath.Dir(fullPath)
		
		// Create subdirectory if needed
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create subdirectory %s: %v", dir, err)
		}
		
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test command file %s: %v", fullPath, err)
		}
	}
	
	// Test loading commands from directory
	t.Run("load commands from directory", func(t *testing.T) {
		commands, err := loadCommandsFromDir(commandsDir, UserCommandPrefix)
		if err != nil {
			t.Fatalf("Failed to load commands: %v", err)
		}
		
		if len(commands) != 4 {
			t.Errorf("Expected 4 commands, got %d", len(commands))
		}
		
		// Check that command IDs are prefixed correctly
		for _, cmd := range commands {
			if !strings.HasPrefix(cmd.ID, UserCommandPrefix) {
				t.Errorf("Command ID %s should have prefix %s", cmd.ID, UserCommandPrefix)
			}
		}
	})
	
	// Test nested command ID generation
	t.Run("nested command ID", func(t *testing.T) {
		commands, err := loadCommandsFromDir(commandsDir, UserCommandPrefix)
		if err != nil {
			t.Fatalf("Failed to load commands: %v", err)
		}
		
		// Find the nested command
		var nestedCmd *Command
		for _, cmd := range commands {
			if strings.Contains(cmd.ID, "subdir:nested") {
				nestedCmd = &cmd
				break
			}
		}
		
		if nestedCmd == nil {
			t.Error("Expected to find nested command with subdir:nested in ID")
		}
	})
}

func TestLoadCommandsFromDirEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	commandsDir := filepath.Join(tempDir, "empty-commands")
	
	// Don't create the directory - it should be created automatically
	commands, err := loadCommandsFromDir(commandsDir, UserCommandPrefix)
	if err != nil {
		t.Fatalf("Failed to load from non-existent directory: %v", err)
	}
	
	if len(commands) != 0 {
		t.Errorf("Expected 0 commands from empty directory, got %d", len(commands))
	}
	
	// Check that directory was created
	if _, err := os.Stat(commandsDir); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

func TestLoadCommandsFromDirInvalidFiles(t *testing.T) {
	tempDir := t.TempDir()
	commandsDir := filepath.Join(tempDir, "invalid-commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create non-markdown files that should be ignored
	testFiles := map[string]string{
		"test.txt":  "This is not a markdown file",
		"README":    "This has no extension",
		"script.sh": "#!/bin/bash\necho 'shell script'",
		"valid.md":  "echo 'This should be loaded'",
	}
	
	for filename, content := range testFiles {
		fullPath := filepath.Join(commandsDir, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", fullPath, err)
		}
	}
	
	commands, err := loadCommandsFromDir(commandsDir, ProjectCommandPrefix)
	if err != nil {
		t.Fatalf("Failed to load commands: %v", err)
	}
	
	// Should only load the .md file
	if len(commands) != 1 {
		t.Errorf("Expected 1 command (only .md file), got %d", len(commands))
	}
	
	if commands[0].ID != "project:valid" {
		t.Errorf("Expected command ID 'project:valid', got '%s'", commands[0].ID)
	}
}

func TestLoadCommandsFromDirFileReadError(t *testing.T) {
	tempDir := t.TempDir()
	commandsDir := filepath.Join(tempDir, "error-commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create a command file
	validPath := filepath.Join(commandsDir, "valid.md")
	if err := os.WriteFile(validPath, []byte("echo 'valid'"), 0644); err != nil {
		t.Fatalf("Failed to write valid command: %v", err)
	}
	
	// Create a file with no read permissions (on Unix systems)
	if os.Getenv("CI") == "" { // Skip on CI where permissions might not work as expected
		unreadablePath := filepath.Join(commandsDir, "unreadable.md")
		if err := os.WriteFile(unreadablePath, []byte("echo 'unreadable'"), 0000); err != nil {
			t.Fatalf("Failed to write unreadable command: %v", err)
		}
		defer os.Chmod(unreadablePath, 0644) // Clean up for deletion
		
		_, err = loadCommandsFromDir(commandsDir, UserCommandPrefix)
		if err == nil {
			t.Error("Expected error when reading unreadable file")
		}
	}
}

func TestCommandRunCustomMsg(t *testing.T) {
	msg := CommandRunCustomMsg{
		Content: "echo 'test'",
		Args:    map[string]string{"NAME": "World"},
	}
	
	if msg.Content != "echo 'test'" {
		t.Errorf("Expected content 'echo 'test'', got '%s'", msg.Content)
	}
	
	if msg.Args["NAME"] != "World" {
		t.Errorf("Expected arg NAME to be 'World', got '%s'", msg.Args["NAME"])
	}
}

func TestCommandArgumentExtraction(t *testing.T) {
	testCases := []struct {
		name        string
		content     string
		expectedArgs []string
	}{
		{
			name:        "no arguments",
			content:     "echo 'Hello World'",
			expectedArgs: []string{},
		},
		{
			name:        "single argument",
			content:     "echo 'Hello $NAME'",
			expectedArgs: []string{"NAME"},
		},
		{
			name:        "multiple arguments",
			content:     "deploy $APP to $ENV with $VERSION",
			expectedArgs: []string{"APP", "ENV", "VERSION"},
		},
		{
			name:        "duplicate arguments",
			content:     "echo $NAME $NAME $NAME",
			expectedArgs: []string{"NAME"},
		},
		{
			name:        "mixed valid and invalid",
			content:     "echo $VALID $invalid $123INVALID $ANOTHER_VALID",
			expectedArgs: []string{"VALID", "ANOTHER_VALID"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matches := namedArgPattern.FindAllStringSubmatch(tc.content, -1)
			
			// Extract unique argument names (same logic as in loadCommandsFromDir)
			argNames := make([]string, 0)
			argMap := make(map[string]bool)
			
			for _, match := range matches {
				argName := match[1] // Group 1 is the name without $
				if !argMap[argName] {
					argMap[argName] = true
					argNames = append(argNames, argName)
				}
			}
			
			if len(argNames) != len(tc.expectedArgs) {
				t.Errorf("Expected %d arguments, got %d", len(tc.expectedArgs), len(argNames))
				return
			}
			
			// Check that all expected arguments are found
			for _, expectedArg := range tc.expectedArgs {
				found := false
				for _, actualArg := range argNames {
					if actualArg == expectedArg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected argument %s not found in %v", expectedArg, argNames)
				}
			}
		})
	}
}

func TestCommandPrefixConstants(t *testing.T) {
	if UserCommandPrefix != "user:" {
		t.Errorf("Expected UserCommandPrefix to be 'user:', got '%s'", UserCommandPrefix)
	}
	
	if ProjectCommandPrefix != "project:" {
		t.Errorf("Expected ProjectCommandPrefix to be 'project:', got '%s'", ProjectCommandPrefix)
	}
}