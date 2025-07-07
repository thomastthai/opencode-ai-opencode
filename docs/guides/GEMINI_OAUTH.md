# Gemini OAuth2 Authentication Guide

OpenCode supports Google OAuth2 authentication for Gemini models, providing a secure alternative to API keys. This guide covers everything you need to know about setting up and using OAuth2 authentication.

## Prerequisites

Before starting, you need to obtain OAuth2 credentials from Google Cloud Console. If you haven't done this yet, see the [Getting OAuth2 Credentials](#getting-oauth2-credentials) section below.

## Configuration

OpenCode requires OAuth2 credentials (Client ID and Client Secret) to be configured before you can use OAuth2 authentication.

### Method 1: Environment Variables

Set these environment variables in your shell:

```bash
export GEMINI_OAUTH_CLIENT_ID="your-client-id.apps.googleusercontent.com"
export GEMINI_OAUTH_CLIENT_SECRET="your-client-secret"
```

Add these to your shell profile (`.bashrc`, `.zshrc`, etc.) to make them permanent.

### Method 2: Configuration File

Add the OAuth2 configuration to your `.opencode.json` file:

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

## Getting OAuth2 Credentials

Follow these detailed steps to obtain OAuth2 credentials from Google:

### Step 1: Access Google Cloud Console

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Sign in with your Google account

### Step 2: Create or Select a Project

1. Click on the project dropdown at the top of the page
2. Either:
   - Click "New Project" to create a new project
   - Select an existing project from the list

### Step 3: Enable the Generative Language API

1. In the sidebar, go to "APIs & Services" → "Library"
2. Search for "Generative Language API"
3. Click on the API result
4. Click the "Enable" button

### Step 4: Configure OAuth Consent Screen (if needed)

If this is your first time setting up OAuth2 for this project:

