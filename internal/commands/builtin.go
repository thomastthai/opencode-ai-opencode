package commands

import (
	"context"
	"fmt"

	"github.com/opencode-ai/opencode/internal/config"
)

func init() {
	// Register essential built-in commands
	RegisterBuiltIn(
		NewCommand("help", "Help", "Show available commands and keyboard shortcuts").
			WithType(BuiltinCommand).
			WithHandler(handleHelp).
			WithAliases([]string{"h"}).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("exit", "Exit", "Exit OpenCode").
			WithType(BuiltinCommand).
			WithHandler(handleExit).
			WithAliases([]string{"quit", "q"}).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("clear", "Clear", "Clear current session and start new").
			WithType(BuiltinCommand).
			WithHandler(handleClear).
			WithAliases([]string{"cls", "new"}).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("list", "List Commands", "List all available commands").
			WithType(BuiltinCommand).
			WithHandler(handleListCommands).
			WithAliases([]string{"ls", "commands"}).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("init", "Initialize", "Create/Update the OpenCode.md memory file").
			WithType(BuiltinCommand).
			WithHandler(handleInit).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("compact", "Compact", "Summarize current session and create new one").
			WithType(BuiltinCommand).
			WithHandler(handleCompact).
			WithAliases([]string{"summary"}).
			Build(),
	)

	// OAuth2 Authentication commands
	RegisterBuiltIn(
		NewCommand("login", "Login", "Authenticate with OAuth2 providers").
			WithType(BuiltinCommand).
			WithSubCommands(
				NewCommand("gemini", "Login Gemini", "Login to Gemini with OAuth2").
					WithType(BuiltinCommand).
					WithHandler(handleLoginGemini).
					Build(),
			).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("logout", "Logout", "Clear OAuth2 authentication").
			WithType(BuiltinCommand).
			WithSubCommands(
				NewCommand("gemini", "Logout Gemini", "Clear Gemini OAuth2 tokens").
					WithType(BuiltinCommand).
					WithHandler(handleLogoutGemini).
					Build(),
			).
			Build(),
	)

	RegisterBuiltIn(
		NewCommand("auth", "Auth Status", "Show authentication status").
			WithType(BuiltinCommand).
			WithSubCommands(
				NewCommand("status", "Auth Status", "Show current authentication status").
					WithType(BuiltinCommand).
					WithHandler(handleAuthStatus).
					Build(),
				NewCommand("method", "Auth Method", "Show/set authentication method").
					WithType(BuiltinCommand).
					WithHandler(handleAuthMethod).
					Build(),
			).
			Build(),
	)

	// Register a command with sub-commands for testing
	gitCmd := NewCommand("git", "Git", "Git commands").
		WithType(BuiltinCommand).
		WithSubCommands(
			NewCommand("commit", "Commit", "Commit changes").
				WithType(BuiltinCommand).
				WithHandler(handleGitCommit).
				Build(),
			NewCommand("push", "Push", "Push changes").
				WithType(BuiltinCommand).
				WithHandler(handleGitPush).
				Build(),
		).
		Build()
	RegisterBuiltInHierarchy(gitCmd)
}

func handleHelp(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("OpenCode Commands:")
	fmt.Println("  /help, /h        - Show this help")
	fmt.Println("  /exit, /quit, /q - Exit OpenCode")
	fmt.Println("  /clear, /cls     - Clear current session")
	fmt.Println("  /list, /ls       - List all commands")
	fmt.Println("  /init            - Initialize project")
	fmt.Println("  /compact         - Compact session")
	fmt.Println()
	fmt.Println("Authentication Commands:")
	fmt.Println("  /login gemini    - Login to Gemini with OAuth2")
	fmt.Println("  /logout gemini   - Logout from Gemini OAuth2")
	fmt.Println("  /auth status     - Show authentication status")
	fmt.Println("  /auth method     - Show authentication method info")
	fmt.Println()
	fmt.Println("Keyboard Shortcuts:")
	fmt.Println("  Ctrl+K  - Command palette")
	fmt.Println("  Ctrl+S  - Switch session")
	fmt.Println("  Ctrl+O  - Model selection")
	fmt.Println("  Ctrl+H  - Toggle help")
	fmt.Println("  Ctrl+C  - Quit")
	fmt.Println()
	fmt.Println("For OAuth2 setup help, run: /login gemini")
	return nil
}

func handleExit(ctx context.Context, args map[string]interface{}) error {
	// This should trigger application exit
	fmt.Println("Exiting OpenCode...")
	return fmt.Errorf("exit_requested")
}

func handleClear(ctx context.Context, args map[string]interface{}) error {
	// This should trigger a new session
	fmt.Println("Starting new session...")
	return fmt.Errorf("clear_session_requested")
}

