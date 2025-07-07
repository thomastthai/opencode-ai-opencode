package commands

import (
	"fmt"
	"strconv"
	"strings"
)

// OptionType represents the type of an option value
type OptionType int

const (
	OptionTypeBool   OptionType = iota // Flag with no value (--verbose)
	OptionTypeString                    // String value (--name=value)
	OptionTypeInt                       // Integer value (--count=5)
	OptionTypeFloat                     // Float value (--threshold=0.5)
	OptionTypeList                      // List of values (--tags=a,b,c)
)

// Option represents a command-line option/switch
type Option struct {
	// Identification
	Name      string   // Long name (e.g., "verbose")
	ShortName string   // Short name (e.g., "v")
	Aliases   []string // Additional aliases

	// Type and validation
	Type         OptionType
	Required     bool
	DefaultValue interface{}
	Choices      []string // Valid choices for string options
	MinValue     float64  // For numeric options
	MaxValue     float64  // For numeric options

	// Documentation
	Description string
	Example     string

	// Behavior
	Repeatable bool   // Can be specified multiple times
	Hidden     bool   // Hide from help
	Global     bool   // Available for all commands in topic
	Conflicts  []string // Names of conflicting options
}

// OptionValue represents a parsed option value
type OptionValue struct {
	Option   *Option
	Value    interface{}
	Source   string // Where the value came from (flag, env, default)
	Repeated []interface{} // For repeatable options
}

// ParsedOptions represents all parsed options for a command
type ParsedOptions struct {
	values   map[string]*OptionValue
	positional []string // Remaining positional arguments
}

// NewParsedOptions creates a new ParsedOptions instance
func NewParsedOptions() *ParsedOptions {
	return &ParsedOptions{
		values:     make(map[string]*OptionValue),
		positional: []string{},
	}
}

// Set adds or updates an option value
func (po *ParsedOptions) Set(name string, value *OptionValue) {
	po.values[name] = value
}

// Get retrieves an option value by name
func (po *ParsedOptions) Get(name string) (*OptionValue, bool) {
	val, exists := po.values[name]
	return val, exists
}

// GetString retrieves a string option value
func (po *ParsedOptions) GetString(name string) (string, bool) {
	if po == nil || po.values == nil {
		return "", false
	}
	val, exists := po.values[name]
	if !exists {
		return "", false
	}
	str, ok := val.Value.(string)
	return str, ok
}

// GetInt retrieves an integer option value
func (po *ParsedOptions) GetInt(name string) (int, bool) {
	if po == nil || po.values == nil {
		return 0, false
	}
	val, exists := po.values[name]
	if !exists {
		return 0, false
	}
	num, ok := val.Value.(int)
	return num, ok
}

// GetFloat retrieves a float option value
func (po *ParsedOptions) GetFloat(name string) (float64, bool) {
	val, exists := po.values[name]
	if !exists {
		return 0, false
	}
	num, ok := val.Value.(float64)
	return num, ok
}

// GetBool retrieves a boolean option value
func (po *ParsedOptions) GetBool(name string) bool {
	if po == nil || po.values == nil {
		return false
	}
	val, exists := po.values[name]
	if !exists {
		return false
	}
	b, _ := val.Value.(bool)
	return b
}

// GetList retrieves a list option value
func (po *ParsedOptions) GetList(name string) ([]string, bool) {
	val, exists := po.values[name]
	if !exists {
		return nil, false
	}
	list, ok := val.Value.([]string)
	return list, ok
}

// SetPositional sets the positional arguments
func (po *ParsedOptions) SetPositional(args []string) {
	po.positional = args
}

// GetPositional returns the positional arguments
func (po *ParsedOptions) GetPositional() []string {
	return po.positional
}

// Validate checks if all required options are present and valid
func (po *ParsedOptions) Validate(options []*Option) error {
	// Check required options
	for _, opt := range options {
		if opt.Required {
			if _, exists := po.values[opt.Name]; !exists {
				return fmt.Errorf("required option --%s is missing", opt.Name)
			}
		}
	}

	// Check for conflicts
	for name, val := range po.values {
		if val.Option != nil && len(val.Option.Conflicts) > 0 {
			for _, conflict := range val.Option.Conflicts {
				if _, exists := po.values[conflict]; exists {
					return fmt.Errorf("option --%s conflicts with --%s", name, conflict)
				}
			}
		}
	}

	return nil
}

// ParseOptionValue parses a string value according to the option type
func ParseOptionValue(opt *Option, value string) (interface{}, error) {
	switch opt.Type {
	case OptionTypeBool:
		// Bool options don't need a value, presence means true
		if value == "" {
			return true, nil
		}
		return strconv.ParseBool(value)

	case OptionTypeString:
		// Validate against choices if specified
		if len(opt.Choices) > 0 {
			valid := false
			for _, choice := range opt.Choices {
				if value == choice {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("invalid choice '%s', must be one of: %s", 
					value, strings.Join(opt.Choices, ", "))
			}
		}
		return value, nil

	case OptionTypeInt:
		num, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid integer value: %s", value)
		}
		// Validate range if specified
		if opt.MinValue != 0 || opt.MaxValue != 0 {
			if float64(num) < opt.MinValue || float64(num) > opt.MaxValue {
				return nil, fmt.Errorf("value %d is out of range [%.0f, %.0f]", 
					num, opt.MinValue, opt.MaxValue)
			}
		}
		return num, nil

	case OptionTypeFloat:
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float value: %s", value)
		}
		// Validate range if specified
		if opt.MinValue != 0 || opt.MaxValue != 0 {
			if num < opt.MinValue || num > opt.MaxValue {
				return nil, fmt.Errorf("value %f is out of range [%f, %f]", 
					num, opt.MinValue, opt.MaxValue)
			}
		}
		return num, nil

	case OptionTypeList:
		// Split by comma
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unknown option type: %v", opt.Type)
	}
}

// FormatOptionName formats an option for display (e.g., "-v, --verbose")
func FormatOptionName(opt *Option) string {
	parts := []string{}
	
	if opt.ShortName != "" {
		parts = append(parts, "-"+opt.ShortName)
	}
	
	parts = append(parts, "--"+opt.Name)
	
	// Add value placeholder for non-bool options
	if opt.Type != OptionTypeBool {
		valuePlaceholder := "VALUE"
		switch opt.Type {
		case OptionTypeInt:
			valuePlaceholder = "N"
		case OptionTypeFloat:
			valuePlaceholder = "NUM"
		case OptionTypeList:
			valuePlaceholder = "LIST"
		}
		parts[len(parts)-1] += "=" + valuePlaceholder
	}
	
	return strings.Join(parts, ", ")
}