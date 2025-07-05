> [!NOTE]  
> This is the original OpenCode repository, now continuing at [Charm](https://github.com/charmbracelet) with its original creator, [Kujtim Hoxha](https://github.com/kujtimiihoxha).  
> Development is continuing under a new name as we prepare for a public relaunch.  
> Follow [@charmcli](https://x.com/charmcli) or join our [Discord](https://charm.sh/chat) for updates.

# ⌬ OpenCode

<p align="center"><img src="https://github.com/user-attachments/assets/9ae61ef6-70e5-4876-bc45-5bcb4e52c714" width="800"></p>

> **⚠️ Early Development Notice:** This project is in early development and is not yet ready for production use. Features may change, break, or be incomplete. Use at your own risk.

A powerful terminal-based AI assistant for developers, providing intelligent coding assistance directly in your terminal.

## Overview

OpenCode is a Go-based CLI application that brings AI assistance to your terminal. It provides a TUI (Terminal User Interface) for interacting with various AI models to help with coding tasks, debugging, and more.

<p>For a quick video overview, check out
<a href="https://www.youtube.com/watch?v=P8luPmEa1QI"><img width="25" src="https://upload.wikimedia.org/wikipedia/commons/0/09/YouTube_full-color_icon_%282017%29.svg"> OpenCode + Gemini 2.5 Pro: BYE Claude Code! I'm SWITCHING To the FASTEST AI Coder!</a></p>

<a href="https://www.youtube.com/watch?v=P8luPmEa1QI"><img width="550" src="https://i3.ytimg.com/vi/P8luPmEa1QI/maxresdefault.jpg"></a><p>

## Features

- **Interactive TUI**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a smooth terminal experience
- **Enhanced Command System**: Powerful slash command interface with search, categorization, and visual grouping
- **Multiple AI Providers**: Support for OpenAI, Anthropic Claude, Google Gemini, AWS Bedrock, Groq, Azure OpenAI, and OpenRouter
- **OAuth2 Authentication**: Seamless OAuth2 integration with visual TUI dialogs for secure authentication
- **Session Management**: Save and manage multiple conversation sessions
- **Tool Integration**: AI can execute commands, search files, and modify code
- **Vim-like Editor**: Integrated editor with text input capabilities
- **Persistent Storage**: SQLite database for storing conversations and sessions
- **LSP Integration**: Language Server Protocol support for code intelligence
- **File Change Tracking**: Track and visualize file changes during sessions
- **External Editor Support**: Open your preferred editor for composing messages
- **Custom Commands**: Create powerful custom commands with multiple named placeholders

## Installation

### Using the Install Script

```bash
# Install the latest version
curl -fsSL https://raw.githubusercontent.com/opencode-ai/opencode/refs/heads/main/install | bash

# Install a specific version
curl -fsSL https://raw.githubusercontent.com/opencode-ai/opencode/refs/heads/main/install | VERSION=0.1.0 bash
```

### Using Homebrew (macOS and Linux)

```bash
brew install opencode-ai/tap/opencode
```

### Using AUR (Arch Linux)

```bash
# Using yay
yay -S opencode-ai-bin

# Using paru
paru -S opencode-ai-bin
```

### Using Go

```bash
go install github.com/opencode-ai/opencode@latest
```

## Configuration

OpenCode looks for configuration in the following locations:

- `$HOME/.opencode.json`
- `$XDG_CONFIG_HOME/opencode/.opencode.json`
- `./.opencode.json` (local directory)

### Auto Compact Feature

OpenCode includes an auto compact feature that automatically summarizes your conversation when it approaches the model's context window limit. When enabled (default setting), this feature:

- Monitors token usage during your conversation
- Automatically triggers summarization when usage reaches 95% of the model's context window
- Creates a new session with the summary, allowing you to continue your work without losing context
- Helps prevent "out of context" errors that can occur with long conversations

You can enable or disable this feature in your configuration file:

```json
{
  "autoCompact": true // default is true
}
```

### Environment Variables

You can configure OpenCode using environment variables:

| Environment Variable       | Purpose                                                                          |
| -------------------------- | -------------------------------------------------------------------------------- |
| `ANTHROPIC_API_KEY`        | For Claude models                                                                |
| `OPENAI_API_KEY`           | For OpenAI models                                                                |
| `GEMINI_API_KEY`           | For Google Gemini models                                                         |
| `GEMINI_TOKEN`             | For Gemini OAuth2 authentication (see [Gemini OAuth](#gemini-oauth-authentication)) |
| `GEMINI_OAUTH_CLIENT_ID`   | OAuth2 Client ID for Gemini authentication                                      |
| `GEMINI_OAUTH_CLIENT_SECRET` | OAuth2 Client Secret for Gemini authentication                                |
| `GITHUB_TOKEN`             | For Github Copilot models (see [Using Github Copilot](#using-github-copilot))    |
| `VERTEXAI_PROJECT`         | For Google Cloud VertexAI (Gemini)                                               |
| `VERTEXAI_LOCATION`        | For Google Cloud VertexAI (Gemini)                                               |
| `GROQ_API_KEY`             | For Groq models                                                                  |
| `AWS_ACCESS_KEY_ID`        | For AWS Bedrock (Claude)                                                         |
| `AWS_SECRET_ACCESS_KEY`    | For AWS Bedrock (Claude)                                                         |
| `AWS_REGION`               | For AWS Bedrock (Claude)                                                         |
| `AZURE_OPENAI_ENDPOINT`    | For Azure OpenAI models                                                          |
| `AZURE_OPENAI_API_KEY`     | For Azure OpenAI models (optional when using Entra ID)                           |
| `AZURE_OPENAI_API_VERSION` | For Azure OpenAI models                                                          |
| `LOCAL_ENDPOINT`           | For self-hosted models                                                           |
| `SHELL`                    | Default shell to use (if not specified in config)                                |

### Shell Configuration

OpenCode allows you to configure the shell used by the bash tool. By default, it uses the shell specified in the `SHELL` environment variable, or falls back to `/bin/bash` if not set.

You can override this in your configuration file:

```json
{
  "shell": {
    "path": "/bin/zsh",
    "args": ["-l"]
  }
}
```

This is useful if you want to use a different shell than your default system shell, or if you need to pass specific arguments to the shell.

### Configuration File Structure

```json
{
  "data": {
    "directory": ".opencode"
  },
  "providers": {
    "openai": {
      "apiKey": "your-api-key",
      "disabled": false
    },
    "anthropic": {
      "apiKey": "your-api-key",
      "disabled": false
    },
    "gemini": {
      "apiKey": "your-api-key",
      "authMethod": "auto",
      "oauth2": {
        "clientId": "your-client-id.apps.googleusercontent.com",
        "clientSecret": "your-client-secret"
      },
      "disabled": false
    },
    "copilot": {
      "disabled": false
    },
    "groq": {
      "apiKey": "your-api-key",
      "disabled": false
    },
    "openrouter": {
      "apiKey": "your-api-key",
      "disabled": false
    }
  },
  "agents": {
    "coder": {
      "model": "claude-3.7-sonnet",
      "maxTokens": 5000
    },
    "task": {
      "model": "claude-3.7-sonnet",
      "maxTokens": 5000
    },
    "title": {
      "model": "claude-3.7-sonnet",
      "maxTokens": 80
    }
  },
  "shell": {
    "path": "/bin/bash",
    "args": ["-l"]
  },
  "mcpServers": {
    "example": {
      "type": "stdio",
      "command": "path/to/mcp-server",
      "env": [],
      "args": []
    }
  },
  "lsp": {
    "go": {
      "disabled": false,
      "command": "gopls"
    }
  },
  "debug": false,
  "debugLSP": false,
  "autoCompact": true
}
```

## Supported AI Models

OpenCode supports a variety of AI models from different providers:

### OpenAI

- GPT-4.1 family (gpt-4.1, gpt-4.1-mini, gpt-4.1-nano)
- GPT-4.5 Preview
- GPT-4o family (gpt-4o, gpt-4o-mini)
- O1 family (o1, o1-pro, o1-mini)
- O3 family (o3, o3-mini)
- O4 Mini

### Anthropic

- Claude 4 Sonnet
- Claude 4 Opus
- Claude 3.5 Sonnet
- Claude 3.5 Haiku
- Claude 3.7 Sonnet
- Claude 3 Haiku
- Claude 3 Opus

### GitHub Copilot

- GPT-3.5 Turbo
- GPT-4
- GPT-4o
- GPT-4o Mini
- GPT-4.1
- Claude 3.5 Sonnet
- Claude 3.7 Sonnet
- Claude 3.7 Sonnet Thinking
- Claude Sonnet 4
- O1
- O3 Mini
- O4 Mini
- Gemini 2.0 Flash
- Gemini 2.5 Pro

### Google

- Gemini 2.5
- Gemini 2.5 Flash
- Gemini 2.0 Flash
- Gemini 2.0 Flash Lite

### AWS Bedrock

- Claude 3.7 Sonnet

### Groq

- Llama 4 Maverick (17b-128e-instruct)
- Llama 4 Scout (17b-16e-instruct)
- QWEN QWQ-32b
- Deepseek R1 distill Llama 70b
- Llama 3.3 70b Versatile

### Azure OpenAI

- GPT-4.1 family (gpt-4.1, gpt-4.1-mini, gpt-4.1-nano)
- GPT-4.5 Preview
- GPT-4o family (gpt-4o, gpt-4o-mini)
- O1 family (o1, o1-mini)
- O3 family (o3, o3-mini)
- O4 Mini

### Google Cloud VertexAI

- Gemini 2.5
- Gemini 2.5 Flash

## Usage

```bash
# Start OpenCode
opencode

# Start with debug logging
opencode -d

# Start with a specific working directory
opencode -c /path/to/project
```

## Non-interactive Prompt Mode

You can run OpenCode in non-interactive mode by passing a prompt directly as a command-line argument. This is useful for scripting, automation, or when you want a quick answer without launching the full TUI.

```bash
# Run a single prompt and print the AI's response to the terminal
opencode -p "Explain the use of context in Go"

# Get response in JSON format
opencode -p "Explain the use of context in Go" -f json

# Run without showing the spinner (useful for scripts)
opencode -p "Explain the use of context in Go" -q
```

In this mode, OpenCode will process your prompt, print the result to standard output, and then exit. All permissions are auto-approved for the session.

By default, a spinner animation is displayed while the model is processing your query. You can disable this spinner with the `-q` or `--quiet` flag, which is particularly useful when running OpenCode from scripts or automated workflows.

### Output Formats

OpenCode supports the following output formats in non-interactive mode:

| Format | Description                     |
| ------ | ------------------------------- |
| `text` | Plain text output (default)     |
| `json` | Output wrapped in a JSON object |

The output format is implemented as a strongly-typed `OutputFormat` in the codebase, ensuring type safety and validation when processing outputs.

## Command-line Flags

| Flag              | Short | Description                                         |
| ----------------- | ----- | --------------------------------------------------- |
| `--help`          | `-h`  | Display help information                            |
| `--debug`         | `-d`  | Enable debug mode                                   |
| `--cwd`           | `-c`  | Set current working directory                       |
| `--prompt`        | `-p`  | Run a single prompt in non-interactive mode         |
| `--output-format` | `-f`  | Output format for non-interactive mode (text, json) |
| `--quiet`         | `-q`  | Hide spinner in non-interactive mode                |

## Keyboard Shortcuts

### Global Shortcuts

| Shortcut | Action                                                  |
| -------- | ------------------------------------------------------- |
| `Ctrl+C` | Quit application                                        |
| `Ctrl+?` | Toggle help dialog                                      |
| `?`      | Toggle help dialog (when not in editing mode)           |
| `Ctrl+L` | View logs                                               |
| `Ctrl+A` | Switch session                                          |
| `Ctrl+K` | Enhanced command palette with search and categorization |
| `Ctrl+O` | Toggle model selection dialog                           |
| `Esc`    | Close current overlay/dialog or return to previous mode |

### Chat Page Shortcuts

| Shortcut | Action                                  |
| -------- | --------------------------------------- |
| `Ctrl+N` | Create new session                      |
| `Ctrl+X` | Cancel current operation/generation     |
| `i`      | Focus editor (when not in writing mode) |
| `Esc`    | Exit writing mode and focus messages    |

### Editor Shortcuts

| Shortcut            | Action                                    |
| ------------------- | ----------------------------------------- |
| `Ctrl+S`            | Send message (when editor is focused)     |
| `Enter` or `Ctrl+S` | Send message (when editor is not focused) |
| `Ctrl+E`            | Open external editor                      |
| `Esc`               | Blur editor and focus messages            |

### Session Dialog Shortcuts

| Shortcut   | Action           |
| ---------- | ---------------- |
| `↑` or `k` | Previous session |
| `↓` or `j` | Next session     |
| `Enter`    | Select session   |
| `Esc`      | Close dialog     |

### Model Dialog Shortcuts

| Shortcut   | Action            |
| ---------- | ----------------- |
| `↑` or `k` | Move up           |
| `↓` or `j` | Move down         |
| `←` or `h` | Previous provider |
| `→` or `l` | Next provider     |
| `Esc`      | Close dialog      |

### Command Dialog Shortcuts

| Shortcut        | Action                                    |
| --------------- | ----------------------------------------- |
| `↑` or `k`      | Navigate up through commands              |
| `↓` or `j`      | Navigate down through commands            |
| `Enter`         | Execute selected command                  |
| `/`             | Enter search mode                         |
| `Ctrl+U`        | Clear search filter                       |
| `Esc`           | Close dialog or exit search mode          |

### Permission Dialog Shortcuts

| Shortcut                | Action                       |
| ----------------------- | ---------------------------- |
| `←` or `left`           | Switch options left          |
| `→` or `right` or `tab` | Switch options right         |
| `Enter` or `space`      | Confirm selection            |
| `a`                     | Allow permission             |
| `A`                     | Allow permission for session |
| `d`                     | Deny permission              |

### Logs Page Shortcuts

| Shortcut           | Action              |
| ------------------ | ------------------- |
| `Backspace` or `q` | Return to chat page |

## AI Assistant Tools

OpenCode's AI assistant has access to various tools to help with coding tasks:

### File and Code Tools

| Tool          | Description                 | Parameters                                                                               |
| ------------- | --------------------------- | ---------------------------------------------------------------------------------------- |
| `glob`        | Find files by pattern       | `pattern` (required), `path` (optional)                                                  |
| `grep`        | Search file contents        | `pattern` (required), `path` (optional), `include` (optional), `literal_text` (optional) |
| `ls`          | List directory contents     | `path` (optional), `ignore` (optional array of patterns)                                 |
| `view`        | View file contents          | `file_path` (required), `offset` (optional), `limit` (optional)                          |
| `write`       | Write to files              | `file_path` (required), `content` (required)                                             |
| `edit`        | Edit files                  | Various parameters for file editing                                                      |
| `patch`       | Apply patches to files      | `file_path` (required), `diff` (required)                                                |
| `diagnostics` | Get diagnostics information | `file_path` (optional)                                                                   |

### Other Tools

| Tool          | Description                            | Parameters                                                                                |
| ------------- | -------------------------------------- | ----------------------------------------------------------------------------------------- |
| `bash`        | Execute shell commands                 | `command` (required), `timeout` (optional)                                                |
| `fetch`       | Fetch data from URLs                   | `url` (required), `format` (required), `timeout` (optional)                               |
| `sourcegraph` | Search code across public repositories | `query` (required), `count` (optional), `context_window` (optional), `timeout` (optional) |
| `agent`       | Run sub-tasks with the AI agent        | `prompt` (required)                                                                       |

## Architecture

OpenCode is built with a modular architecture:

- **cmd**: Command-line interface using Cobra
- **internal/app**: Core application services
- **internal/config**: Configuration management
- **internal/db**: Database operations and migrations
- **internal/llm**: LLM providers and tools integration
- **internal/tui**: Terminal UI components and layouts
- **internal/logging**: Logging infrastructure
- **internal/message**: Message handling
- **internal/session**: Session management
- **internal/lsp**: Language Server Protocol integration

## Enhanced Command System

OpenCode features a powerful slash command interface that provides quick access to both built-in commands and custom user-defined commands.

### Built-in Commands

OpenCode includes essential built-in commands that are available immediately:

| Command | Aliases | Description |
|---------|---------|-------------|
| `/help` | `/h` | Show available commands and keyboard shortcuts |
| `/exit` | `/quit`, `/q` | Exit OpenCode application |
| `/clear` | `/cls`, `/new` | Clear current session and start a new one |
| `/list` | `/ls`, `/commands` | List all available commands |
| `/init` | | Initialize project with OpenCode.md memory file |
| `/compact` | `/summary` | Summarize current session and create new one |

### Command Palette Features

Access the enhanced command palette with `Ctrl+K`:

#### **Visual Organization**
- **⚡ Built-in Commands**: Core OpenCode functionality
- **👤 User Commands**: Personal commands available across all projects  
- **📁 Project Commands**: Project-specific commands

#### **Smart Search**
- Press `/` to enter search mode
- Real-time filtering as you type
- Search across command names, descriptions, aliases, and categories
- Press `Esc` to clear search or exit search mode

#### **Recently Used**
- Commands automatically track usage frequency
- Quick access to your most frequently used commands
- Smart sorting based on recent activity

#### **Enhanced Navigation**
- `↑↓` or `k/j` to navigate commands
- Visual grouping with section headers and icons
- Command aliases displayed for easy discovery
- Contextual help text at the bottom

### Command Execution

Commands are executed differently based on their type:

1. **Built-in Commands**: Execute immediately with their defined behavior
2. **Custom Commands**: Send their content as prompts to the AI assistant
3. **Parameterized Commands**: Show arguments dialog for commands with `$VARIABLE` placeholders

## Custom Commands

OpenCode supports custom commands that can be created by users to quickly send predefined prompts to the AI assistant.

### Creating Custom Commands

Custom commands are predefined prompts stored as Markdown files in one of these locations:

1.  **👤 User Commands**: Available across all projects
    -   `$XDG_CONFIG_HOME/opencode/commands/` (e.g., `~/.config/opencode/commands/`)
    -   `$HOME/.opencode/commands/`

2.  **📁 Project Commands**: Specific to the current project
    -   `<PROJECT DIR>/.opencode/commands/`

Each `.md` file in these directories becomes a custom command. The file name (without the `.md` extension) becomes the command's ID. Commands are automatically categorized and displayed with appropriate icons in the command palette.

### Command Format

Commands can be simple markdown files or include YAML frontmatter for more advanced configuration.

**Basic Command:**

```markdown
RUN git ls-files
READ README.md
```

**Command with YAML Frontmatter:**

```markdown
---
name: "Test Go Project"
description: "Run all Go tests in the current project."
aliases: ["go-test", "test-go"]
---
go test ./...
```

When you run a custom command, the content of the markdown file (without the frontmatter) is sent to the AI.

## Development

### Prerequisites

- Go 1.24.0 or higher

### Building from Source

```bash
# Clone the repository
git clone https://github.com/opencode-ai/opencode.git
cd opencode

# Build
go build -o opencode

# Run
./opencode
```

### Testing

OpenCode includes comprehensive test coverage for all core functionality:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test packages
go test ./internal/config
go test ./internal/commands
go test ./internal/llm/oauth

# Run tests by pattern
go test -run TestOAuth2 ./...
```

#### Test Coverage Areas

- **OAuth2 Authentication**: Complete test suite for Google OAuth2 integration
  - Token management (save, load, refresh, clear)
  - XDG Base Directory Specification compliance
  - File permissions and security
  - Multi-location token handling
  - Environment variable vs config file priority
  - TUI dialog integration and user experience
  - Error handling and edge cases
  - Concurrent access protection

- **Command System**: Built-in and custom command functionality
- **Configuration Management**: Multi-source config loading and validation
- **TUI Components**: Interactive dialog and UI element testing
- **Provider Integration**: LLM provider authentication and communication

The test suite ensures reliability, security, and proper user experience across all authentication methods and configurations.

### Adding Built-in Commands

To add a new built-in command:

1.  Create a new Go file in the `commands` package (e.g., `commands/my_command.go`).
2.  In this file, define a handler function for your command. The handler should have the signature `func(ctx context.Context, args map[string]interface{}) error`.
3.  In an `init()` function in the same file, register your command using the `RegisterBuiltIn` function.

**Example:**

```go
// commands/my_command.go
package commands

import (
	"context"
	"fmt"
)

func handleMyCommand(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("Hello from my command!")
	return nil
}

func init() {
	RegisterBuiltIn(
		NewCommand("my-command", "My Command", "This is my custom command.").
			WithType(BuiltinCommand).
			WithHandler(handleMyCommand).
			Build(),
	)
}
```

## MCP (Model Context Protocol)

OpenCode implements the Model Context Protocol (MCP) to extend its capabilities through external tools. MCP provides a standardized way for the AI assistant to interact with external services and tools.

### MCP Features

- **External Tool Integration**: Connect to external tools and services via a standardized protocol
- **Tool Discovery**: Automatically discover available tools from MCP servers
- **Multiple Connection Types**:
  - **Stdio**: Communicate with tools via standard input/output
  - **SSE**: Communicate with tools via Server-Sent Events
- **Security**: Permission system for controlling access to MCP tools

### Configuring MCP Servers

MCP servers are defined in the configuration file under the `mcpServers` section:

```json
{
  "mcpServers": {
    "example": {
      "type": "stdio",
      "command": "path/to/mcp-server",
      "env": [],
      "args": []
    },
    "web-example": {
      "type": "sse",
      "url": "https://example.com/mcp",
      "headers": {
        "Authorization": "Bearer token"
      }
    }
  }
}
```

### MCP Tool Usage

Once configured, MCP tools are automatically available to the AI assistant alongside built-in tools. They follow the same permission model as other tools, requiring user approval before execution.

## LSP (Language Server Protocol)

OpenCode integrates with Language Server Protocol to provide code intelligence features across multiple programming languages.

### LSP Features

- **Multi-language Support**: Connect to language servers for different programming languages
- **Diagnostics**: Receive error checking and linting information
- **File Watching**: Automatically notify language servers of file changes

### Configuring LSP

Language servers are configured in the configuration file under the `lsp` section:

```json
{
  "lsp": {
    "go": {
      "disabled": false,
      "command": "gopls"
    },
    "typescript": {
      "disabled": false,
      "command": "typescript-language-server",
      "args": ["--stdio"]
    }
  }
}
```

### LSP Integration with AI

The AI assistant can access LSP features through the `diagnostics` tool, allowing it to:

- Check for errors in your code
- Suggest fixes based on diagnostics

While the LSP client implementation supports the full LSP protocol (including completions, hover, definition, etc.), currently only diagnostics are exposed to the AI assistant.

## Using Github Copilot

_Copilot support is currently experimental._

### Requirements
- [Copilot chat in the IDE](https://github.com/settings/copilot) enabled in GitHub settings
- One of:
  - VSCode Github Copilot chat extension
  - Github `gh` CLI
  - Neovim Github Copilot plugin (`copilot.vim` or `copilot.lua`)
  - Github token with copilot permissions

If using one of the above plugins or cli tools, make sure you use the authenticate
the tool with your github account. This should create a github token at one of the following locations:
- ~/.config/github-copilot/[hosts,apps].json
- $XDG_CONFIG_HOME/github-copilot/[hosts,apps].json

If using an explicit github token, you may either set the $GITHUB_TOKEN environment variable or add it to the opencode.json config file at `providers.copilot.apiKey`.

## Gemini OAuth Authentication

OpenCode supports Google OAuth2 authentication for Gemini models, providing an alternative to API keys. OAuth2 offers several advantages:

- **No API Key Management**: Use your Google account instead of managing API keys
- **Enhanced Security**: OAuth2 tokens can be revoked and have expiration times
- **Automatic Refresh**: Tokens are automatically refreshed when they expire
- **XDG Compliant**: Follows modern Linux desktop standards for configuration files

### 🚀 Quick Start

1. **Set up OAuth2 credentials** (see [Getting OAuth2 Credentials](#getting-oauth2-credentials) below)
2. **Configure OpenCode** with your credentials
3. **Run the login command**: `/login gemini`
4. **Authenticate in your browser** when it opens
5. **Start using Gemini models!**

### Configuration

Before using OAuth2 authentication, you need to configure your OAuth2 credentials. OpenCode provides two methods:

#### Method 1: Environment Variables (Recommended for Development)
```bash
export GEMINI_OAUTH_CLIENT_ID="your-client-id.apps.googleusercontent.com"
export GEMINI_OAUTH_CLIENT_SECRET="your-client-secret"
```

#### Method 2: Configuration File (Recommended for Production)
Add to your `.opencode.json` file:
```json
{
  "providers": {
    "gemini": {
      "authMethod": "oauth2",
      "oauth2": {
        "clientId": "your-client-id.apps.googleusercontent.com",
        "clientSecret": "your-client-secret"
      }
    }
  }
}
```

### Getting OAuth2 Credentials

Follow these steps to obtain your OAuth2 credentials from Google:

1. **Go to [Google Cloud Console](https://console.cloud.google.com)**
2. **Create or Select Project**:
   - Click "Select a project" at the top
   - Either create a new project or select an existing one
3. **Enable the API**:
   - Go to "APIs & Services" → "Library"
   - Search for "Generative Language API"
   - Click on it and press "Enable"
4. **Create OAuth2 Credentials**:
   - Go to "APIs & Services" → "Credentials"
   - Click "Create Credentials" → "OAuth 2.0 Client IDs"
   - If prompted, configure the OAuth consent screen first
   - Choose "Desktop application" as the application type
   - Give it a name (e.g., "OpenCode")
   - Click "Create"
5. **Copy Your Credentials**:
   - Copy the "Client ID" and "Client Secret"
   - Use these in your OpenCode configuration

### Authentication Commands

OpenCode provides built-in commands for managing OAuth2 authentication:

| Command | Description |
|---------|-------------|
| `/login gemini` | Start OAuth2 authentication flow |
| `/logout gemini` | Clear stored OAuth2 tokens |
| `/auth status` | Show current authentication status |
| `/auth method` | Display authentication method configuration |

### Authentication Flow

When you run `/login gemini` in the Terminal UI:

1. **TUI Dialog**: OpenCode displays an OAuth2 authentication dialog with a spinner
2. **Credential Validation**: Checks if OAuth2 credentials are configured
3. **Local Server**: Starts a temporary HTTP server on a random port
4. **Browser Launch**: Your default browser opens to Google's OAuth2 page automatically
5. **Visual Feedback**: The dialog shows authentication progress and browser instructions
6. **User Authentication**: You log in with your Google account and grant permissions
7. **Token Storage**: OAuth2 tokens are securely saved to XDG-compliant locations
8. **Success Confirmation**: Dialog shows success message when authentication completes
9. **Ready to Use**: You can now use Gemini models with OAuth2 authentication

The TUI provides a seamless visual experience with real-time feedback during the authentication process.

### Authentication Method Selection

You can configure which authentication method OpenCode should prefer:

| Method | Behavior | Configuration |
|--------|----------|---------------|
| `api_key` | Use only API key authentication | `"authMethod": "api_key"` |
| `oauth2` | Use only OAuth2 authentication | `"authMethod": "oauth2"` |
| `auto` | Try OAuth2 first, fallback to API key | `"authMethod": "auto"` (default) |

### Authentication Priority Order

OpenCode will automatically detect Gemini credentials in this order:

1. **API Key**: `GEMINI_API_KEY` environment variable
2. **OAuth Token**: `GEMINI_TOKEN` environment variable  
3. **OAuth Files**: Automatically loaded from XDG-compliant locations
4. **Provider Config**: API key from configuration file

> **Note**: The priority order can be overridden by setting a specific `authMethod` in your configuration.

### OAuth Token File Locations

OpenCode looks for OAuth token files in the following locations (in order):

- `$XDG_CONFIG_HOME/gemini/oauth_creds.json` (if `XDG_CONFIG_HOME` is set)
- `~/.config/gemini/oauth_creds.json` (standard XDG location)
- `~/.gemini/oauth_creds.json` (fallback location)

### OAuth Token File Format

The OAuth credentials file should be a JSON file with this structure:

```json
{
  "access_token": "your-oauth-access-token",
  "refresh_token": "your-refresh-token",
  "token_type": "Bearer",
  "expiry": "2024-12-31T23:59:59Z"
}
```

### Benefits

- **No API Key Management**: Use your Google account instead of managing API keys
- **XDG Compliant**: Follows modern Linux desktop standards for configuration files
- **Automatic Detection**: Works seamlessly once tokens are in place
- **Backwards Compatible**: Existing API key setups continue to work

For detailed information about generating OAuth tokens, see [GEMINI_OAUTH.md](GEMINI_OAUTH.md).

### Troubleshooting OAuth2

If you encounter issues with OAuth2 authentication:

1. **Run `/auth status`** to check your current authentication state
2. **Verify credentials** - ensure your Client ID and Client Secret are correct
3. **Check API enablement** - make sure the Generative Language API is enabled in Google Cloud Console
4. **Reset authentication** - try `/logout gemini` followed by `/login gemini`
5. **Dialog issues** - if the OAuth2 dialog shows errors, use the retry option or check the console for detailed error messages
6. **Browser problems** - if the browser doesn't open automatically, check the dialog for the authentication URL to copy manually

**TUI Features:**
- **Visual feedback**: The OAuth2 dialog shows real-time progress and status
- **Error handling**: Clear error messages with retry functionality in the dialog
- **Instructions**: The dialog provides helpful guidance during authentication

**Need help?** Run `/login gemini` without configuration - OpenCode will display detailed setup instructions in the OAuth2 dialog.

## Using a self-hosted model provider

OpenCode can also load and use models from a self-hosted (OpenAI-like) provider.
This is useful for developers who want to experiment with custom models.

### Configuring a self-hosted provider

You can use a self-hosted model by setting the `LOCAL_ENDPOINT` environment variable.
This will cause OpenCode to load and use the models from the specified endpoint.

```bash
LOCAL_ENDPOINT=http://localhost:1235/v1
```

### Configuring a self-hosted model

You can also configure a self-hosted model in the configuration file under the `agents` section:

```json
{
  "agents": {
    "coder": {
      "model": "local.granite-3.3-2b-instruct@q8_0",
      "reasoningEffort": "high"
    }
  }
}
```

## Development

### Prerequisites

- Go 1.24.0 or higher

### Building from Source

```bash
# Clone the repository
git clone https://github.com/opencode-ai/opencode.git
cd opencode

# Build
go build -o opencode

# Run
./opencode
```

## Acknowledgments

OpenCode gratefully acknowledges the contributions and support from these key individuals:

- [@isaacphi](https://github.com/isaacphi) - For the [mcp-language-server](https://github.com/isaacphi/mcp-language-server) project which provided the foundation for our LSP client implementation
- [@adamdottv](https://github.com/adamdottv) - For the design direction and UI/UX architecture

Special thanks to the broader open source community whose tools and libraries have made this project possible.

## License

OpenCode is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Here's how you can contribute:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure to update tests as appropriate and follow the existing code style.
