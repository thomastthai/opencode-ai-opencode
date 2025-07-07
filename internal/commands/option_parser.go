package commands

import (
	"fmt"
	"strings"
)

// OptionParser handles parsing of command-line options
type OptionParser struct {
	options map[string]*Option // Map of all available options by name
}

// NewOptionParser creates a new option parser with the given options
func NewOptionParser(options []*Option) *OptionParser {
	optMap := make(map[string]*Option)
	
	// Build map of all option names and aliases
	for _, opt := range options {
		optMap[opt.Name] = opt
		if opt.ShortName != "" {
			optMap[opt.ShortName] = opt
		}
		for _, alias := range opt.Aliases {
			optMap[alias] = opt
		}
	}
	
	return &OptionParser{
		options: optMap,
	}
}

// ParseArgs parses a slice of arguments and extracts options and positional args
func (op *OptionParser) ParseArgs(args []string) (*ParsedOptions, error) {
	parsed := NewParsedOptions()
	positional := []string{}
	
	i := 0
	for i < len(args) {
		arg := args[i]
		
		// Special case: -- marks end of options
		if arg == "--" {
			i++
			// Add remaining as positional
			for ; i < len(args); i++ {
				positional = append(positional, args[i])
			}
			break
		}
		
		// Check if it's an option
		if strings.HasPrefix(arg, "--") {
			// Long option
			name, value, err := op.parseLongOption(arg, args, &i)
			if err != nil {
				return nil, err
			}
			
			opt, exists := op.options[name]
			if !exists {
				return nil, fmt.Errorf("unknown option: --%s", name)
			}
			
			parsedValue, err := op.parseValue(opt, value, parsed)
			if err != nil {
				return nil, fmt.Errorf("error parsing --%s: %w", name, err)
			}
			
			op.addParsedOption(parsed, opt, parsedValue)
			
		} else if strings.HasPrefix(arg, "-") && arg != "-" {
			// Short option(s)
			names, values, err := op.parseShortOptions(arg, args, &i)
			if err != nil {
				return nil, err
			}
			
			for j, name := range names {
				opt, exists := op.options[name]
				if !exists {
					return nil, fmt.Errorf("unknown option: -%s", name)
				}
				
				value := ""
				if j < len(values) {
					value = values[j]
				}
				
				parsedValue, err := op.parseValue(opt, value, parsed)
				if err != nil {
					return nil, fmt.Errorf("error parsing -%s: %w", name, err)
				}
				
				op.addParsedOption(parsed, opt, parsedValue)
			}
			
		} else {
			// Positional argument
			positional = append(positional, arg)
			i++
		}
	}
	
	// Apply defaults for missing options
	for _, opt := range op.options {
		if _, exists := parsed.Get(opt.Name); !exists && opt.DefaultValue != nil {
			parsed.Set(opt.Name, &OptionValue{
				Option: opt,
				Value:  opt.DefaultValue,
				Source: "default",
			})
		}
	}
	
	parsed.SetPositional(positional)
	return parsed, nil
}

// parseLongOption parses a long option like --name=value or --flag
func (op *OptionParser) parseLongOption(arg string, args []string, index *int) (string, string, error) {
	// Remove -- prefix
	opt := strings.TrimPrefix(arg, "--")
	
	// Check for = separator
	if idx := strings.Index(opt, "="); idx != -1 {
		name := opt[:idx]
		value := opt[idx+1:]
		*index++
		return name, value, nil
	}
	
	// No = separator, check if next arg is the value
	name := opt
	*index++
	
	// Look up the option to see if it expects a value
	if option, exists := op.options[name]; exists && option.Type == OptionTypeBool {
		// Boolean flag, no value needed
		return name, "", nil
	}
	
	// Check if there's a next argument that's not an option
	if *index < len(args) && !strings.HasPrefix(args[*index], "-") {
		value := args[*index]
		*index++
		return name, value, nil
	}
	
	// No value provided
	return name, "", nil
}

// parseShortOptions parses short options like -v or -abc or -n5
func (op *OptionParser) parseShortOptions(arg string, args []string, index *int) ([]string, []string, error) {
	// Remove - prefix
	opts := strings.TrimPrefix(arg, "-")
	
	names := []string{}
	values := []string{}
	
	for i := 0; i < len(opts); i++ {
		name := string(opts[i])
		names = append(names, name)
		
		// Check if this option expects a value
		option, exists := op.options[name]
		if !exists {
			continue
		}
		
		if option.Type != OptionTypeBool {
			// This option expects a value
			if i+1 < len(opts) {
				// Value is concatenated (e.g., -n5 or -ofile)
				value := opts[i+1:]
				values = append(values, value)
				*index++
				// We've consumed the rest of the string
				return names[:i+1], []string{value}, nil
			}
			
			// Check next argument
			*index++
			if *index < len(args) && !strings.HasPrefix(args[*index], "-") {
				values = append(values, args[*index])
				*index++
			}
			// Only return the options processed so far
			return names[:i+1], values, nil
		}
	}
	
	*index++
	return names, values, nil
}

// parseValue parses a value according to the option type
func (op *OptionParser) parseValue(opt *Option, value string, parsed *ParsedOptions) (interface{}, error) {
	// For repeatable options, we might need to append
	if opt.Repeatable {
		if _, exists := parsed.Get(opt.Name); exists {
			// This is a repeated option
			return ParseOptionValue(opt, value)
		}
	}
	
	return ParseOptionValue(opt, value)
}

// addParsedOption adds a parsed option to the results
func (op *OptionParser) addParsedOption(parsed *ParsedOptions, opt *Option, value interface{}) {
	if opt.Repeatable {
		// Handle repeatable options
		if existing, exists := parsed.Get(opt.Name); exists {
			existing.Repeated = append(existing.Repeated, value)
			return
		}
		
		// First occurrence
		parsed.Set(opt.Name, &OptionValue{
			Option:   opt,
			Value:    value,
			Source:   "flag",
			Repeated: []interface{}{value},
		})
	} else {
		// Normal option
		parsed.Set(opt.Name, &OptionValue{
			Option: opt,
			Value:  value,
			Source: "flag",
		})
	}
}

// ExtractOptionsFromArgs separates options from positional arguments
// Returns the options part and the remaining positional args
func ExtractOptionsFromArgs(args []string) ([]string, []string) {
	options := []string{}
	positional := []string{}
	
	// Find the first non-option argument
	inPositional := false
	for _, arg := range args {
		if inPositional || (!strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=")) {
			inPositional = true
			positional = append(positional, arg)
		} else {
			options = append(options, arg)
		}
	}
	
	return options, positional
}