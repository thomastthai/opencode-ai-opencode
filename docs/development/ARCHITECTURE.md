# Architecture and Coding Principles

This document outlines the core architectural patterns and coding principles for OpenCode. These guidelines should be followed when creating new features or modifying existing code to ensure maintainability, testability, and consistency.

## Core Architecture

### Layered Architecture

```
┌─────────────────────────────────────────┐
│             TUI Layer                   │
│  (Bubble Tea components, pages, dialogs)│
├─────────────────────────────────────────┤
│          Application Layer              │
│    (App orchestration, services)        │
├─────────────────────────────────────────┤
│           Domain Layer                  │
│  (Agents, sessions, messages, tools)    │
├─────────────────────────────────────────┤
│         Infrastructure Layer            │
│     (DB, LSP, providers, config)        │
└─────────────────────────────────────────┘
```

### Key Architectural Patterns

1. **Service Pattern**: Core functionality is exposed through service interfaces (SessionService, MessageService, etc.)
2. **Provider Pattern**: Extensible implementations for LLM providers, completion providers, etc.
3. **Command Pattern**: Executable commands with consistent interface for both built-in and custom commands
4. **Observer Pattern**: Pub/sub for session and message updates
5. **Factory Pattern**: Consistent initialization of complex objects

## Coding Principles

### 1. State Management

**Principle**: State transitions must be atomic and complete.

When implementing state changes:
- Update ALL related state in a single operation
- Never leave the system in a partial state
- Validate state transitions

```go
// ❌ Bad: Partial state update
func (c *CompletionDialog) SetProvider(p Provider) {
    c.provider = p  // Query, items, and UI strings are now stale!
}

// ✅ Good: Complete state update
func (c *CompletionDialog) SetProvider(p Provider) {
    c.provider = p
    c.query = ""  // Reset query for new provider
    c.items, _ = p.GetChildEntries("")  // Load items
    c.listView.SetItems(c.items)
    c.listView.SetEmptyMessage(p.GetEmptyMessage())
}
```

### 2. Defensive Programming

**Principle**: Assume nothing, validate everything.

#### Nil Safety
```go
// ❌ Bad: Assumes non-nil
func (p *ChatPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if p.app.Recovery.IsRecovering() {  // Panic if Recovery is nil
        return p, nil
    }
}

// ✅ Good: Defensive check
func (p *ChatPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if p.app.Recovery != nil && p.app.Recovery.IsRecovering() {
        return p, nil
    }
}
```

#### Constructor Validation
```go
// ✅ Best: Validate in constructor
func NewChatPage(app *App) *ChatPage {
    if app == nil {
        panic("app cannot be nil")
    }
    if app.Recovery == nil {
        panic("app.Recovery is required")
    }
    return &ChatPage{app: app}
}
```

### 3. Explicit Dependencies

**Principle**: Make all dependencies visible and injectable.

```go
// ❌ Bad: Hidden dependencies
func NewDialog() *Dialog {
    return &Dialog{
        config: config.Get(),  // Hidden global dependency
    }
}

// ✅ Good: Explicit dependencies
func NewDialog(cfg *Config) *Dialog {
    if cfg == nil {
        panic("config is required")
    }
    return &Dialog{
        config: cfg,
    }
}
```

### 4. Testability by Design

**Principle**: Design interfaces and structures that are easy to test.

#### Observable State
```go
type CompletionDialog interface {
    // Core functionality
    SetProvider(p Provider)
    Update(msg tea.Msg) (tea.Model, tea.Cmd)
    View() string
    
    // Observable state for testing
    GetProvider() Provider
    GetItems() []Item
    GetEmptyMessage() string
}
```

#### Separation of Concerns
```go
// ✅ Good: Separate business logic from UI
type CompletionLogic struct {
    provider Provider
    items    []Item
    query    string
}

func (l *CompletionLogic) SetProvider(p Provider) {
    // Pure business logic, easy to test
}

type CompletionDialog struct {
    logic *CompletionLogic
    // UI-specific fields
}
```

### 5. Error Handling

**Principle**: Errors should be handled explicitly and appropriately.

```go
// ❌ Bad: Silently ignoring errors
items, _ := provider.GetChildEntries(query)

// ✅ Good: Handle or propagate errors
items, err := provider.GetChildEntries(query)
if err != nil {
    logging.Error("Failed to get entries", err)
    // Decide: return error, use fallback, or show error state
    return c, c.showError(err)
}
```

### 6. Factory Functions

**Principle**: Use factory functions to ensure proper initialization.

