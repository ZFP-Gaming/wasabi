# Authentication Capability

## ADDED Requirements

### Requirement: Discord OAuth Authorization
The system SHALL implement Discord OAuth 2.0 authorization code flow to authenticate users.

#### Scenario: User initiates login
- **WHEN** user clicks "Login with Discord" button
- **THEN** user is redirected to Discord authorization page with correct client_id, redirect_uri, scope, and state parameters

#### Scenario: User grants authorization
- **WHEN** user approves the authorization request on Discord
- **THEN** Discord redirects to callback URL with authorization code and state

#### Scenario: User denies authorization
- **WHEN** user denies the authorization request on Discord
- **THEN** Discord redirects to callback URL with error parameter
- **AND** frontend displays appropriate error message

### Requirement: OAuth Callback Handling
The system SHALL exchange authorization codes for access tokens and create user sessions.

#### Scenario: Valid authorization code received
- **WHEN** callback handler receives valid authorization code
- **THEN** system exchanges code for Discord access token via POST to Discord token endpoint
- **AND** system fetches user profile from Discord API using access token
- **AND** system creates JWT token containing user ID, username, discriminator, and avatar URL
- **AND** system returns JWT token to client as httpOnly cookie

#### Scenario: Invalid authorization code
- **WHEN** callback handler receives invalid or expired authorization code
- **THEN** system responds with 401 Unauthorized error
- **AND** error message indicates authentication failure

#### Scenario: State parameter mismatch
- **WHEN** callback state parameter does not match stored state
- **THEN** system rejects the request with 400 Bad Request error
- **AND** error indicates potential CSRF attack

### Requirement: Session Token Management
The system SHALL generate, validate, and manage JWT tokens for authenticated sessions.

#### Scenario: Token generation
- **WHEN** user successfully authenticates via Discord OAuth
- **THEN** system generates JWT token signed with server secret
- **AND** token contains claims: user_id, username, discriminator, avatar, issued_at, expires_at
- **AND** token expires after 24 hours

#### Scenario: Token validation
- **WHEN** authenticated request includes valid JWT token
- **THEN** system validates token signature and expiration
- **AND** system extracts user information from token claims
- **AND** request proceeds to protected endpoint

#### Scenario: Expired token
- **WHEN** authenticated request includes expired JWT token
- **THEN** system responds with 401 Unauthorized
- **AND** response includes WWW-Authenticate header indicating token expired

#### Scenario: Invalid token signature
- **WHEN** authenticated request includes JWT token with invalid signature
- **THEN** system responds with 401 Unauthorized
- **AND** system logs potential security incident

### Requirement: Protected Endpoints
The system SHALL require authentication for file management operations.

#### Scenario: Authenticated file upload
- **WHEN** authenticated user uploads file via POST /upload
- **THEN** system validates JWT token before processing upload
- **AND** file upload proceeds as normal if token is valid

#### Scenario: Unauthenticated file upload attempt
- **WHEN** unauthenticated user attempts POST /upload without valid token
- **THEN** system responds with 401 Unauthorized
- **AND** response body indicates authentication required

#### Scenario: Authenticated file listing
- **WHEN** authenticated user requests GET /files
- **THEN** system validates JWT token before returning file list
- **AND** all files are returned (no per-user filtering in MVP)

#### Scenario: Authenticated file operations
- **WHEN** authenticated user performs GET/PUT/DELETE on /files/{name}
- **THEN** system validates JWT token before processing operation
- **AND** operation proceeds if token is valid

### Requirement: Logout Functionality
The system SHALL provide logout capability to terminate user sessions.

#### Scenario: User logout
- **WHEN** authenticated user sends POST /auth/logout request
- **THEN** system clears JWT cookie by setting Max-Age=0
- **AND** system responds with 200 OK
- **AND** subsequent requests without new token are unauthorized

#### Scenario: Logout without session
- **WHEN** unauthenticated user sends POST /auth/logout
- **THEN** system responds with 200 OK (idempotent)

### Requirement: User Profile Display
The system SHALL display authenticated user information in the frontend.

#### Scenario: Show logged-in user
- **WHEN** user successfully authenticates
- **THEN** frontend displays user's Discord username
- **AND** frontend displays user's Discord avatar image
- **AND** logout button is visible

#### Scenario: Persist login across page refreshes
- **WHEN** authenticated user refreshes the page
- **THEN** system validates JWT token from cookie
- **AND** user remains logged in without re-authentication
- **AND** user profile continues to display

#### Scenario: Handle authentication expiry
- **WHEN** user's JWT token expires during session
- **THEN** next API request returns 401 Unauthorized
- **AND** frontend clears local auth state
- **AND** frontend redirects user to login page

### Requirement: Configuration Management
The system SHALL support environment-based OAuth configuration.

#### Scenario: Required configuration present
- **WHEN** server starts with DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET, and JWT_SECRET in environment
- **THEN** OAuth handlers initialize successfully
- **AND** server logs confirmation of auth configuration

#### Scenario: Missing OAuth configuration
- **WHEN** server starts without required Discord OAuth variables
- **THEN** server logs error indicating missing configuration
- **AND** server exits with non-zero status code
- **AND** error message specifies which variables are missing

#### Scenario: Development vs production redirect URIs
- **WHEN** REDIRECT_URI environment variable is set
- **THEN** system uses configured redirect URI for OAuth flow
- **AND** system validates that redirect URI matches registered OAuth app settings
