# Change: Add Discord OAuth Login

## Why
Users need a way to authenticate with the application. Adding Discord OAuth login provides a seamless authentication experience that leverages existing Discord accounts, eliminating the need for users to create and manage separate credentials. This is particularly useful for applications targeting gaming or community-focused audiences where Discord is already widely used.

## What Changes
- Add Discord OAuth 2.0 authentication flow to the backend
- Implement OAuth callback handler to exchange authorization codes for access tokens
- Store user session data (Discord user ID, username, avatar) in memory or persistent storage
- Add session management with JWT tokens for authenticated API requests
- Create frontend login UI with "Login with Discord" button
- Protect existing file management endpoints with authentication middleware
- Add user profile display in the frontend showing Discord username and avatar

## Impact
- **Affected specs**: `auth` (new capability)
- **Affected code**:
  - Backend: `main.go` (add OAuth handlers, session middleware)
  - Frontend: `src/components/` (new Login component), `src/services/api.js` (auth helpers)
  - Configuration: `.env.example` (add Discord client ID and secret)
- **Breaking changes**: File management endpoints will require authentication (users must log in first)
