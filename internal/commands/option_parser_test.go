package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionParser_ParseArgs(t *testing.T) {
	// Define test options
	testOptions := []*Option{
		{
			Name:      "verbose",
			ShortName: "v",
			Type:      OptionTypeBool,
		},
		{
			Name:      "output",
			ShortName: "o",
			Type:      OptionTypeString,
		},
		{
			Name: "format",
			Type: OptionTypeString,
			Choices: []string{"json", "yaml", "xml"},
		},
		{
			Name:     "count",
			ShortName: "c",
			Type:     OptionTypeInt,
			MinValue: 1,
			MaxValue: 100,
		},
		{
			Name: "tags",
			Type: OptionTypeList,
		},
		{
			Name:       "define",
			ShortName:  "D",
			Type:       OptionTypeString,
			Repeatable: true,
		},
		{
			Name:         "mode",
			Type:         OptionTypeString,
			DefaultValue: "normal",
		},
	}

	parser := NewOptionParser(testOptions)

	tests := []struct {
		name           string
		args           []string
		expectedOpts   map[string]interface{}
		expectedPos    []string
		expectedErr    bool
	}{
		{
			name: "no arguments",
			args: []string{},
			expectedOpts: map[string]interface{}{
				"mode": "normal", // default value
			},
			expectedPos: []string{},
			expectedErr: false,
		},
		{
			name: "positional only",
			args: []string{"file1", "file2"},
			expectedOpts: map[string]interface{}{
				"mode": "normal",
			},
			expectedPos: []string{"file1", "file2"},
			expectedErr: false,
		},
		{
			name: "bool flag",
			args: []string{"--verbose", "file1"},
			expectedOpts: map[string]interface{}{
				"verbose": true,
				"mode":    "normal",
			},
			expectedPos: []string{"file1"},
			expectedErr: false,
		},
		{
			name: "short bool flag",
			args: []string{"-v", "file1"},
			expectedOpts: map[string]interface{}{
				"verbose": true,
				"mode":    "normal",
			},
			expectedPos: []string{"file1"},
			expectedErr: false,
		},
		{
			name: "string option with equals",
			args: []string{"--output=result.txt", "file1"},
			expectedOpts: map[string]interface{}{
				"output": "result.txt",
				"mode":   "normal",
			},
			expectedPos: []string{"file1"},
			expectedErr: false,
		},
		{
			name: "string option with space",
			args: []string{"--output", "result.txt", "file1"},
			expectedOpts: map[string]interface{}{
				"output": "result.txt",
				"mode":   "normal",
			},
			expectedPos: []string{"file1"},
			expectedErr: false,
		},
		{
			name: "short string option",
			args: []string{"-o", "result.txt", "file1"},
			expectedOpts: map[string]interface{}{
				"output": "result.txt",
				"mode":   "normal",
			},
			expectedPos: []string{"file1"},
			expectedErr: false,
		},
		{
			name: "short string option concatenated",
			args: []string{"-oresult.txt", "file1"},
			expectedOpts: map[string]interface{}{
				"output": "result.txt",
				"mode":   "normal",
			},
			expectedPos: []string{"file1"},
			expectedErr: false,
		},
		{
			name: "int option",
			args: []string{"--count=5", "file1"},
			expectedOpts: map[string]interface{}{
				"count": 5,
				"mode":  "normal",
			},
			expectedPos: []string{"file1"},
			expectedErr: false,
		},
		{
			name: "list option",
			args: []string{"--tags=dev,test,prod", "file1"},
			expectedOpts: map[string]interface{}{
				"tags": []string{"dev", "test", "prod"},
				"mode": "normal",
			},
			expectedPos: []string{"file1"},
			expectedErr: false,
		},
		{
			name: "multiple options",
			args: []string{"-v", "--output=result.txt", "--format", "json", "file1"},
			expectedOpts: map[string]interface{}{
				"verbose": true,
				"output":  "result.txt",
				"format":  "json",
				"mode":    "normal",
			},
			expectedPos: []string{"file1"},
			expectedErr: false,
		},
		{
			name: "unknown option",
			args: []string{"--unknown", "file1"},
			expectedOpts: map[string]interface{}{},
			expectedPos: []string{},
			expectedErr: true,
		},
		{
			name: "invalid choice",
			args: []string{"--format=invalid", "file1"},
			expectedOpts: map[string]interface{}{},
			expectedPos: []string{},
			expectedErr: true,
		},
		{
			name: "int out of range",
			args: []string{"--count=150", "file1"},
			expectedOpts: map[string]interface{}{},
			expectedPos: []string{},
			expectedErr: true,
		},
		{
			name: "repeatable option",
			args: []string{"-D", "VAR1=value1", "-D", "VAR2=value2", "file1"},
			expectedOpts: map[string]interface{}{
				"define": "VAR1=value1", // First value is stored in Value field
				"mode":   "normal",
			},
			expectedPos: []string{"file1"},
			expectedErr: false,
		},
		{
			name: "mixed short and long options",
			args: []string{"-v", "-o", "output.txt", "--count", "10", "--", "-file-with-dash"},
			expectedOpts: map[string]interface{}{
				"verbose": true,
				"output":  "output.txt",
				"count":   10,
				"mode":    "normal",
			},
			expectedPos: []string{"-file-with-dash"},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseArgs(tt.args)
			
			if tt.expectedErr {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			assert.NotNil(t, result)
			
			// Check options
			for name, expectedValue := range tt.expectedOpts {
				switch v := expectedValue.(type) {
				case string:
					actual, exists := result.GetString(name)
					assert.True(t, exists, "Option %s should exist", name)
					assert.Equal(t, v, actual)
				case int:
					actual, exists := result.GetInt(name)
					assert.True(t, exists, "Option %s should exist", name)
					assert.Equal(t, v, actual)
				case bool:
					actual := result.GetBool(name)
					assert.Equal(t, v, actual)
				case []string:
					actual, exists := result.GetList(name)
					assert.True(t, exists, "Option %s should exist", name)
					assert.Equal(t, v, actual)
				}
			}
			
			// Check positional args
			assert.Equal(t, tt.expectedPos, result.GetPositional())
		})
	}
}

