package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencode-ai/opencode/internal/app"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/llm/models"
)

// ContextProvider provides dynamic completions based on current context
type ContextProvider struct {
	app *app.App
}

// NewContextProvider creates a new context provider
func NewContextProvider(app *app.App) *ContextProvider {
	return &ContextProvider{app: app}
}

// GetSessionCompletions returns available session IDs/names
func (c *ContextProvider) GetSessionCompletions() []CommandCompletion {
	if c.app == nil || c.app.Sessions == nil {
		return nil
	}

	sessions, err := c.app.Sessions.List(context.Background())
	if err != nil {
		return nil
	}

	completions := make([]CommandCompletion, 0, len(sessions))
	for _, session := range sessions {
		// Show both ID and title for clarity
		display := fmt.Sprintf("%s (%s)", session.Title, session.ID[:8])
		completions = append(completions, CommandCompletion{
			Value:       session.ID,
			Display:     display,
			Description: fmt.Sprintf("%d messages", session.MessageCount),
			Complete:    session.ID,
		})
	}

	return completions
}

// GetModelCompletions returns available AI models
func (c *ContextProvider) GetModelCompletions() []CommandCompletion {
	cfg := config.Get()
	if cfg == nil {
		return nil
	}

	// Get all available models from the supported models map
	completions := make([]CommandCompletion, 0, len(models.SupportedModels))

	for modelID, model := range models.SupportedModels {
		// Check if provider is enabled in config
		providerCfg, exists := cfg.Providers[model.Provider]
		if !exists || providerCfg.Disabled {
			continue
		}

		completions = append(completions, CommandCompletion{
			Value:       string(modelID),
			Display:     model.Name,
			Description: fmt.Sprintf("Provider: %s", model.Provider),
			Complete:    string(modelID),
		})
	}

	return completions
}

// GetFileCompletions returns file paths from current directory
func (c *ContextProvider) GetFileCompletions(prefix string) []CommandCompletion {
	dir := "."
	if prefix != "" && strings.Contains(prefix, "/") {
		dir = filepath.Dir(prefix)
		prefix = filepath.Base(prefix)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	completions := make([]CommandCompletion, 0)
	for _, entry := range entries {
		name := entry.Name()
		
		// Skip hidden files unless prefix starts with .
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".") {
			continue
		}

		// Filter by prefix
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}

		path := filepath.Join(dir, name)
		if dir == "." {
			path = name
		}

		icon := "📄"
		if entry.IsDir() {
			icon = "📁"
			path += "/"
		}

		completions = append(completions, CommandCompletion{
			Value:       path,
			Display:     fmt.Sprintf("%s %s", icon, name),
			Description: "",
			Complete:    path,
		})
	}

	return completions
}

// GetProviderCompletions returns available auth providers
func (c *ContextProvider) GetProviderCompletions() []CommandCompletion {
	return []CommandCompletion{
		{
			Value:       "gemini",
			Display:     "Gemini",
			Description: "Google Gemini AI",
			Complete:    "gemini",
		},
		{
			Value:       "anthropic",
			Display:     "Anthropic",
			Description: "Anthropic Claude",
			Complete:    "anthropic",
		},
		{
			Value:       "openai",
			Display:     "OpenAI",
			Description: "OpenAI GPT models",
			Complete:    "openai",
		},
	}
}

// GetDynamicCompletions returns context-aware completions for specific commands
func GetDynamicCompletions(topic, command string, args []string, app *app.App) []CommandCompletion {
	provider := NewContextProvider(app)

	// Handle different command contexts
	switch topic {
	case "session":
		switch command {
		case "switch", "show", "delete":
			return provider.GetSessionCompletions()
		}
		
	case "config":
		switch command {
		case "model":
			if len(args) == 0 {
				return provider.GetModelCompletions()
			}
		}
		
	case "auth":
		switch command {
		case "login", "logout":
			if len(args) == 0 {
				return provider.GetProviderCompletions()
			}
		}
		
	case "project":
		switch command {
		case "add-dir":
			prefix := ""
			if len(args) > 0 {
				prefix = args[len(args)-1]
			}
			return provider.GetFileCompletions(prefix)
		}
	}

	return nil
}