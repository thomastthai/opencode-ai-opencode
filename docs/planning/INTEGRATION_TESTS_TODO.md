# Integration Testing TODO

This document outlines comprehensive integration tests that should be implemented to ensure all OpenCode components work together correctly in realistic scenarios.

## Authentication & OAuth2 Integration

### OAuth2 Flow Integration
- [ ] End-to-end TUI OAuth2 dialog interaction with real keyboard input
- [ ] Dialog state transitions during actual OAuth2 flow
- [ ] Error handling in TUI when OAuth2 fails
- [ ] Provider integration with actual OAuth2 tokens
- [ ] Token refresh integration with Gemini API calls
- [ ] Authentication method switching (API key → OAuth2 → fallback)

### Config System Authentication
- [ ] Loading OAuth2 credentials from config files vs environment variables
- [ ] Config validation with invalid OAuth2 credentials
- [ ] Multi-provider authentication scenarios
- [ ] Authentication persistence across application restarts

### Cross-Platform Authentication
- [ ] XDG compliance on different operating systems
- [ ] File permission handling across platforms
- [ ] Browser opening mechanisms
- [ ] Token file corruption recovery

## Database Integration

### Session Management
- [ ] Complete session lifecycle (create, save, load, update, delete)
- [ ] Session persistence across application restarts
- [ ] Database migration scenarios
- [ ] Concurrent session access
- [ ] Session restoration with refreshed tokens
- [ ] Multi-user session handling

### Message Storage
- [ ] Large message handling and pagination
- [ ] Message search and filtering
- [ ] Database corruption recovery
- [ ] Auto-compaction with database state
- [ ] Token storage in SQLite database
- [ ] Session data with OAuth2 credentials
- [ ] Data migration with authentication changes

## TUI Integration

### Component Integration
- [ ] Complete UI workflows (chat → commands → model selection)
- [ ] Keyboard navigation across different dialogs
- [ ] State management between UI components
- [ ] Error handling in UI flows
- [ ] OAuth2 authentication persistence across sessions

### Editor Integration
- [ ] External editor integration (`Ctrl+E`)
- [ ] File change detection during editing
- [ ] Multi-line input handling
- [ ] Vim-like editor functionality

## LLM Provider Integration

### Multi-Provider Scenarios
- [ ] Provider switching during active sessions
- [ ] Fallback mechanisms when providers fail
- [ ] Rate limiting and retry logic
- [ ] Model availability checking
- [ ] Cross-provider authentication flows
- [ ] Token expiration handling during active sessions
- [ ] Provider failover scenarios

### Tool Integration
- [ ] AI agents using file system tools (`glob`, `grep`, `view`, `edit`)
- [ ] Shell command execution with different shells
- [ ] Sourcegraph API integration
- [ ] Tool permission workflows
- [ ] Real AI model interactions (with test accounts)

## Command System Integration

### Built-in Commands
- [ ] Complete command execution workflows
- [ ] Command aliases and discovery
- [ ] Command palette search functionality
- [ ] Recently used command tracking
- [ ] Built-in commands with OAuth2 authentication
- [ ] Command execution with expired tokens

### Custom Commands
- [ ] Loading commands from multiple directories
- [ ] YAML frontmatter parsing
- [ ] Parameterized command execution
- [ ] Command file watching/reloading
- [ ] Custom commands requiring authenticated API calls

## LSP Integration

### Language Server Integration
- [ ] LSP server lifecycle management
- [ ] Diagnostics integration with AI tools
- [ ] File watching and change notifications
- [ ] Multi-language server scenarios

### Diagnostics Tool
- [ ] Real diagnostics from actual language servers
- [ ] Error reporting integration
- [ ] Performance with large codebases

## MCP Server Integration

### MCP Protocol
- [ ] Stdio and SSE connection types
- [ ] Tool discovery and execution
- [ ] Server lifecycle management
- [ ] Error handling and reconnection
- [ ] OAuth2-authenticated MCP connections
- [ ] Token refresh for long-running MCP sessions
- [ ] MCP server failures during OAuth2 flow

### External Tool Integration
- [ ] Real MCP servers with actual tools
- [ ] Permission system with MCP tools
- [ ] Performance with multiple MCP servers

## Application Integration

### Non-Interactive Mode
- [ ] Complete prompt processing pipeline
- [ ] Output formatting (text/JSON)
- [ ] Permission auto-approval
- [ ] Error handling in headless mode

### Configuration System
- [ ] Multi-source config loading priority
- [ ] Environment variable override scenarios
- [ ] Config validation and error reporting
- [ ] Hot config reloading

### File System Integration
- [ ] Project initialization with `.opencode` directories
- [ ] File change tracking across sessions
- [ ] Working directory management
- [ ] Cross-platform path handling

## Network & Security Integration

### Network Resilience
- [ ] OAuth2 flow with network interruptions
- [ ] Token refresh with connectivity issues
- [ ] Callback server failures
- [ ] API rate limiting and retry logic
- [ ] Network failure recovery
- [ ] Proxy and firewall scenarios
- [ ] Concurrent API requests

### Security Testing
- [ ] Concurrent token access protection
- [ ] CSRF protection validation
- [ ] Secure token file permissions
- [ ] Authentication audit trails

## End-to-End Scenarios

### Complete User Workflows
- [ ] New user onboarding flow
- [ ] Complete coding session with multiple tools
- [ ] Project setup and configuration
- [ ] Session management across multiple projects
- [ ] Error recovery and user guidance

### Performance & Resource Management
- [ ] Performance testing with large projects
- [ ] Memory usage and resource management
- [ ] Concurrent user scenarios
- [ ] Large conversation handling

## Test Infrastructure Requirements

### Testing Framework Needs
- [ ] Mock servers for external APIs
- [ ] Test database fixtures
- [ ] TUI interaction simulation
- [ ] Cross-platform test environments
- [ ] Performance benchmarking tools
- [ ] Network simulation tools

### CI/CD Integration
- [ ] Automated integration test runs
- [ ] External service testing (with proper credentials)
- [ ] Cross-platform compatibility testing
- [ ] Performance regression detection

---

## Priority Levels

**High Priority (Core Functionality)**
- OAuth2 end-to-end flows
- Database session management
- TUI component integration
- Provider authentication scenarios

**Medium Priority (User Experience)**
- Command system workflows
- File system integration
- Error handling scenarios
- Cross-platform compatibility

**Lower Priority (Advanced Features)**
- MCP server integration
- LSP advanced scenarios
- Performance optimization tests
- Network resilience testing

---

*This document should be updated as integration tests are implemented and new integration scenarios are identified.*