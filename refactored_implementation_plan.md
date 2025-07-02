# Refactored Implementation Plan: Extensible Slash Commands with Autocomplete, Popups, and Sub-Commands

## Overview

This document outlines the detailed action plan for refactoring the OpenCode command system to support extensible slash commands with autocomplete, popups, and sub-commands. The goal is to create a unified, flexible command framework that enhances user experience while maintaining backward compatibility.

## Current State Analysis

### Existing Command System
- Commands are defined in `internal/tui/components/dialog/commands.go`
- Custom commands loaded from markdown files in user/project directories
- Simple command structure: ID, Title, Description, Handler
- Basic command dialog with list selection
- Named argument support with placeholders (`$VARIABLE_NAME`)

### Existing Registry Pattern
- Provider registry in `internal/llm/provider/registry.go` shows good patterns for:
  - Thread-safe registration
  - Factory patterns
  - Metadata storage
  - Error handling

## Target Architecture

### 1. Command Registry System (`commands/registry.go`)
- **Command Interface**: Base interface all commands must implement
- **Command Types**: Support for different command categories
  - Built-in commands (Initialize Project, Compact Session)
  - User commands (from `~/.config/opencode/commands/`)
  - Project commands (from `.opencode/commands/`)
  - Plugin commands (extensible for future MCP integration)
- **Sub-command Support**: Hierarchical command structure
- **Metadata Storage**: Rich command metadata for autocomplete

### 2. Autocomplete System
- **Fuzzy Search**: Integrate with existing `lithammer/fuzzysearch` dependency
- **Context-aware Suggestions**: Based on current state/mode
- **Real-time Filtering**: As user types slash commands
- **Keyboard Navigation**: Arrow keys, tab completion

### 3. Popup Command Interface
- **Command Palette**: Triggered by `/` or `Ctrl+K`
- **Preview Pane**: Show command descriptions and examples
- **Argument Input**: Interactive argument collection
- **Visual Hierarchy**: Clear distinction between categories and sub-commands

### 4. Sub-command Architecture
- **Nested Structure**: Support for `command:subcommand:action` syntax
- **Inherited Context**: Sub-commands inherit parent command context
- **Dynamic Loading**: Sub-commands can be dynamically registered
- **Help System**: Automatic help generation for command trees

## Implementation Phases

### Phase 1: Foundation (This PR)
- [x] Create feature branch
- [ ] Document implementation plan
- [ ] Scaffold `commands/registry.go` with core interfaces
- [ ] Define base command types and registry structure
- [ ] Ensure no breaking changes to existing system

### Phase 2: Registry Implementation
- [ ] Implement thread-safe command registry
- [ ] Create command factory patterns
- [ ] Add command metadata support
- [ ] Implement command validation
- [ ] Add comprehensive error handling

### Phase 3: Autocomplete Engine
- [ ] Integrate fuzzy search for command matching
- [ ] Implement real-time command filtering
- [ ] Add context-aware suggestions
- [ ] Create keyboard navigation system
- [ ] Add command ranking/scoring

### Phase 4: Enhanced UI Components
- [ ] Refactor command dialog for autocomplete
- [ ] Add command preview pane
- [ ] Implement argument collection UI
- [ ] Create visual command categories
- [ ] Add help tooltips and documentation

### Phase 5: Sub-command System
- [ ] Implement hierarchical command structure
- [ ] Add nested command registration
- [ ] Create sub-command discovery
- [ ] Implement context inheritance
- [ ] Add dynamic command loading

### Phase 6: Integration & Migration
- [ ] Migrate existing built-in commands
- [ ] Update custom command loading
- [ ] Integrate with existing TUI system
- [ ] Add backward compatibility layer
- [ ] Update documentation

### Phase 7: Advanced Features
- [ ] Add command aliases
- [ ] Implement command history
- [ ] Add favorite commands
- [ ] Create command analytics
- [ ] Add plugin command support

## Technical Specifications

### Command Interface Design
```go
type Command interface {
    ID() string
    Name() string
    Description() string
    Category() string
    Execute(ctx context.Context, args map[string]interface{}) error
    ValidateArgs(args map[string]interface{}) error
    GetArguments() []ArgumentDefinition
    GetSubCommands() []Command
}

type CommandType int
const (
    BuiltinCommand CommandType = iota
    UserCommand
    ProjectCommand
    PluginCommand
)
```

### Registry Interface Design
```go
type CommandRegistry interface {
    Register(cmd Command) error
    Unregister(id string) error
    Get(id string) (Command, bool)
    List() []Command
    Search(query string) []Command
    GetByCategory(category string) []Command
}
```

### Autocomplete Interface Design
```go
type AutocompleteEngine interface {
    Search(query string, context CommandContext) []CommandSuggestion
    UpdateContext(context CommandContext)
    SetRankingStrategy(strategy RankingStrategy)
}
```

## Backward Compatibility

### Existing Features to Preserve
- All current built-in commands (Initialize Project, Compact Session)
- Custom command loading from markdown files
- Named argument system with `$VARIABLE` syntax
- Command prefixes (`user:`, `project:`)
- Existing keyboard shortcuts and navigation

### Migration Strategy
- Maintain existing `dialog.Command` struct during transition
- Create adapters between old and new command systems
- Gradual migration of command handlers
- Deprecation warnings for old APIs
- Comprehensive testing of existing workflows

## Testing Strategy

### Unit Tests
- Command registration and discovery
- Autocomplete search algorithms
- Argument validation and parsing
- Sub-command resolution
- Error handling scenarios

### Integration Tests
- Command execution workflows
- UI component interactions
- Keyboard navigation
- Custom command loading
- Cross-component communication

### Regression Tests
- All existing command functionality
- Custom command compatibility
- Performance benchmarks
- Memory usage validation

## Performance Considerations

### Optimization Targets
- Sub-millisecond autocomplete response times
- Efficient command indexing and search
- Minimal memory footprint for command metadata
- Lazy loading of large command sets
- Caching of frequently used commands

### Monitoring
- Command execution timing
- Autocomplete performance metrics
- Memory usage tracking
- User interaction analytics

## Documentation Updates

### User Documentation
- Updated command usage examples
- Autocomplete feature guide
- Sub-command syntax documentation
- Migration guide for custom commands

### Developer Documentation
- Command development guide
- Registry API documentation
- Extension patterns
- Best practices

## Risk Mitigation

### Potential Issues
- Performance degradation with large command sets
- UI complexity with nested commands
- Backward compatibility challenges
- User experience disruption during transition

### Mitigation Strategies
- Incremental feature rollout
- A/B testing for UI changes
- Comprehensive automated testing
- User feedback collection
- Rollback procedures

## Success Metrics

### User Experience
- Reduced time to find and execute commands
- Improved command discoverability
- Higher user satisfaction scores
- Reduced support requests

### Technical Metrics
- Faster command execution times
- Improved code maintainability
- Reduced bug reports
- Better test coverage

## Timeline

- **Phase 1 (Foundation)**: 1-2 weeks
- **Phase 2 (Registry)**: 2-3 weeks  
- **Phase 3 (Autocomplete)**: 2-3 weeks
- **Phase 4 (UI Components)**: 3-4 weeks
- **Phase 5 (Sub-commands)**: 2-3 weeks
- **Phase 6 (Integration)**: 2-3 weeks
- **Phase 7 (Advanced)**: 3-4 weeks

**Total Estimated Duration**: 15-22 weeks

## Next Steps

1. Complete Phase 1 foundation work
2. Get stakeholder approval for architecture
3. Begin Phase 2 implementation
4. Set up continuous integration for new components
5. Create detailed design documents for UI components