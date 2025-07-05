package command

import (
	"strings"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
)

// CommandScope represents the scope/source of a command
type CommandScope string

const (
	BuiltinScope CommandScope = "builtin"
	UserScope    CommandScope = "user"
	ProjectScope CommandScope = "project"
)

// Command represents a command that can be executed
type Command struct {
	ID          string
	Title       string
	Description string
	Content     string
	Handler     func(cmd Command) tea.Cmd
	Scope       CommandScope
	Source      string    // file path for custom commands
	Category    string    // e.g., "git", "testing", "deployment"
	Aliases     []string  // alternative names
	LastUsed    time.Time // for recently used tracking
}

// GetIcon returns the icon for the command based on its scope
func (c Command) GetIcon() string {
	switch c.Scope {
	case BuiltinScope:
		return "⚡"
	case UserScope:
		return "👤"
	case ProjectScope:
		return "📁"
	default:
		return "•"
	}
}

// GetScopeDisplayName returns a human-readable scope name
func (c Command) GetScopeDisplayName() string {
	switch c.Scope {
	case BuiltinScope:
		return "Built-in"
	case UserScope:
		return "User"
	case ProjectScope:
		return "Project"
	default:
		return "Unknown"
	}
}

// HasPlaceholders checks if the command content has $VARIABLE placeholders
func (c Command) HasPlaceholders() bool {
	return strings.Contains(c.Content, "$")
}

// MatchesSearch checks if the command matches a search query
func (c Command) MatchesSearch(query string) bool {
	query = strings.ToLower(query)
	
	// Check title
	if strings.Contains(strings.ToLower(c.Title), query) {
		return true
	}
	
	// Check description
	if strings.Contains(strings.ToLower(c.Description), query) {
		return true
	}
	
	// Check ID
	if strings.Contains(strings.ToLower(c.ID), query) {
		return true
	}
	
	// Check aliases
	for _, alias := range c.Aliases {
		if strings.Contains(strings.ToLower(alias), query) {
			return true
		}
	}
	
	// Check category
	if strings.Contains(strings.ToLower(c.Category), query) {
		return true
	}
	
	return false
}
