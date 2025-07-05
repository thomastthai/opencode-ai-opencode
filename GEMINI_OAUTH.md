# Gemini Web Authentication

OpenCode now supports Google OAuth2 authentication for Gemini, similar to how GitHub Copilot authentication works. This allows you to use your Google account credentials instead of just API keys.

## How It Works

OpenCode will automatically look for Gemini OAuth2 tokens in the following locations (in order):

1. **Environment Variable**: `GEMINI_TOKEN`
2. **XDG Config Home**: `$XDG_CONFIG_HOME/gemini/oauth_creds.json`
3. **Standard Config**: `~/.config/gemini/oauth_creds.json`
4. **Fallback Location**: `~/.gemini/oauth_creds.json`

## Token File Format

The OAuth credentials file should be a JSON file with the following structure:

```json
{
  "access_token": "your-oauth-access-token",
  "refresh_token": "your-refresh-token",
  "token_type": "Bearer",
  "expiry": "2024-12-31T23:59:59Z"
}
```

## Priority Order

OpenCode will use credentials in this order:

1. `GEMINI_API_KEY` environment variable
2. `GEMINI_TOKEN` environment variable  
3. OAuth token files (in the order listed above)

## Generating OAuth Tokens

You can use the Google OAuth2 flow to generate tokens. Here's a simple example using the OAuth2 client credentials from the Gemini CLI:

- **Client ID**: `681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com`
- **Scopes**: 
  - `https://www.googleapis.com/auth/cloud-platform`
  - `https://www.googleapis.com/auth/userinfo.email`
  - `https://www.googleapis.com/auth/userinfo.profile`

## Benefits

- **No API Key Required**: Use your Google account instead of managing API keys
- **XDG Compliant**: Follows modern Linux desktop standards for configuration files
- **Backwards Compatible**: Existing API key setups continue to work
- **Automatic Detection**: Works seamlessly once tokens are in place

## Security

- Token files are read with standard file permissions
- Only the `access_token` field is used for API authentication
- Follows the same security patterns as GitHub Copilot integration