func TestOptionParser_RepeatableOptions(t *testing.T) {
	options := []*Option{
		{
			Name:       "define",
			ShortName:  "D",
			Type:       OptionTypeString,
			Repeatable: true,
		},
	}
	
	parser := NewOptionParser(options)
	
	args := []string{"-D", "VAR1=value1", "-D", "VAR2=value2", "-D", "VAR3=value3"}
	result, err := parser.ParseArgs(args)
	
	assert.NoError(t, err)
	
	// Check that we have the option
	optVal, exists := result.Get("define")
	assert.True(t, exists)
	assert.NotNil(t, optVal)
	
	// Check repeated values
	assert.Equal(t, 3, len(optVal.Repeated))
	assert.Equal(t, "VAR1=value1", optVal.Repeated[0])
	assert.Equal(t, "VAR2=value2", optVal.Repeated[1])
	assert.Equal(t, "VAR3=value3", optVal.Repeated[2])
}

func TestExtractOptionsFromArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantOptions []string
		wantPositional []string
	}{
		{
			name:        "no options",
			args:        []string{"file1", "file2"},
			wantOptions: []string{},
			wantPositional: []string{"file1", "file2"},
		},
		{
			name:        "options before positional",
			args:        []string{"--verbose", "-o", "output.txt", "file1"},
			wantOptions: []string{"--verbose", "-o"},
			wantPositional: []string{"output.txt", "file1"},
		},
		{
			name:        "options with equals",
			args:        []string{"--output=result.txt", "--format=json", "file1"},
			wantOptions: []string{"--output=result.txt", "--format=json"},
			wantPositional: []string{"file1"},
		},
		{
			name:        "mixed options and positional",
			args:        []string{"--verbose", "file1", "file2"},
			wantOptions: []string{"--verbose"},
			wantPositional: []string{"file1", "file2"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options, positional := ExtractOptionsFromArgs(tt.args)
			assert.Equal(t, tt.wantOptions, options)
			assert.Equal(t, tt.wantPositional, positional)
		})
	}
}