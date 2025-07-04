package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryAdapter_ConvertToRegistryCommand(t *testing.T) {
	adapter := NewRegistryAdapter()

	parsed := ParsedCommand{
		ID:           "test-cmd",
		FilePath:     "/path/to/test-cmd.md",
		RelativePath: "test-cmd.md",
		Content:      "RUN echo 'hello $NAME'",
		Metadata: CommandMetadata{
			Name:        "Test Command",
			Description: "A test command",
			Category:    "testing",
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
		},
		SourceType: UserCommand,
	}

	cmd := adapter.ConvertToRegistryCommand(parsed)

	assert.Equal(t, "test-cmd", cmd.ID())
	assert.Equal(t, "Test Command", cmd.Name())
	assert.Equal(t, "A test command", cmd.Description())
	assert.Equal(t, "testing", cmd.Category())
	assert.Equal(t, UserCommand, cmd.Type())
	assert.Equal(t, []string{"test", "tc"}, cmd.GetAliases())
	assert.Equal(t, "test-cmd NAME=World", cmd.Example())

	args := cmd.GetArguments()
	require.Len(t, args, 1)
	assert.Equal(t, "NAME", args[0].Name)
}

func TestCreateCommandHandler(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		args          map[string]interface{}
		expectedError string
	}{
		{
			name:          "simple substitution",
			content:       "echo $MSG",
			args:          map[string]interface{}{"MSG": "hello"},
			expectedError: "custom command execution: echo hello",
		},
		{
			name:          "no-op substitution",
			content:       "echo this is not an $ARG",
			args:          map[string]interface{}{},
			expectedError: "custom command execution: echo this is not an $ARG",
		},
		{
			name:          "adjacent arguments",
			content:       "echo $ARG1$ARG2",
			args:          map[string]interface{}{"ARG1": "hello", "ARG2": "world"},
			expectedError: "custom command execution: echo helloworld",
		},
		{
			name:          "arguments with special characters",
			content:       "echo ($ARG1)",
			args:          map[string]interface{}{"ARG1": "hello"},
			expectedError: "custom command execution: echo (hello)",
		},
		{
			name:          "incomplete substitution",
			content:       "echo $ARG1 $ARG2",
			args:          map[string]interface{}{"ARG1": "hello"},
			expectedError: "custom command execution: echo hello $ARG2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := ParsedCommand{Content: tt.content}
			handler := createCommandHandler(parsed)
			err := handler(context.Background(), tt.args)
			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err.Error())
		})
	}
}
