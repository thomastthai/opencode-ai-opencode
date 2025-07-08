# Debugging Guide

This guide documents debugging techniques and strategies for troubleshooting issues in OpenCode.

## General Debugging Principles

### 1. Trace the Entire Flow
When debugging an issue, especially in event-driven systems like TUIs, it's crucial to trace the entire flow of data from input to output. Add logging at every major transition point to understand where things go wrong.

### 2. Use File-Based Logging for TUI Applications
Since TUI applications take over the terminal, traditional stderr logging isn't easily visible. Use file-based logging that can be monitored in a separate terminal:

```go
debugLog := fmt.Sprintf("[Component] Event: %q\n", data)
if f, err := os.OpenFile("/tmp/app-debug.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644); err == nil {
    f.WriteString(debugLog)
    f.Close()
}
```

Monitor with: `tail -f /tmp/app-debug.log`

### 3. Log State Before and After Changes
Always log the state before and after modifications to identify unexpected transformations:

```go
logging.Debug("[Component] Before update:", "value", currentValue)
// ... perform update ...
logging.Debug("[Component] After update:", "value", newValue)
```

## Case Study: Tab Completion Double Slash Bug

### Problem
When users typed `/session<tab>`, backspaced everything, then typed `/se<tab>`, the result was `//se` instead of `/session`.

### Debugging Approach

1. **Added logging at the tab handler entry point** to see what input was received:
   ```go
   logging.Debug("[HandleTabKey] Input from pseudoSearchTextArea:", "input", input)
   ```

2. **Traced through the completion provider** to see what it returned:
   ```go
   logging.Debug("[HandleTabKey] After HandleTabCompletion:", "completed", completed)
   ```

3. **Logged every keystroke in the completion dialog** to understand state changes:
   ```go
   logging.Debug("[completionDialog] TextArea updated:", "value", fullValue, "key", msg.String())
   ```

4. **Added specific backspace handling logs** to track state during deletion:
   ```go
   logging.Debug("[completionDialog] After backspace:", "value", currentValue)
   ```

### Key Discovery
The logs revealed that after backspacing to `/`, the completion dialog remained open. When the user typed `/` again, it was appended to the existing `/`, creating `//`.

### Solution
Modified the backspace handler to close the dialog when only the trigger character remains:

```go
if len(currentValue) == 1 && (currentValue == "/" || currentValue == "@") {
    return c, c.close()
}
```

## Debugging Strategies

### 1. Strategic Logging Placement
Place logs at:
- Entry points of event handlers
- Before and after state mutations
- At decision points (if/switch statements)
- Error boundaries
- Message passing points

### 2. Include Context in Logs
Always include relevant context:
```go
logging.Debug("[Component] Event processed", 
    "event", eventType,
    "currentState", state,
    "input", input,
    "result", result)
```

### 3. Use Temporary Info-Level Logs
If debug logs aren't showing up, temporarily use info-level logs:
```go
// Temporary for debugging - change back to Debug before committing
logging.Info("[DEBUG] Critical value:", "value", value)
```

### 4. Create Reproducible Test Cases
Document the exact steps to reproduce the issue:
1. Start the application
2. Type specific sequence
3. Note expected vs actual behavior

### 5. Binary Search for Issues
If you have a complex flow, use binary search:
1. Add a log in the middle of the flow
2. If the issue is before, focus on the first half
3. If the issue is after, focus on the second half
4. Repeat until you find the exact location

## Common Debugging Patterns

### Event Flow Tracing
```go
// At event source
logging.Debug("[Source] Sending event", "type", eventType, "data", data)

// At event handler
logging.Debug("[Handler] Received event", "type", eventType, "data", data)

// At event completion
logging.Debug("[Handler] Completed event", "type", eventType, "result", result)
```

### State Transition Debugging
```go
logging.Debug("[StateMachine] Transition", 
    "from", oldState, 
    "to", newState, 
    "trigger", trigger)
```

### Message Passing Debug
```go
// Before sending
logging.Debug("[Sender] Sending message", "type", msgType, "content", content)

// After receiving
logging.Debug("[Receiver] Received message", "type", msgType, "content", content)
```

## Tools and Commands

### Viewing Logs in OpenCode
- Press `Ctrl+L` to view logs (Note: This may conflict with tmux)
- Logs are stored in memory and can be viewed in the TUI

### External Log Monitoring
```bash
# Watch debug log file
tail -f /tmp/opencode-debug.log

# Filter for specific components
tail -f /tmp/opencode-debug.log | grep HandleTabKey

# Watch with highlighting
tail -f /tmp/opencode-debug.log | grep --color=always -E "ERROR|WARNING|"
```

### Debug Build Flags
```bash
# Run with debug logging
./opencode -d

# Run with development debug (logs to file)
OPENCODE_DEV_DEBUG=true ./opencode
```

## Best Practices

1. **Clean up debug logs** after fixing the issue
2. **Convert useful debug logs** to permanent debug-level logs
3. **Document the debugging process** for future reference
4. **Add tests** to prevent regression
5. **Consider adding permanent instrumentation** for frequently problematic areas

## Troubleshooting Specific Issues

### TUI Key Handling Issues
- Log key events at multiple levels (raw input, processed input, final action)
- Check for key binding conflicts
- Verify focus state of components

### Completion/Autocomplete Issues
- Log the full query at each transformation
- Track dialog open/close states
- Monitor text area value changes

### State Synchronization Issues
- Log all state updates with timestamps
- Track message flow between components
- Verify event ordering

Remember: Systematic debugging with comprehensive logging will help identify issues faster than guessing or making random changes.