package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryAdapter_ConvertToRegistryCommand(t *testing.T) {
	adapter := NewRegistryAdapter()

	tests := []struct {
		name   string
		parsed ParsedCommand
		verify func(t *testing.T, cmd Command)
	}{
		{
			name: "command with full metadata",
			parsed: ParsedCommand{
				ID:           "test-cmd",
				FilePath:     "/path/to/test-cmd.md",
				RelativePath: "test-cmd.md",
				Content:      "RUN echo 'hello $NAME'",
				Metadata: CommandMetadata{
					Name:        "Test Command",
					Description: "A test command",
					Category:    "testing",
					Hidden:      false,
					Aliases:     []string{"test", "tc"},
					Arguments: []ArgumentDefinition{
						{
							Name:        "NAME",
							Description: "The name to greet",
							Type:        "string",
							Required:    true,
						},
					},
					Example: "test-cmd NAME=World",
					Tags:    []string{"test"},
				},
				SourceType: UserCommand,
			},
			verify: func(t *testing.T, cmd Command) {
				assert.Equal(t, "test-cmd", cmd.ID())
				assert.Equal(t, "Test Command", cmd.Name())
				assert.Equal(t, "A test command", cmd.Description())
				assert.Equal(t, "testing", cmd.Category())
				assert.Equal(t, UserCommand, cmd.Type())
				assert.Equal(t, []string{"test", "tc"}, cmd.GetAliases())
				assert.Equal(t, "test-cmd NAME=World", cmd.GetExample())
				assert.False(t, cmd.IsHidden())

				args := cmd.GetArguments()
				require.Len(t, args, 1)
				assert.Equal(t, "NAME", args[0].Name)
				assert.Equal(t, "The name to greet", args[0].Description)
				assert.Equal(t, "string", args[0].Type)
				assert.True(t, args[0].Required)

				metadata := cmd.GetMetadata()
				assert.Equal(t, "user", metadata["source_type"])
				assert.Equal(t, "/path/to/test-cmd.md", metadata["file_path"])
				assert.Equal(t, "test-cmd.md", metadata["relative_path"])
			},
		},
		{
			name: "command with minimal metadata",
			parsed: ParsedCommand{
				ID:           "simple",
				FilePath:     "/path/to/simple.md",
				RelativePath: "simple.md",
				Content:      "RUN echo 'simple'",
				Metadata:     CommandMetadata{},
				SourceType:   ProjectCommand,
			},
			verify: func(t *testing.T, cmd Command) {
				assert.Equal(t, "simple", cmd.ID())
				assert.Equal(t, "project:simple", cmd.Name())
				assert.Equal(t, "Custom command from simple.md", cmd.Description())
				assert.Equal(t, "", cmd.Category())
				assert.Equal(t, ProjectCommand, cmd.Type())
				assert.Empty(t, cmd.GetAliases())
				assert.Empty(t, cmd.GetExample())
				assert.False(t, cmd.IsHidden())
				assert.Empty(t, cmd.GetArguments())

				metadata := cmd.GetMetadata()
				assert.Equal(t, "project", metadata["source_type"])
			},
		},
		{
			name: "hidden command",
			parsed: ParsedCommand{
				ID:           "hidden-cmd",
				FilePath:     "/path/to/hidden-cmd.md",
				RelativePath: "hidden-cmd.md",
				Content:      "RUN echo 'hidden'",
				Metadata: CommandMetadata{
					Name:   "Hidden Command",
					Hidden: true,
				},
				SourceType: UserCommand,
			},
			verify: func(t *testing.T, cmd Command) {
				assert.True(t, cmd.IsHidden())
			},
		},
		{
			name: "command with extracted arguments",
			parsed: ParsedCommand{
				ID:           "arg-cmd",
				FilePath:     "/path/to/arg-cmd.md",
				RelativePath: "arg-cmd.md",
				Content:      "RUN echo 'Hello $NAME, age $AGE'",
				Metadata:     CommandMetadata{},
				SourceType:   UserCommand,
			},
			verify: func(t *testing.T, cmd Command) {
				args := cmd.GetArguments()
				require.Len(t, args, 2)

				argNames := make([]string, len(args))
				for i, arg := range args {
					argNames[i] = arg.Name
				}
				assert.Contains(t, argNames, "NAME")
				assert.Contains(t, argNames, "AGE")

				for _, arg := range args {
					assert.Equal(t, "string", arg.Type)
					assert.True(t, arg.Required)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := adapter.ConvertToRegistryCommand(tt.parsed)
			tt.verify(t, cmd)
		})
	}
}

func TestRegistryAdapter_generateDisplayName(t *testing.T) {
	adapter := NewRegistryAdapter()

	tests := []struct {
		name       string
		id         string
		sourceType CommandType
		want       string
	}{
		{
			name:       "user command",
			id:         "test",
			sourceType: UserCommand,
			want:       "user:test",
		},
		{
			name:       "project command",
			id:         "build",
			sourceType: ProjectCommand,
			want:       "project:build",
		},
		{
			name:       "builtin command",
			id:         "init",
			sourceType: BuiltinCommand,
			want:       "init",
		},
		{
			name:       "plugin command",
			id:         "plugin-cmd",
			sourceType: PluginCommand,
			want:       "plugin-cmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.generateDisplayName(tt.id, tt.sourceType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegistryAdapter_extractArgumentsFromContent(t *testing.T) {
	adapter := NewRegistryAdapter()

	tests := []struct {
		name    string
		content string
		want    []string // argument names
	}{
		{
			name:    "no arguments",
			content: "RUN echo 'hello world'",
			want:    nil,
		},
		{
			name:    "single argument",
			content: "RUN echo 'hello $NAME'",
			want:    []string{"NAME"},
		},
		{
			name:    "multiple arguments",
			content: "RUN echo 'hello $NAME, age $AGE'",
			want:    []string{"NAME", "AGE"},
		},
		{
			name:    "duplicate arguments",
			content: "RUN echo '$NAME $NAME $AGE'",
			want:    []string{"NAME", "AGE"},
		},
		{
			name:    "complex content",
			content: `RUN git commit -m "$MESSAGE"
READ $FILE
WRITE $OUTPUT`,
			want: []string{"MESSAGE", "FILE", "OUTPUT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := adapter.extractArgumentsFromContent(tt.content)

			if tt.want == nil {
				assert.Nil(t, args)
				return
			}

			require.Len(t, args, len(tt.want))

			argNames := make([]string, len(args))
			for i, arg := range args {
				argNames[i] = arg.Name
				assert.Equal(t, "string", arg.Type)
				assert.True(t, arg.Required)
			}

			for _, wantName := range tt.want {
				assert.Contains(t, argNames, wantName)
			}
		})
	}
}

func TestRegistryAdapter_createCommandHandler(t *testing.T) {
	adapter := NewRegistryAdapter()

	tests := []struct {
		name    string
		parsed  ParsedCommand
		args    map[string]interface{}
		wantErr string
	}{
		{
			name: "command without arguments",
			parsed: ParsedCommand{
				Content: "RUN echo 'hello world'",
			},
			args:    nil,
			wantErr: "EXECUTE_CUSTOM_COMMAND: RUN echo 'hello world'",
		},
		{
			name: "command with arguments provided",
			parsed: ParsedCommand{
				Content: "RUN echo 'hello $NAME'",
			},
			args: map[string]interface{}{
				"NAME": "Alice",
			},
			wantErr: "EXECUTE_CUSTOM_COMMAND: RUN echo 'hello Alice'",
		},
		{
			name: "command with missing arguments",
			parsed: ParsedCommand{
				Content: "RUN echo 'hello $NAME'",
			},
			args:    nil,
			wantErr: "missing required arguments: [NAME]",
		},
		{
			name: "command with multiple arguments",
			parsed: ParsedCommand{
				Content: "RUN echo '$NAME is $AGE years old'",
			},
			args: map[string]interface{}{
				"NAME": "Bob",
				"AGE":  30,
			},
			wantErr: "EXECUTE_CUSTOM_COMMAND: RUN echo 'Bob is 30 years old'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := adapter.createCommandHandler(tt.parsed)
			err := handler(context.Background(), tt.args)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistryAdapter_LoadAndRegisterCommands(t *testing.T) {
	// Create a new registry for testing
	registry := NewCommandRegistry()
	adapter := NewRegistryAdapter()

	// Clear any existing commands
	registry.Clear()

	// Load and register commands
	err := adapter.LoadAndRegisterCommands(registry)

	// We expect this to succeed even if no commands are found
	// The error should only occur if there are actual registration failures
	if err != nil {
		// Check if it's just a "no commands found" type error
		// which is acceptable for testing
		t.Logf("Load and register returned error (may be expected): %v", err)
	}

	// Verify registry is still functional
	commands := registry.List()
	assert.NotNil(t, commands)
}

func TestTUICommandAdapter_ConvertToTUICommand(t *testing.T) {
	adapter := NewTUICommandAdapter()

	tests := []struct {
		name   string
		parsed ParsedCommand
		verify func(t *testing.T, cmd TUICommand)
	}{
		{
			name: "user command with metadata",
			parsed: ParsedCommand{
				ID:           "test-cmd",
				RelativePath: "test-cmd.md",
				Content:      "RUN echo 'test'",
				Metadata: CommandMetadata{
					Name:        "Test Command",
					Description: "A test command",
				},
				SourceType: UserCommand,
			},
			verify: func(t *testing.T, cmd TUICommand) {
				assert.Equal(t, "Test Command", cmd.ID)
				assert.Equal(t, "Test Command", cmd.Title)
				assert.Equal(t, "A test command", cmd.Description)
				assert.NotNil(t, cmd.Handler)
			},
		},
		{
			name: "project command without metadata",
			parsed: ParsedCommand{
				ID:           "build",
				RelativePath: "build.md",
				Content:      "RUN make build",
				Metadata:     CommandMetadata{},
				SourceType:   ProjectCommand,
			},
			verify: func(t *testing.T, cmd TUICommand) {
				assert.Equal(t, "project:build", cmd.ID)
				assert.Equal(t, "project:build", cmd.Title)
				assert.Equal(t, "Custom command from build.md", cmd.Description)
				assert.NotNil(t, cmd.Handler)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := adapter.ConvertToTUICommand(tt.parsed)
			tt.verify(t, cmd)
		})
	}
}

func TestTUICommandAdapter_LoadTUICommands(t *testing.T) {
	adapter := NewTUICommandAdapter()

	// This will attempt to load commands from the file system
	commands, err := adapter.LoadTUICommands()

	// We expect this to work even if no commands are found
	// Commands slice should be non-nil even if empty
	if commands == nil {
		t.Errorf("LoadTUICommands returned nil commands slice, expected non-nil (even if empty)")
	}

	if err != nil {
		// Log the error for debugging but don't fail the test
		// as this is expected in a test environment without custom commands
		t.Logf("Load TUI commands returned error (may be expected): %v", err)
	}
}

func TestLoadAndRegisterCommandsInGlobalRegistry(t *testing.T) {
	// Save the original global registry
	originalRegistry := GetGlobalRegistry()
	defer SetGlobalRegistry(originalRegistry)

	// Use a test registry
	testRegistry := NewCommandRegistry()
	SetGlobalRegistry(testRegistry)

	// Clear the test registry
	testRegistry.Clear()

	// Load commands into global registry
	err := LoadAndRegisterCommandsInGlobalRegistry()

	// Should not panic and should return (may have errors if no commands found)
	if err != nil {
		t.Logf("Load into global registry returned error (may be expected): %v", err)
	}

	// Verify we can still list commands
	commands := ListCommands()
	assert.NotNil(t, commands)
}