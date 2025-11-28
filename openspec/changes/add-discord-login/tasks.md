# Implementation Tasks

## 1. Backend Setup
- [x] 1.1 Register Discord OAuth application and obtain client ID/secret
- [x] 1.2 Add Discord OAuth configuration to `.env.example` (CLIENT_ID, CLIENT_SECRET, REDIRECT_URI)
- [x] 1.3 Add session secret for JWT signing to environment configuration

## 2. Backend Authentication
- [x] 2.1 Implement OAuth authorization redirect handler (`GET /auth/discord`)
- [x] 2.2 Implement OAuth callback handler (`GET /auth/discord/callback`)
- [x] 2.3 Exchange authorization code for Discord access token
- [x] 2.4 Fetch Discord user profile (user ID, username, discriminator, avatar)
- [x] 2.5 Generate and sign JWT token with user session data
- [x] 2.6 Return JWT token to frontend (via cookie or JSON response)

## 3. Session Management
- [x] 3.1 Create authentication middleware to validate JWT tokens
- [x] 3.2 Extract user information from validated tokens
- [x] 3.3 Protect existing endpoints (`/upload`, `/files`, `/files/*`) with auth middleware
- [x] 3.4 Implement logout handler (`POST /auth/logout`) to invalidate sessions

## 4. Frontend Login UI
- [x] 4.1 Create `Login.jsx` component with Discord login button
- [x] 4.2 Add redirect to Discord OAuth authorization URL
- [x] 4.3 Handle OAuth callback route and store JWT token
- [x] 4.4 Create `UserProfile.jsx` component to display logged-in user info
- [x] 4.5 Add logout button functionality

## 5. Frontend Integration
- [x] 5.1 Update `api.js` to include JWT token in all API requests (Authorization header)
- [x] 5.2 Add token storage (localStorage or sessionStorage)
- [x] 5.3 Implement automatic redirect to login page for unauthenticated requests
- [x] 5.4 Update `App.jsx` to show login page vs main app based on auth state

## 6. Testing & Documentation
- [x] 6.1 Test full OAuth flow (authorization, callback, token generation)
- [x] 6.2 Test protected endpoints with and without authentication
- [x] 6.3 Test token expiration and refresh scenarios
- [x] 6.4 Update README.md with Discord OAuth setup instructions
- [x] 6.5 Document required environment variables
