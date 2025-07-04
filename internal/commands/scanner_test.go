package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandScanner_parseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantMeta    CommandMetadata
		wantContent string
		wantErr     bool
	}{
		{
			name: "valid frontmatter",
			content: `--- 
name: Test Command
description: A test command
category: testing
hidden: false
aliases:
  - test
  - tc
arguments:
  - name: arg1
    description: First argument
    type: string
    required: true
example: "test arg1"
tags:
  - test
  - example
---
This is the command content.

RUN echo "hello world"
`,
			wantMeta: CommandMetadata{
				Name:        "Test Command",
				Description: "A test command",
				Category:    "testing",
				Hidden:      false,
				Aliases:     []string{"test", "tc"},
				Arguments: []ArgumentDefinition{
					{
						Name:        "arg1",
						Description: "First argument",
						Type:        "string",
						Required:    true,
					},
				},
				Example: "test arg1",
				Tags:    []string{"test", "example"},
			},
			wantContent: `This is the command content.

RUN echo "hello world"`,
			wantErr: false,
		},
		{
			name: "no frontmatter",
			content: `This is just content without frontmatter.

RUN echo "hello world"
`,
			wantMeta:    CommandMetadata{},
			wantContent: `This is just content without frontmatter.

RUN echo "hello world"`,
			wantErr: false,
		},
		{
			name: "empty frontmatter",
			content: `--- 
---
Command content here.
`,
			wantMeta:    CommandMetadata{},
			wantContent: "Command content here.",
			wantErr:     false,
		},
		{
			name: "invalid yaml",
			content: `--- 
name: Test Command
invalid: [unclosed array
---
Content here.
`,
			wantMeta:    CommandMetadata{},
			wantContent: "",
			wantErr:     true,
		},
		{
			name: "partial frontmatter",
			content: `--- 
name: Partial Command
description: Only some fields
---
Content with partial frontmatter.
`,
			wantMeta: CommandMetadata{
				Name:        "Partial Command",
				Description: "Only some fields",
			},
			wantContent: "Content with partial frontmatter.",
			wantErr:     false,
		},
	}

	scanner := NewCommandScanner(DefaultScanOptions())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, content, err := scanner.parseFrontmatter([]byte(tt.content))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMeta, meta)
			assert.Equal(t, tt.wantContent, content)
		})
	}
}