```go
// ✅ Good: Factory ensures complete initialization
func NewApp(ctx context.Context, db *sql.DB) (*App, error) {
    sessions := session.NewService(db)
    messages := message.NewService(db)
    recovery := recovery.NewService()
    
    app := &App{
        Sessions: sessions,
        Messages: messages,
        Recovery: recovery,
        // All required fields initialized
    }
    
    // Additional setup
    if err := app.initialize(ctx); err != nil {
        return nil, fmt.Errorf("failed to initialize app: %w", err)
    }
    
    return app, nil
}
```

## Component Guidelines

### TUI Components

1. **State Isolation**: Each component manages its own state
2. **Message Passing**: Components communicate via Bubble Tea messages
3. **Immutability**: Update methods return new instances
4. **View Purity**: View() methods should be pure functions of state

### Service Layer

1. **Interface First**: Define service interfaces before implementations
2. **Context Awareness**: All service methods should accept context.Context
3. **Error Transparency**: Return meaningful errors, don't hide failures
4. **Thread Safety**: Services must be safe for concurrent use

### Provider Implementations

1. **Interface Compliance**: Implement all required provider interfaces
2. **Capability Declaration**: Use capability interfaces (StreamProvider, ToolCallProvider)
3. **Configuration Validation**: Validate configuration in constructor
4. **Resource Management**: Properly close connections and clean up resources

## Common Patterns

### The Complete State Update Pattern

When implementing methods that change state:

```go
func (c *Component) TransitionTo(newState State) error {
    // 1. Validate transition
    if !c.canTransitionTo(newState) {
        return ErrInvalidTransition
    }
    
    // 2. Prepare new state
    oldState := c.state
    
    // 3. Update ALL related fields atomically
    c.state = newState
    c.updateDependentFields(newState)
    c.clearStaleData(oldState)
    
    // 4. Notify observers
    c.notifyStateChange(oldState, newState)
    
    return nil
}
```

### The Safe Initialization Pattern

For complex objects with multiple dependencies:

```go
type ComponentOptions struct {
    Required1 Service1  // Required
    Required2 Service2  // Required
    Optional1 *Config   // Optional
}

func NewComponent(opts ComponentOptions) (*Component, error) {
    // Validate required dependencies
    if opts.Required1 == nil {
        return nil, errors.New("Required1 is required")
    }
    if opts.Required2 == nil {
        return nil, errors.New("Required2 is required")
    }
    
    // Apply defaults for optional
    if opts.Optional1 == nil {
        opts.Optional1 = DefaultConfig()
    }
    
    return &Component{
        service1: opts.Required1,
        service2: opts.Required2,
        config:   opts.Optional1,
    }, nil
}
```

### The Provider Switch Pattern

When implementing provider/strategy switching:

```go
func (c *Component) SwitchProvider(provider Provider) error {
    if provider == nil {
        return errors.New("provider cannot be nil")
    }
    
    // Store old state for rollback
    oldProvider := c.provider
    oldItems := c.items
    
    // Attempt switch
    c.provider = provider
    
    // Update all dependent state
    items, err := provider.Initialize()
    if err != nil {
        // Rollback on error
        c.provider = oldProvider
        c.items = oldItems
        return fmt.Errorf("failed to initialize provider: %w", err)
    }
    
    // Complete the switch
    c.items = items
    c.resetUserState()  // Clear queries, selections, etc.
    c.updateUI()
    
    return nil
}
```

## Anti-Patterns to Avoid

### 1. Partial State Updates
```go
// ❌ Don't do this
c.provider = newProvider
// Forget to update items, query, UI...
```

### 2. Hidden Global Dependencies
```go
// ❌ Don't do this
func NewService() *Service {
    return &Service{
        db: global.GetDB(),  // Hidden dependency
    }
}
```

### 3. Defensive Programming Theatre
```go
// ❌ Don't just check and continue
if app.Recovery == nil {
    // Log and continue as if nothing is wrong
    log.Println("warning: Recovery is nil")
}
app.Recovery.DoSomething()  // Still crashes!
```

### 4. Test-Only Interfaces
```go
// ❌ Don't add methods just for tests
type Component interface {
    // ... production methods ...
    
    // Test-only methods cluttering the interface
    SetTestMode(bool)
    GetInternalState() internalState
}
```

Instead, design observable interfaces that are useful for both production monitoring and testing.

## Summary

Following these principles leads to code that is:
- **Predictable**: State transitions are complete and validated
- **Testable**: Dependencies are explicit and state is observable  
- **Maintainable**: Clear patterns and consistent structure
- **Reliable**: Defensive programming prevents runtime panics
- **Debuggable**: Clear error handling and state visibility

When in doubt, favor:
- Explicit over implicit
- Complete operations over partial updates
- Validation over assumption
- Observable state over hidden internals