1. Go to "APIs & Services" → "OAuth consent screen"
2. Choose "External" user type (unless you're using a Google Workspace account)
3. Fill in the required fields:
   - App name: "OpenCode" (or your preferred name)
   - User support email: Your email address
   - Developer contact information: Your email address
4. Click "Save and Continue"
5. Skip the "Scopes" section (click "Save and Continue")
6. Skip the "Test users" section (click "Save and Continue")

### Step 5: Create OAuth2 Credentials

1. Go to "APIs & Services" → "Credentials"
2. Click "Create Credentials" → "OAuth 2.0 Client IDs"
3. Choose "Desktop application" as the application type
4. Enter a name for your OAuth2 client (e.g., "OpenCode Desktop")
5. Click "Create"

### Step 6: Copy Your Credentials

1. A dialog will appear with your Client ID and Client Secret
2. Copy both values and store them securely
3. Use these values in your OpenCode configuration

## How OAuth2 Authentication Works

OpenCode will automatically look for Gemini authentication in the following order:

1. **Environment Variable**: `GEMINI_TOKEN` (direct token override)
2. **XDG Config Home**: `$XDG_CONFIG_HOME/gemini/oauth_creds.json`
3. **Standard Config**: `~/.config/gemini/oauth_creds.json`
4. **Fallback Location**: `~/.gemini/oauth_creds.json`

## Using OAuth2 Authentication

### Initial Setup

1. **Configure Credentials**: Use one of the configuration methods above
2. **Run Login Command**: In OpenCode's Terminal UI, execute `/login gemini`
3. **OAuth2 Dialog**: A visual dialog appears with authentication progress
4. **Browser Launch**: Your browser opens to Google's OAuth2 page automatically
5. **Visual Feedback**: The dialog shows a spinner and instructions during authentication
6. **Grant Permissions**: Allow OpenCode to access your Google account in the browser
7. **Success Confirmation**: The dialog displays a success message when complete
8. **Done**: OAuth2 tokens are automatically saved and ready to use

The Terminal UI provides a seamless visual experience with real-time feedback throughout the process.

### Authentication Commands

| Command | Description |
|---------|-------------|
| `/login gemini` | Start OAuth2 authentication flow |
| `/logout gemini` | Clear stored OAuth2 tokens |
| `/auth status` | Show current authentication status |
| `/auth method` | Display authentication method configuration |

### Authentication Flow Details

When you run `/login gemini` in the Terminal UI:

1. **Dialog Display**: OpenCode shows an OAuth2 authentication dialog with a spinner
2. **Credential Validation**: Checks if OAuth2 credentials are configured in background
3. **Local Server**: Starts a temporary HTTP server on a random unused port
4. **Browser Launch**: Opens your default browser to Google's OAuth2 authorization page
5. **Visual Progress**: Dialog shows authentication status and helpful instructions
6. **User Authentication**: You log in with your Google account and grant permissions
7. **Token Exchange**: OAuth2 authorization code is exchanged for access and refresh tokens
8. **Secure Storage**: Tokens are saved to XDG-compliant locations with proper file permissions
9. **Success Display**: Dialog shows confirmation message when authentication completes
10. **Ready to Use**: You can now use Gemini models with OAuth2 authentication

**TUI Experience:**
- Real-time visual feedback during the entire process
- Clear instructions if browser doesn't open automatically
- Error handling with retry functionality
- Success confirmation with user-friendly messages

## Token File Format

The OAuth credentials file is automatically generated with this structure:

```json
{
  "access_token": "your-oauth-access-token",
  "refresh_token": "your-refresh-token",
  "token_type": "Bearer",
  "expiry": "2024-12-31T23:59:59Z"
}
```

> **Note**: You don't need to create this file manually. It's automatically generated when you complete the OAuth2 flow.

## Authentication Method Selection

You can control which authentication method OpenCode uses:

### Configuration Options

| Method | Behavior | Configuration |
|--------|----------|---------------|
| `api_key` | Use only API key authentication | `"authMethod": "api_key"` |
| `oauth2` | Use only OAuth2 authentication | `"authMethod": "oauth2"` |
| `auto` | Try OAuth2 first, fallback to API key | `"authMethod": "auto"` (default) |

### Example Configuration

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

## Credential Priority Order

OpenCode will use Gemini credentials in this order:

1. **API Key**: `GEMINI_API_KEY` environment variable
2. **Direct Token**: `GEMINI_TOKEN` environment variable  
3. **OAuth Files**: Automatically loaded from XDG-compliant locations
4. **Config File**: API key from configuration file

> **Note**: The priority order can be overridden by setting a specific `authMethod` in your configuration.

## Benefits of OAuth2

- **Enhanced Security**: Tokens have expiration times and can be revoked
- **No API Key Management**: Use your Google account instead of managing API keys
- **Automatic Refresh**: Tokens are automatically refreshed when they expire
- **XDG Compliant**: Follows modern Linux desktop standards for configuration files
- **Backwards Compatible**: Existing API key setups continue to work
- **Audit Trail**: OAuth2 provides better audit trails in Google Cloud Console

## Security Features

- **Secure Storage**: Token files are saved with restrictive file permissions (0600)
- **Temporary Server**: Local OAuth2 server runs only during authentication and uses random ports
- **Token Refresh**: Expired tokens are automatically refreshed without user intervention
- **Scope Limited**: Only requests necessary permissions for Gemini API access
- **Local Only**: No tokens or credentials are sent to external services (except Google)

## Troubleshooting

### Common Issues

1. **"OAuth2 credentials not configured"**
   - Ensure `GEMINI_OAUTH_CLIENT_ID` and `GEMINI_OAUTH_CLIENT_SECRET` are set
   - Or verify your `.opencode.json` configuration is correct

2. **Browser doesn't open**
   - The OAuth2 dialog will show instructions and the authentication URL
   - Copy the URL from the dialog or console and paste it into your browser manually

3. **"Invalid client" error**
   - Verify your Client ID and Client Secret are correct
   - Ensure the Generative Language API is enabled in your Google Cloud project

4. **Permissions error**
   - Make sure you're using the correct Google account
   - Check that your OAuth2 consent screen is properly configured

5. **TUI Dialog Issues**
   - If the dialog shows an error, use the retry button (press 'r') to try again
   - Check the error message in the dialog for specific guidance
   - If dialog becomes unresponsive, press Esc to close and try again

### TUI-Specific Features

**Visual Indicators:**
- **Spinner**: Shows when authentication is in progress
- **Success Icon**: Green checkmark when authentication completes
- **Error Icon**: Red X with error message if authentication fails
- **Instructions**: Helpful text about browser opening and next steps

**Interactive Controls:**
- **Retry (r)**: Restart authentication process after an error
- **Close (Esc)**: Close the dialog (available when not authenticating)
- **Error Details**: Detailed error messages displayed in the dialog

### Getting Help

If you encounter issues:

1. Run `/auth status` to check your current authentication state
2. Check the OpenCode logs for detailed error messages
3. Verify your Google Cloud Console configuration
4. Try `/logout gemini` followed by `/login gemini` to reset authentication

## Testing and Quality Assurance

The OAuth2 implementation includes comprehensive test coverage to ensure reliability and security:

### Test Coverage Areas

- **Token Management**: Save, load, refresh, and clear operations
- **XDG Compliance**: Proper config directory usage following Linux desktop standards
- **Security**: File permissions (0600), credential validation, concurrent access protection
- **Multi-location Support**: Token file handling across different XDG paths
- **Configuration Priority**: Environment variables vs config files vs direct tokens
- **TUI Integration**: Visual dialog behavior, user interaction, error handling
- **Error Scenarios**: Invalid credentials, corrupted files, network failures, browser issues
- **Cross-platform Compatibility**: File system differences and permission handling

### Running OAuth2 Tests

To run the OAuth2-specific tests:

```bash
# Run all OAuth2 tests
go test -run TestOAuth2 ./...

# Run OAuth2 service tests
go test ./internal/llm/oauth

# Run OAuth2 config integration tests
go test ./internal/config -run TestOAuth2

# Run OAuth2 TUI dialog tests
go test ./internal/tui/components/dialog -run TestOAuth2

# Run OAuth2 command tests
go test ./internal/commands -run TestOAuth2
```

### Quality Metrics

The OAuth2 implementation has been thoroughly tested for:

- ✅ **Functionality**: All OAuth2 flows work correctly
- ✅ **Security**: Proper file permissions and credential handling
- ✅ **Reliability**: Error handling and edge case coverage
- ✅ **User Experience**: Clear feedback and intuitive TUI integration
- ✅ **Standards Compliance**: XDG Base Directory Specification adherence
- ✅ **Cross-platform**: Works on Linux, macOS, and Windows

This ensures that OAuth2 authentication is production-ready and provides a secure, reliable experience for users.