func handleListCommands(ctx context.Context, args map[string]interface{}) error {
	registry := GetGlobalRegistry()
	commands := registry.List()
	
	fmt.Println("Available Commands:")
	for _, cmd := range commands {
		fmt.Printf("  /%s - %s\n", cmd.ID(), cmd.Description())
		if len(cmd.GetAliases()) > 0 {
			fmt.Printf("    Aliases: %v\n", cmd.GetAliases())
		}
	}
	return nil
}

func handleInit(ctx context.Context, args map[string]interface{}) error {
	// This should send the init prompt to the AI
	return fmt.Errorf("init_project_requested")
}

func handleCompact(ctx context.Context, args map[string]interface{}) error {
	// This should trigger session compacting
	return fmt.Errorf("compact_session_requested")
}

func handleGitCommit(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("git commit")
	return nil
}

func handleGitPush(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("git push")
	return nil
}

// OAuth2 Authentication handlers

func handleLoginGemini(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("Starting Gemini OAuth2 login...")
	
	// Check if OAuth2 credentials are configured
	service := config.GetOAuth2Service()
	if !service.HasValidCredentials() {
		fmt.Println()
		fmt.Println("❌ OAuth2 credentials not configured!")
		fmt.Println()
		fmt.Println("To use Gemini OAuth2 authentication, you need to configure your OAuth2 credentials:")
		fmt.Println()
		fmt.Println("Method 1: Environment Variables")
		fmt.Println("  export GEMINI_OAUTH_CLIENT_ID=\"your-client-id.apps.googleusercontent.com\"")
		fmt.Println("  export GEMINI_OAUTH_CLIENT_SECRET=\"your-client-secret\"")
		fmt.Println()
		fmt.Println("Method 2: Configuration File (.opencode.json)")
		fmt.Println("  {")
		fmt.Println("    \"providers\": {")
		fmt.Println("      \"gemini\": {")
		fmt.Println("        \"oauth2\": {")
		fmt.Println("          \"clientId\": \"your-client-id.apps.googleusercontent.com\",")
		fmt.Println("          \"clientSecret\": \"your-client-secret\"")
		fmt.Println("        }")
		fmt.Println("      }")
		fmt.Println("    }")
		fmt.Println("  }")
		fmt.Println()
		fmt.Println("To get OAuth2 credentials:")
		fmt.Println("1. Go to Google Cloud Console (https://console.cloud.google.com)")
		fmt.Println("2. Create a new project or select an existing one")
		fmt.Println("3. Enable the Generative Language API")
		fmt.Println("4. Go to 'Credentials' → 'Create Credentials' → 'OAuth 2.0 Client IDs'")
		fmt.Println("5. Set application type to 'Desktop application'")
		fmt.Println("6. Copy the Client ID and Client Secret")
		return fmt.Errorf("OAuth2 credentials not configured")
	}
	
	if err := config.LoginWithGeminiOAuth2(ctx); err != nil {
		return fmt.Errorf("OAuth2 login failed: %w", err)
	}
	
	fmt.Println("Successfully authenticated with Gemini OAuth2!")
	fmt.Println("You can now use Gemini models in OpenCode.")
	return nil
}

func handleLogoutGemini(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("Clearing Gemini OAuth2 tokens...")
	
	if err := config.LogoutGeminiOAuth2(); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	
	fmt.Println("Successfully logged out from Gemini OAuth2.")
	fmt.Println("You will need to authenticate again to use Gemini models.")
	return nil
}

func handleAuthStatus(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("Authentication Status:")
	fmt.Println("====================")
	
	// Show Gemini authentication status
	status, isValid := config.GetGeminiAuthStatus()
	fmt.Printf("Gemini: %s", status)
	if isValid {
		fmt.Printf(" ✓")
	} else {
		fmt.Printf(" ✗")
	}
	fmt.Println()
	
	// Show current authentication method preference
	method := config.GetGeminiAuthMethod()
	fmt.Printf("Preferred Method: %s\n", method)
	
	return nil
}

func handleAuthMethod(ctx context.Context, args map[string]interface{}) error {
	fmt.Println("Authentication Method Configuration:")
	fmt.Println("===================================")
	
	current := config.GetGeminiAuthMethod()
	fmt.Printf("Current method: %s\n", current)
	fmt.Println()
	fmt.Println("Available methods:")
	fmt.Println("  api_key - Use API key authentication only")
	fmt.Println("  oauth2  - Use OAuth2 authentication only")
	fmt.Println("  auto    - Try OAuth2 first, fallback to API key")
	fmt.Println()
	fmt.Println("To change method, use configuration file or environment variables.")
	fmt.Println("Example: Set 'authMethod': 'oauth2' in providers.gemini section")
	
	return nil
}