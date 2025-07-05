# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

**Build and Run:**
```bash
go build -o opencode                    # Build the main binary
./opencode                              # Run the application
./opencode -p "prompt text"             # Run in non-interactive mode
```

**Testing:**
```bash
go test ./...                           # Run all tests
go test ./internal/commands             # Run tests for specific package
go test -v ./internal/commands -run TestSpecificTest  # Run single test with verbose output
```

**Development:**
```bash
go run main.go                          # Run without building
go mod tidy                             # Clean up dependencies
go vet ./...                            # Static analysis
go fmt ./...                            # Format code
```

## Architecture Overview

OpenCode is a terminal-based AI assistant built with Go, featuring:

- **cmd/**: CLI interface using Cobra framework
- **internal/app/**: Core application orchestration and non-interactive mode
- **internal/llm/**: LLM provider abstractions, agent system, and tool integrations
  - **agent/**: AI agent orchestration with tool calling capabilities
  - **provider/**: Multiple AI provider implementations (OpenAI, Anthropic, Gemini, etc.)
  - **tools/**: Built-in tools for file operations, shell execution, diagnostics
- **internal/tui/**: Terminal UI built with Bubble Tea framework
- **internal/commands/**: Command registry system for built-in and custom commands
- **internal/config/**: Configuration management with provider-specific settings
- **internal/db/**: SQLite persistence for sessions and messages
- **internal/lsp/**: Language Server Protocol client integration

## Key Patterns

- **Provider Interface**: All LLM providers implement `Provider` interface with optional extensions (StreamProvider, ToolCallingProvider, etc.)
- **Agent System**: Tools are registered with agents and executed via permission-controlled calls
- **Command System**: Built-in commands use builder pattern, custom commands loaded from markdown files
- **TUI Architecture**: Component-based UI with overlay dialogs and page navigation
- **Session Management**: Database-persisted conversations with auto-compaction

## Development Notes

- The codebase currently has build errors in `internal/llm/agent/agent-tool.go` - undefined variables need fixing
- Built-in commands are registered via `RegisterBuiltIn()` in init functions
- Custom commands are loaded from `.opencode/commands/` directories as markdown files
- LSP integration provides diagnostics but full protocol support is implemented
- MCP (Model Context Protocol) support for external tool integration