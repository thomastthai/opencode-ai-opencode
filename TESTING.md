# Testing Guidelines for OpenCode

This document outlines testing best practices and patterns learned from real bugs in the codebase. LLMs and developers should follow these guidelines when creating or modifying tests.

## Core Testing Principles

### 1. Test User-Visible Behavior, Not Just Internal State

**❌ Bad: Testing only internal state**
```go
// This test passed but missed the bug where users saw "No file matches found" instead of "No commands found"
assert.Equal(t, "commands", p.completionDialog.GetId(), "Provider should be 'commands'")
```

**✅ Good: Testing what users actually see**
```go
assert.Equal(t, "commands", p.completionDialog.GetId(), "Provider should be 'commands'")
assert.Equal(t, "No commands found", p.completionDialog.GetEmptyMessage(), "Empty message should be for commands")
assert.Contains(t, p.completionDialog.View(), "No commands found", "View should display correct message")
```

### 2. Test State Transitions Completely

When testing state changes, verify ALL related state updates, not just the primary change.

**❌ Bad: Incomplete state transition testing**
```go
// Only tests that provider changed, not that UI updated
p.SetProvider(commandProvider)
assert.Equal(t, "commands", p.GetId())
```

**✅ Good: Complete state transition testing**
```go
p.SetProvider(commandProvider)
assert.Equal(t, "commands", p.GetId())
assert.Greater(t, len(p.GetListItems()), 0, "Should have loaded command items")
assert.Equal(t, "No commands found", p.GetEmptyMessage())
```

### 3. Initialize Test Objects Properly

Always initialize test objects with all required dependencies to avoid nil pointer panics.

**❌ Bad: Minimal test setup**
```go
testApp := &app.App{} // Missing required services
model := New(testApp)
```

**✅ Good: Complete test setup**
```go
testApp := &app.App{
    Recovery: recovery.NewService(),
    Sessions: session.NewMockService(),
    // ... other required services
}
model := New(testApp)
```

## State Management Testing

### Testing Provider/Strategy Switches

When testing components that switch between different providers or strategies:

1. Test the initial state
2. Test the state after switching
3. Test switching back to the original state
4. Test rapid switches
5. Test edge cases (nil providers, empty results)

Example:
```go
// Test complete provider switch cycle
func TestProviderSwitching(t *testing.T) {
    dialog := NewCompletionDialog()
    
    // Initial state with command provider
    dialog.SetProvider(commandProvider)
    assert.Equal(t, "No commands found", dialog.GetEmptyMessage())
    
    // Switch to file provider
    dialog.SetProvider(fileProvider)
    assert.Equal(t, "No file matches found", dialog.GetEmptyMessage())
    
    // Switch back - ensure state is fully restored
    dialog.SetProvider(commandProvider)
    assert.Equal(t, "No commands found", dialog.GetEmptyMessage())
}
```

### Testing View/UI Output

Always test the actual rendered output when testing UI components:

```go
func TestEmptyStateMessages(t *testing.T) {
    dialog := NewCompletionDialog()
    dialog.SetProvider(commandProvider)
    
    // Force empty state
    dialog.SetItems([]Item{})
    
    // Test the actual view output
    view := dialog.View()
    assert.Contains(t, view, "No commands found")
    assert.NotContains(t, view, "No file matches found")
}
```

## Defensive Programming in Tests

### 1. Add Helper Methods for Testing

Add methods to your interfaces specifically for testing internal state:

```go
type CompletionDialog interface {
    // Production methods
    SetProvider(provider Provider)
    View() string
    
    // Test helper methods
    GetListItems() []Item        // For verifying items loaded
    GetEmptyMessage() string      // For verifying correct message
}
```

### 2. Use Factory Functions

Create factory functions that ensure proper initialization:

```go
// Instead of allowing partial initialization
func NewTestApp() *App {
    return &App{
        Recovery: recovery.NewService(),
        Sessions: session.NewMockService(),
        Messages: message.NewMockService(),
        // ... all required fields
    }
}
```

### 3. Validate State Transitions

Add validation in your state transition methods:

```go
func (c *Component) SetProvider(provider Provider) {
    if provider == nil {
        panic("provider cannot be nil")
    }
    
    c.provider = provider
    c.updateDependentState() // Ensure all related state is updated
}
```

