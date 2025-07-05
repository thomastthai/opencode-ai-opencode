package command

import (
	"testing"
	"time"
)

func TestCommandScopes(t *testing.T) {
	testCases := []struct {
		scope        CommandScope
		expectedIcon string
		expectedName string
	}{
		{BuiltinScope, "⚡", "Built-in"},
		{UserScope, "👤", "User"},
		{ProjectScope, "📁", "Project"},
		{CommandScope("unknown"), "•", "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.scope), func(t *testing.T) {
			cmd := Command{Scope: tc.scope}
			
			icon := cmd.GetIcon()
			if icon != tc.expectedIcon {
				t.Errorf("Expected icon %s for scope %s, got %s", tc.expectedIcon, tc.scope, icon)
			}
			
			name := cmd.GetScopeDisplayName()
			if name != tc.expectedName {
				t.Errorf("Expected display name %s for scope %s, got %s", tc.expectedName, tc.scope, name)
			}
		})
	}
}

func TestGetIcon(t *testing.T) {
	tests := []struct {
		name     string
		command  Command
		expected string
	}{
		{
			name:     "builtin command icon",
			command:  Command{Scope: BuiltinScope},
			expected: "⚡",
		},
		{
			name:     "user command icon",
			command:  Command{Scope: UserScope},
			expected: "👤",
		},
		{
			name:     "project command icon",
			command:  Command{Scope: ProjectScope},
			expected: "📁",
		},
		{
			name:     "unknown scope icon",
			command:  Command{Scope: CommandScope("invalid")},
			expected: "•",
		},
		{
			name:     "empty scope icon",
			command:  Command{Scope: CommandScope("")},
			expected: "•",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.command.GetIcon()
			if result != tt.expected {
				t.Errorf("GetIcon() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetScopeDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		command  Command
		expected string
	}{
		{
			name:     "builtin scope display name",
			command:  Command{Scope: BuiltinScope},
			expected: "Built-in",
		},
		{
			name:     "user scope display name",
			command:  Command{Scope: UserScope},
			expected: "User",
		},
		{
			name:     "project scope display name",
			command:  Command{Scope: ProjectScope},
			expected: "Project",
		},
		{
			name:     "unknown scope display name",
			command:  Command{Scope: CommandScope("invalid")},
			expected: "Unknown",
		},
		{
			name:     "empty scope display name",
			command:  Command{Scope: CommandScope("")},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.command.GetScopeDisplayName()
			if result != tt.expected {
				t.Errorf("GetScopeDisplayName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasPlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		command  Command
		expected bool
	}{
		{
			name:     "command with dollar sign placeholder",
			command:  Command{Content: "echo Hello $NAME"},
			expected: true,
		},
		{
			name:     "command with multiple placeholders",
			command:  Command{Content: "deploy $APP to $ENV with $VERSION"},
			expected: true,
		},
		{
			name:     "command without placeholders",
			command:  Command{Content: "echo Hello World"},
			expected: false,
		},
		{
			name:     "command with empty content",
			command:  Command{Content: ""},
			expected: false,
		},
		{
			name:     "command with dollar sign in text",
			command:  Command{Content: "The price is 100$ for this item"},
			expected: true,
		},
		{
			name:     "command with escaped dollar",
			command:  Command{Content: "echo The price is \\$100"},
			expected: true, // Still contains $, even if escaped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.command.HasPlaceholders()
			if result != tt.expected {
				t.Errorf("HasPlaceholders() = %v, want %v for content: %s", result, tt.expected, tt.command.Content)
			}
		})
	}
}

func TestMatchesSearch(t *testing.T) {
	cmd := Command{
		ID:          "git-commit",
		Title:       "Git Commit",
		Description: "Commit changes to repository",
		Category:    "version-control",
		Aliases:     []string{"gc", "commit"},
	}

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		// Search by title
		{
			name:     "matches title exact",
			query:    "Git Commit",
			expected: true,
		},
		{
			name:     "matches title partial",
			query:    "git",
			expected: true,
		},
		{
			name:     "matches title case insensitive",
			query:    "GIT",
			expected: true,
		},
		
		// Search by description
		{
			name:     "matches description",
			query:    "repository",
			expected: true,
		},
		{
			name:     "matches description partial",
			query:    "commit",
			expected: true,
		},
		
		// Search by ID
		{
			name:     "matches ID",
			query:    "git-commit",
			expected: true,
		},
		{
			name:     "matches ID partial",
			query:    "git-",
			expected: true,
		},
		
		// Search by aliases
		{
			name:     "matches alias gc",
			query:    "gc",
			expected: true,
		},
		{
			name:     "matches alias commit",
			query:    "commit",
			expected: true,
		},
		
		// Search by category
		{
			name:     "matches category",
			query:    "version-control",
			expected: true,
		},
		{
			name:     "matches category partial",
			query:    "version",
			expected: true,
		},
		
		// No matches
		{
			name:     "no match",
			query:    "deploy",
			expected: false,
		},
		{
			name:     "empty query matches everything",
			query:    "",
			expected: true,
		},
		
		// Edge cases
		{
			name:     "query with spaces",
			query:    "git commit",
			expected: true, // Matches because title contains "Git Commit" (case insensitive)
		},
		{
			name:     "special characters",
			query:    "git-",
			expected: true, // Matches ID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.MatchesSearch(tt.query)
			if result != tt.expected {
				t.Errorf("MatchesSearch(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

func TestMatchesSearchCaseInsensitive(t *testing.T) {
	cmd := Command{
		Title:       "Deploy Application",
		Description: "Deploy the application to production",
		ID:          "deploy-app",
		Category:    "deployment",
		Aliases:     []string{"deploy"},
	}

	caseTests := []struct {
		query    string
		expected bool
	}{
		{"DEPLOY", true},
		{"deploy", true},
		{"Deploy", true},
		{"dEpLoY", true},
		{"APPLICATION", true},
		{"application", true},
		{"Application", true},
		{"PRODUCTION", true},
		{"production", true},
		{"DEPLOYMENT", true},
		{"deployment", true},
	}

	for _, tt := range caseTests {
		t.Run("case_insensitive_"+tt.query, func(t *testing.T) {
			result := cmd.MatchesSearch(tt.query)
			if result != tt.expected {
				t.Errorf("MatchesSearch(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

func TestCommandWithEmptyFields(t *testing.T) {
	cmd := Command{
		ID: "empty-fields-test",
		// Other fields are empty
	}

	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "matches ID when other fields empty",
			query:    "empty",
			expected: true,
		},
		{
			name:     "no match for empty fields",
			query:    "nonexistent",
			expected: false,
		},
		{
			name:     "empty query matches",
			query:    "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.MatchesSearch(tt.query)
			if result != tt.expected {
				t.Errorf("MatchesSearch(%q) = %v, want %v", tt.query, result, tt.expected)
			}
		})
	}
}

func TestCommandLastUsedTracking(t *testing.T) {
	now := time.Now()
	
	cmd := Command{
		ID:       "test-command",
		LastUsed: now,
	}

	if cmd.LastUsed != now {
		t.Errorf("Expected LastUsed to be %v, got %v", now, cmd.LastUsed)
	}

	// Test zero time
	cmdUnused := Command{
		ID:       "unused-command",
		LastUsed: time.Time{},
	}

	if !cmdUnused.LastUsed.IsZero() {
		t.Error("Expected LastUsed to be zero time for unused command")
	}
}

func TestCommandStructureCompleteness(t *testing.T) {
	// Test that a complete command has all expected fields
	cmd := Command{
		ID:          "complete-command",
		Title:       "Complete Command",
		Description: "A complete command for testing",
		Content:     "echo 'complete'",
		Handler:     nil, // Handler can be nil for custom commands
		Scope:       UserScope,
		Source:      "/path/to/command.md",
		Category:    "testing",
		Aliases:     []string{"complete", "comp"},
		LastUsed:    time.Now(),
	}

	// Verify all fields are accessible
	if cmd.ID != "complete-command" {
		t.Error("ID field not properly set")
	}
	if cmd.Title != "Complete Command" {
		t.Error("Title field not properly set")
	}
	if cmd.Description != "A complete command for testing" {
		t.Error("Description field not properly set")
	}
	if cmd.Content != "echo 'complete'" {
		t.Error("Content field not properly set")
	}
	if cmd.Scope != UserScope {
		t.Error("Scope field not properly set")
	}
	if cmd.Source != "/path/to/command.md" {
		t.Error("Source field not properly set")
	}
	if cmd.Category != "testing" {
		t.Error("Category field not properly set")
	}
	if len(cmd.Aliases) != 2 {
		t.Error("Aliases field not properly set")
	}
	if cmd.LastUsed.IsZero() {
		t.Error("LastUsed field not properly set")
	}
}

func TestCommandScopesConstants(t *testing.T) {
	// Test that scope constants are defined correctly
	if BuiltinScope != "builtin" {
		t.Errorf("BuiltinScope should be 'builtin', got %s", BuiltinScope)
	}
	if UserScope != "user" {
		t.Errorf("UserScope should be 'user', got %s", UserScope)
	}
	if ProjectScope != "project" {
		t.Errorf("ProjectScope should be 'project', got %s", ProjectScope)
	}
}

func TestCommandAliasesHandling(t *testing.T) {
	cmd := Command{
		ID:      "test-aliases",
		Aliases: []string{"alias1", "alias2", "alias3"},
	}

	// Test searching by each alias
	for i, alias := range cmd.Aliases {
		t.Run("alias_"+alias, func(t *testing.T) {
			if !cmd.MatchesSearch(alias) {
				t.Errorf("Should match alias %s at index %d", alias, i)
			}
		})
	}

	// Test that aliases are preserved
	if len(cmd.Aliases) != 3 {
		t.Errorf("Expected 3 aliases, got %d", len(cmd.Aliases))
	}
}