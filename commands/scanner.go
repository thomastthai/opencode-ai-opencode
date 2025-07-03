// Package commands provides command scanning functionality that discovers
// custom commands from user and project directories, parses their YAML frontmatter,
// and registers them in the command registry system.
package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CommandMetadata represents the YAML frontmatter metadata in a command file
type CommandMetadata struct {
	// Name is the display name of the command
	Name string `yaml:"name,omitempty"`
	
	// Description provides a detailed description of what the command does
	Description string `yaml:"description,omitempty"`
	
	// Category groups commands together
	Category string `yaml:"category,omitempty"`
	
	// Aliases provides alternative names for the command
	Aliases []string `yaml:"aliases,omitempty"`
	
	// Hidden marks the command as hidden from general listings
	Hidden bool `yaml:"hidden,omitempty"`
	
	// Arguments defines the expected arguments for the command
	Arguments []ArgumentDefinition `yaml:"arguments,omitempty"`
	
	// Example provides usage examples
	Example string `yaml:"example,omitempty"`
	
	// Tags for categorization and searching
	Tags []string `yaml:"tags,omitempty"`
	
	// Custom metadata for extensibility
	Custom map[string]interface{} `yaml:",inline"`
}

// ParsedCommand represents a command discovered during scanning
type ParsedCommand struct {
	// ID is the unique identifier derived from the file path
	ID string
	
	// FilePath is the absolute path to the command file
	FilePath string
	
	// RelativePath is the path relative to the commands directory
	RelativePath string
	
	// Content is the command content (without frontmatter)
	Content string
	
	// Metadata contains the parsed YAML frontmatter
	Metadata CommandMetadata
	
	// SourceType indicates the source of the command
	SourceType CommandType
}

// ScanResult contains the results of a directory scan
type ScanResult struct {
	// Commands contains all discovered commands
	Commands []ParsedCommand
	
	// Errors contains any errors encountered during scanning
	Errors []error
	
	// ScannedPaths contains all directories that were scanned
	ScannedPaths []string
}

// ScanOptions configures the scanning behavior
type ScanOptions struct {
	// IncludeHidden controls whether hidden commands are included
	IncludeHidden bool
	
	// MaxDepth limits how deep to scan subdirectories (0 = unlimited)
	MaxDepth int
	
	// FilePattern specifies the file pattern to match (default: "*.md")
	FilePattern string
}

// DefaultScanOptions returns the default scanning options
func DefaultScanOptions() ScanOptions {
	return ScanOptions{
		IncludeHidden: true,
		MaxDepth:      0,
		FilePattern:   "*.md",
	}
}

// CommandScanner provides functionality to scan directories for command files
type CommandScanner struct {
	options ScanOptions
}

// NewCommandScanner creates a new command scanner with the specified options
func NewCommandScanner(options ScanOptions) *CommandScanner {
	return &CommandScanner{
		options: options,
	}
}

// ScanDirectory scans a directory for command files and returns the results
func (cs *CommandScanner) ScanDirectory(dirPath string, sourceType CommandType) (*ScanResult, error) {
	result := &ScanResult{
		Commands:     make([]ParsedCommand, 0),
		Errors:       make([]error, 0),
		ScannedPaths: []string{dirPath},
	}
	
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return result, nil // Return empty result for non-existent directories
	}
	
	// Use filepath.WalkDir for efficient directory traversal
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error accessing %s: %w", path, err))
			return nil // Continue scanning other files
		}
		
		// Skip directories
		if d.IsDir() {
			// Check max depth if specified
			if cs.options.MaxDepth > 0 {
				relPath, relErr := filepath.Rel(dirPath, path)
				if relErr == nil {
					depth := strings.Count(relPath, string(filepath.Separator))
					if depth >= cs.options.MaxDepth {
						return filepath.SkipDir
					}
				}
			}
			return nil
		}
		
		// Check file pattern
		matched, err := filepath.Match(cs.options.FilePattern, d.Name())
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error matching pattern for %s: %w", path, err))
			return nil
		}
		
		if !matched {
			return nil
		}
		
		// Parse the command file
		command, err := cs.parseCommandFile(path, dirPath, sourceType)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error parsing %s: %w", path, err))
			return nil
		}
		
		// Skip hidden commands if not included
		if command.Metadata.Hidden && !cs.options.IncludeHidden {
			return nil
		}
		
		result.Commands = append(result.Commands, *command)
		return nil
	})
	
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("error walking directory %s: %w", dirPath, err))
	}
	
	return result, nil
}