func TestCommandScanner_generateCommandID(t *testing.T) {
	tests := []struct {
		name    string
		relPath string
		want    string
	}{
		{
			name:    "simple file",
			relPath: "test.md",
			want:    "test",
		},
		{
			name:    "nested file",
			relPath: "category/subcategory/command.md",
			want:    "category:subcategory:command",
		},
		{
			name:    "single subdirectory",
			relPath: "git/commit.md",
			want:    "git:commit",
		},
		{
			name:    "no extension",
			relPath: "command",
			want:    "command",
		},
		{
			name:    "different extension",
			relPath: "test.txt",
			want:    "test",
		},
	}

	scanner := NewCommandScanner(DefaultScanOptions())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scanner.generateCommandID(tt.relPath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCommandScanner_ScanDirectory(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "test-commands-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test command files
	testFiles := map[string]string{
		"simple.md": `RUN echo "simple command"`,
		"with-frontmatter.md": `--- 
name: Test Command
description: A test command with frontmatter
category: testing
hidden: false
---
RUN echo "test command"`,
		"hidden.md": `--- 
name: Hidden Command
hidden: true
---
RUN echo "hidden command"`,
		"subdir/nested.md": `--- 
name: Nested Command
description: A command in a subdirectory
---
RUN echo "nested command"`,
		"invalid.txt": `This is not a markdown file`,
		"subdir/deep/deeply-nested.md": `--- 
name: Deeply Nested
---
RUN echo "deeply nested"`,
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tmpDir, filePath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	t.Run("scan all files", func(t *testing.T) {
		scanner := NewCommandScanner(DefaultScanOptions())
		result, err := scanner.ScanDirectory(tmpDir, UserCommand)
		require.NoError(t, err)

		// Should find 5 markdown files (excluding .txt file)
		assert.Len(t, result.Commands, 5)
		assert.Empty(t, result.Errors)

		// Check command IDs
		commandIDs := make([]string, len(result.Commands))
		for i, cmd := range result.Commands {
			commandIDs[i] = cmd.ID
		}

		expectedIDs := []string{"simple", "with-frontmatter", "hidden", "subdir:nested", "subdir:deep:deeply-nested"}
		for _, expectedID := range expectedIDs {
			assert.Contains(t, commandIDs, expectedID)
		}
	})

	t.Run("exclude hidden commands", func(t *testing.T) {
		opts := DefaultScanOptions()
		opts.IncludeHidden = false
		scanner := NewCommandScanner(opts)
		result, err := scanner.ScanDirectory(tmpDir, UserCommand)
		require.NoError(t, err)

		// Should find 4 commands (excluding hidden one)
		assert.Len(t, result.Commands, 4)

		// Verify hidden command is not included
		for _, cmd := range result.Commands {
			assert.NotEqual(t, "hidden", cmd.ID)
		}
	})

	t.Run("limit depth", func(t *testing.T) {
		opts := DefaultScanOptions()
		opts.MaxDepth = 1
		scanner := NewCommandScanner(opts)
		result, err := scanner.ScanDirectory(tmpDir, UserCommand)
		require.NoError(t, err)

		// Should find commands up to depth 1 (excluding deeply-nested)
		commandIDs := make([]string, len(result.Commands))
		for i, cmd := range result.Commands {
			commandIDs[i] = cmd.ID
		}

		assert.Contains(t, commandIDs, "simple")
		assert.Contains(t, commandIDs, "with-frontmatter")
		assert.Contains(t, commandIDs, "hidden")
		assert.Contains(t, commandIDs, "subdir:nested")
		assert.NotContains(t, commandIDs, "subdir:deep:deeply-nested")
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		scanner := NewCommandScanner(DefaultScanOptions())
		result, err := scanner.ScanDirectory("/nonexistent/path", UserCommand)
		require.NoError(t, err)

		assert.Empty(t, result.Commands)
		assert.Empty(t, result.Errors)
	})

	t.Run("handles oversized file", func(t *testing.T) {
		oversizedFile := filepath.Join(tmpDir, "oversized.md")
		err := os.WriteFile(oversizedFile, make([]byte, maxCommandFileSize+1), 0644)
		require.NoError(t, err)

		scanner := NewCommandScanner(DefaultScanOptions())
		result, err := scanner.ScanDirectory(tmpDir, UserCommand)
		require.NoError(t, err)

		// Should have one error
		assert.Len(t, result.Errors, 1)
		assert.Contains(t, result.Errors[0].Error(), "file exceeds size limit")
	})
}

func TestCommandScanner_parseCommandFile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test-parse-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test-command.md")
	content := `--- 
name: Test Command
description: A test command
category: testing
---
RUN echo "test"
READ file.txt`

	err = os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	scanner := NewCommandScanner(DefaultScanOptions())
	command, err := scanner.parseCommandFile(testFile, tmpDir, UserCommand)
	require.NoError(t, err)

	assert.Equal(t, "test-command", command.ID)
	assert.Equal(t, testFile, command.FilePath)
	assert.Equal(t, "test-command.md", command.RelativePath)
	assert.Equal(t, "RUN echo \"test\"\nREAD file.txt", command.Content)
	assert.Equal(t, UserCommand, command.SourceType)
	assert.Equal(t, "Test Command", command.Metadata.Name)
	assert.Equal(t, "A test command", command.Metadata.Description)
	assert.Equal(t, "testing", command.Metadata.Category)
}

func TestScanUserCommands(t *testing.T) {
	// This test would require setting up actual home directory structure
	// For now, we'll just test that it doesn't panic and returns a result
	result, err := ScanUserCommands()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Commands)
	assert.NotNil(t, result.Errors)
	assert.NotNil(t, result.ScannedPaths)
}

func TestScanProjectCommands(t *testing.T) {
	// Create temporary project directory
	tmpDir, err := os.MkdirTemp("", "test-project-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create .opencode/commands directory
	commandsDir := filepath.Join(tmpDir, ".opencode", "commands")
	err = os.MkdirAll(commandsDir, 0755)
	require.NoError(t, err)

	// Create test command file
	testFile := filepath.Join(commandsDir, "project-cmd.md")
	content := `--- 
name: Project Command
description: A project-specific command
---
RUN make test`

	err = os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	result, err := ScanProjectCommands(tmpDir)
	require.NoError(t, err)

	assert.Len(t, result.Commands, 1)
	assert.Equal(t, "project-cmd", result.Commands[0].ID)
	assert.Equal(t, ProjectCommand, result.Commands[0].SourceType)
	assert.Equal(t, "Project Command", result.Commands[0].Metadata.Name)
}

func TestDefaultScanOptions(t *testing.T) {
	opts := DefaultScanOptions()

	assert.True(t, opts.IncludeHidden)
	assert.Equal(t, 0, opts.MaxDepth)
	assert.Equal(t, "*.md", opts.FilePattern)
}

func TestNewCommandScanner(t *testing.T) {
	opts := ScanOptions{
		IncludeHidden: false,
		MaxDepth:      2,
		FilePattern:   "*.txt",
	}

	scanner := NewCommandScanner(opts)
	assert.NotNil(t, scanner)
	assert.Equal(t, opts, scanner.options)
}

func TestScanResult_Merge(t *testing.T) {
	result1 := &ScanResult{
		Commands:     []ParsedCommand{{ID: "cmd1"}},
		Errors:       []error{assert.AnError},
		ScannedPaths: []string{"/path1"},
	}

	result2 := &ScanResult{
		Commands:     []ParsedCommand{{ID: "cmd2"}},
		Errors:       []error{assert.AnError},
		ScannedPaths: []string{"/path2"},
	}

	result1.Merge(result2)

	assert.Len(t, result1.Commands, 2)
	assert.Len(t, result1.Errors, 2)
	assert.Len(t, result1.ScannedPaths, 2)
	assert.Equal(t, "cmd2", result1.Commands[1].ID)
	assert.Equal(t, "/path2", result1.ScannedPaths[1])
}