## Common Testing Patterns

### 1. The "Backspace and Retype" Pattern

This pattern catches state management bugs in interactive UIs:

```go
func TestBackspaceAndRetype(t *testing.T) {
    // Type something
    editor.Type("/list")
    assert.Equal(t, expectedState1, component.GetState())
    
    // Clear everything
    editor.Clear()
    assert.Equal(t, expectedState2, component.GetState())
    
    // Type again - this often reveals stale state bugs
    editor.Type("/")
    assert.Equal(t, expectedState3, component.GetState())
}
```

### 2. The "Round Trip" Pattern

Test that you can go from A to B and back to A:

```go
func TestRoundTrip(t *testing.T) {
    initialState := component.GetState()
    
    component.DoAction()
    modifiedState := component.GetState()
    assert.NotEqual(t, initialState, modifiedState)
    
    component.UndoAction()
    finalState := component.GetState()
    assert.Equal(t, initialState, finalState, "Should return to initial state")
}
```

### 3. The "Integration Smoke Test" Pattern

Even in unit tests, do basic integration checks:

```go
func TestComponentIntegration(t *testing.T) {
    // Don't just test the component in isolation
    app := NewTestApp()
    component := app.NewComponent()
    
    // Verify it integrates with the app properly
    assert.NotNil(t, component.GetSession())
    assert.NotNil(t, component.GetProvider())
    
    // Do a basic operation that touches multiple systems
    err := component.PerformAction()
    assert.NoError(t, err)
}
```

## Testing Checklist

When writing or reviewing tests, ensure:

- [ ] All user-visible outputs are tested (View(), Render(), etc.)
- [ ] State transitions test all affected state, not just primary changes
- [ ] Test objects are properly initialized with all dependencies
- [ ] Edge cases are tested (nil, empty, invalid inputs)
- [ ] Integration points are tested, not just isolated units
- [ ] The test would have caught known historical bugs

## Anti-Patterns to Avoid

1. **Testing only the happy path** - Always test error cases and edge conditions
2. **Over-mocking** - Too many mocks can hide integration issues
3. **Testing implementation details** - Test behavior, not how it's implemented
4. **Ignoring timing/concurrency** - Test concurrent operations when applicable
5. **Not testing the actual user experience** - A green test suite doesn't mean the UX works

## Example: Complete Test for State Management

Here's an example that incorporates all these principles:

```go
func TestCompletionDialogStateManagement(t *testing.T) {
    // Proper initialization
    app := NewTestApp()
    dialog := NewCompletionDialog(app)
    
    // Test initial state
    t.Run("initial state", func(t *testing.T) {
        assert.False(t, dialog.IsVisible())
        assert.Empty(t, dialog.GetListItems())
    })
    
    // Test provider switching
    t.Run("provider switching", func(t *testing.T) {
        // Set command provider
        dialog.SetProvider(commandProvider)
        assert.Equal(t, "commands", dialog.GetId())
        assert.Equal(t, "No commands found", dialog.GetEmptyMessage())
        
        // Verify view output
        view := dialog.View()
        assert.Contains(t, view, "No commands found")
        
        // Switch to file provider
        dialog.SetProvider(fileProvider)
        assert.Equal(t, "files", dialog.GetId())
        assert.Equal(t, "No file matches found", dialog.GetEmptyMessage())
        
        // Verify view updated
        view = dialog.View()
        assert.Contains(t, view, "No file matches found")
        assert.NotContains(t, view, "No commands found")
    })
    
    // Test edge cases
    t.Run("edge cases", func(t *testing.T) {
        // Nil provider should not panic
        assert.NotPanics(t, func() {
            dialog.SetProvider(nil)
        })
        
        // Rapid switching
        for i := 0; i < 10; i++ {
            if i%2 == 0 {
                dialog.SetProvider(commandProvider)
            } else {
                dialog.SetProvider(fileProvider)
            }
        }
        // Should end in file provider state
        assert.Equal(t, "files", dialog.GetId())
    })
}
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestCompletionDialog ./internal/tui/...

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...
```

## Continuous Improvement

This document should be updated whenever:
- A bug reveals a gap in our testing approach
- New testing patterns prove effective
- Common mistakes are identified in code reviews

Remember: **A test that doesn't catch bugs is just maintenance overhead.**