// parseCommandFile parses a single command file and extracts metadata and content
func (cs *CommandScanner) parseCommandFile(filePath, baseDir string, sourceType CommandType) (*ParsedCommand, error) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Get relative path for ID generation
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}
	
	// Parse frontmatter and content
	metadata, bodyContent, err := cs.parseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}
	
	// Generate command ID from file path
	commandID := cs.generateCommandID(relPath)
	
	return &ParsedCommand{
		ID:           commandID,
		FilePath:     filePath,
		RelativePath: relPath,
		Content:      bodyContent,
		Metadata:     metadata,
		SourceType:   sourceType,
	}, nil
}

// parseFrontmatter parses YAML frontmatter from a markdown file
func (cs *CommandScanner) parseFrontmatter(content []byte) (CommandMetadata, string, error) {
	var metadata CommandMetadata
	
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var frontmatterLines []string
	var bodyLines []string
	
	inFrontmatter := false
	frontmatterEnded := false
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Check for frontmatter delimiter
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				frontmatterEnded = true
				continue
			}
		}
		
		if inFrontmatter && !frontmatterEnded {
			frontmatterLines = append(frontmatterLines, line)
		} else if frontmatterEnded || !inFrontmatter {
			bodyLines = append(bodyLines, line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return metadata, "", fmt.Errorf("error scanning file: %w", err)
	}
	
	// Parse YAML frontmatter if present
	if len(frontmatterLines) > 0 {
		yamlContent := strings.Join(frontmatterLines, "\n")
		if err := yaml.Unmarshal([]byte(yamlContent), &metadata); err != nil {
			return metadata, "", fmt.Errorf("error parsing YAML frontmatter: %w", err)
		}
	}
	
	// Join body content
	bodyContent := strings.Join(bodyLines, "\n")
	
	return metadata, bodyContent, nil
}

// generateCommandID generates a command ID from the relative file path
func (cs *CommandScanner) generateCommandID(relPath string) string {
	// Remove file extension
	commandID := strings.TrimSuffix(relPath, filepath.Ext(relPath))
	
	// Replace directory separators with colons
	commandID = strings.ReplaceAll(commandID, string(filepath.Separator), ":")
	
	return commandID
}

// ScanUserCommands scans user command directories and returns discovered commands
func ScanUserCommands(options ...ScanOptions) (*ScanResult, error) {
	opts := DefaultScanOptions()
	if len(options) > 0 {
		opts = options[0]
	}
	
	scanner := NewCommandScanner(opts)
	result := &ScanResult{
		Commands:     make([]ParsedCommand, 0),
		Errors:       make([]error, 0),
		ScannedPaths: make([]string, 0),
	}
	
	// Scan XDG_CONFIG_HOME/opencode/commands
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		// Default to ~/.config if XDG_CONFIG_HOME is not set
		if home, err := os.UserHomeDir(); err == nil {
			xdgConfigHome = filepath.Join(home, ".config")
		}
	}
	
	if xdgConfigHome != "" {
		userCommandsDir := filepath.Join(xdgConfigHome, "opencode", "commands")
		scanResult, err := scanner.ScanDirectory(userCommandsDir, UserCommand)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error scanning XDG config directory: %w", err))
		} else {
			result.Commands = append(result.Commands, scanResult.Commands...)
			result.Errors = append(result.Errors, scanResult.Errors...)
			result.ScannedPaths = append(result.ScannedPaths, scanResult.ScannedPaths...)
		}
	}
	
	// Scan $HOME/.opencode/commands
	if home, err := os.UserHomeDir(); err == nil {
		homeCommandsDir := filepath.Join(home, ".opencode", "commands")
		scanResult, err := scanner.ScanDirectory(homeCommandsDir, UserCommand)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error scanning home directory: %w", err))
		} else {
			result.Commands = append(result.Commands, scanResult.Commands...)
			result.Errors = append(result.Errors, scanResult.Errors...)
			result.ScannedPaths = append(result.ScannedPaths, scanResult.ScannedPaths...)
		}
	}
	
	return result, nil
}

// ScanProjectCommands scans project command directory and returns discovered commands
func ScanProjectCommands(projectDir string, options ...ScanOptions) (*ScanResult, error) {
	opts := DefaultScanOptions()
	if len(options) > 0 {
		opts = options[0]
	}
	
	scanner := NewCommandScanner(opts)
	
	// Scan project/.opencode/commands
	projectCommandsDir := filepath.Join(projectDir, ".opencode", "commands")
	return scanner.ScanDirectory(projectCommandsDir, ProjectCommand)
}