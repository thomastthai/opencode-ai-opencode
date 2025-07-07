package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOptionValue(t *testing.T) {
	tests := []struct {
		name     string
		opt      *Option
		value    string
		expected interface{}
		wantErr  bool
	}{
		// Bool options
		{
			name: "bool with empty value",
			opt: &Option{
				Name: "verbose",
				Type: OptionTypeBool,
			},
			value:    "",
			expected: true,
			wantErr:  false,
		},
		{
			name: "bool with true value",
			opt: &Option{
				Name: "verbose",
				Type: OptionTypeBool,
			},
			value:    "true",
			expected: true,
			wantErr:  false,
		},
		{
			name: "bool with false value",
			opt: &Option{
				Name: "verbose",
				Type: OptionTypeBool,
			},
			value:    "false",
			expected: false,
			wantErr:  false,
		},
		// String options
		{
			name: "string value",
			opt: &Option{
				Name: "name",
				Type: OptionTypeString,
			},
			value:    "test-value",
			expected: "test-value",
			wantErr:  false,
		},
		{
			name: "string with choices - valid",
			opt: &Option{
				Name:    "level",
				Type:    OptionTypeString,
				Choices: []string{"debug", "info", "warn", "error"},
			},
			value:    "info",
			expected: "info",
			wantErr:  false,
		},
		{
			name: "string with choices - invalid",
			opt: &Option{
				Name:    "level",
				Type:    OptionTypeString,
				Choices: []string{"debug", "info", "warn", "error"},
			},
			value:    "invalid",
			expected: nil,
			wantErr:  true,
		},
		// Int options
		{
			name: "int value",
			opt: &Option{
				Name: "count",
				Type: OptionTypeInt,
			},
			value:    "42",
			expected: 42,
			wantErr:  false,
		},
		{
			name: "int with range - valid",
			opt: &Option{
				Name:     "port",
				Type:     OptionTypeInt,
				MinValue: 1,
				MaxValue: 65535,
			},
			value:    "8080",
			expected: 8080,
			wantErr:  false,
		},
		{
			name: "int with range - too low",
			opt: &Option{
				Name:     "port",
				Type:     OptionTypeInt,
				MinValue: 1,
				MaxValue: 65535,
			},
			value:    "0",
			expected: nil,
			wantErr:  true,
		},
		// Float options
		{
			name: "float value",
			opt: &Option{
				Name: "threshold",
				Type: OptionTypeFloat,
			},
			value:    "0.75",
			expected: 0.75,
			wantErr:  false,
		},
		// List options
		{
			name: "list value",
			opt: &Option{
				Name: "tags",
				Type: OptionTypeList,
			},
			value:    "dev,test,prod",
			expected: []string{"dev", "test", "prod"},
			wantErr:  false,
		},
		{
			name: "list with spaces",
			opt: &Option{
				Name: "tags",
				Type: OptionTypeList,
			},
			value:    "dev, test , prod",
			expected: []string{"dev", "test", "prod"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseOptionValue(tt.opt, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParsedOptions(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		po := NewParsedOptions()
		
		// Set and get string
		po.Set("name", &OptionValue{
			Option: &Option{Name: "name", Type: OptionTypeString},
			Value:  "test",
			Source: "flag",
		})
		
		val, exists := po.GetString("name")
		assert.True(t, exists)
		assert.Equal(t, "test", val)
		
		// Set and get int
		po.Set("count", &OptionValue{
			Option: &Option{Name: "count", Type: OptionTypeInt},
			Value:  42,
			Source: "flag",
		})
		
		intVal, exists := po.GetInt("count")
		assert.True(t, exists)
		assert.Equal(t, 42, intVal)
		
		// Get bool (not set, should return false)
		boolVal := po.GetBool("verbose")
		assert.False(t, boolVal)
		
		// Set positional args
		po.SetPositional([]string{"arg1", "arg2"})
		assert.Equal(t, []string{"arg1", "arg2"}, po.GetPositional())
	})

	t.Run("validation", func(t *testing.T) {
		po := NewParsedOptions()
		
		// Required option missing
		options := []*Option{
			{
				Name:     "required",
				Type:     OptionTypeString,
				Required: true,
			},
		}
		
		err := po.Validate(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required option --required is missing")
		
		// Required option present
		po.Set("required", &OptionValue{
			Option: options[0],
			Value:  "value",
			Source: "flag",
		})
		
		err = po.Validate(options)
		assert.NoError(t, err)
	})

	t.Run("conflicts", func(t *testing.T) {
		po := NewParsedOptions()
		
		opt1 := &Option{
			Name:      "force",
			Type:      OptionTypeBool,
			Conflicts: []string{"interactive"},
		}
		
		opt2 := &Option{
			Name: "interactive",
			Type: OptionTypeBool,
		}
		
		po.Set("force", &OptionValue{
			Option: opt1,
			Value:  true,
			Source: "flag",
		})
		
		po.Set("interactive", &OptionValue{
			Option: opt2,
			Value:  true,
			Source: "flag",
		})
		
		options := []*Option{opt1, opt2}
		err := po.Validate(options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "conflicts with")
	})
}

func TestFormatOptionName(t *testing.T) {
	tests := []struct {
		name     string
		opt      *Option
		expected string
	}{
		{
			name: "bool option with short name",
			opt: &Option{
				Name:      "verbose",
				ShortName: "v",
				Type:      OptionTypeBool,
			},
			expected: "-v, --verbose",
		},
		{
			name: "string option without short name",
			opt: &Option{
				Name: "name",
				Type: OptionTypeString,
			},
			expected: "--name=VALUE",
		},
		{
			name: "int option with short name",
			opt: &Option{
				Name:      "count",
				ShortName: "c",
				Type:      OptionTypeInt,
			},
			expected: "-c, --count=N",
		},
		{
			name: "list option",
			opt: &Option{
				Name: "tags",
				Type: OptionTypeList,
			},
			expected: "--tags=LIST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatOptionName(tt.opt)
			assert.Equal(t, tt.expected, result)
		})
	}
}