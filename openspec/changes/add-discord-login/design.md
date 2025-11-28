# Design: Discord OAuth Login

## Context
Wasabi currently has no authentication mechanism. All file management endpoints are publicly accessible, which poses security risks. The application needs user authentication to:
- Control access to file operations (upload, list, rename, delete)
- Track which users own which files
- Provide personalized user experience

Discord OAuth was chosen because:
- No need to implement password management, email verification, or account recovery
- Wide adoption in gaming/community spaces
- Simple integration with well-documented API
- Provides user profile data (username, avatar) for UI personalization

## Goals / Non-Goals

### Goals
- Implement Discord OAuth 2.0 authorization code flow
- Secure existing file management endpoints with authentication
- Provide seamless login/logout experience
- Store minimal user session data (no persistent database required initially)
- Display user profile information in the UI

### Non-Goals
- Multi-provider authentication (GitHub, Google, etc.) - can be added later
- User registration/password management
- Role-based access control (RBAC) - all authenticated users have same permissions
- File ownership tracking per user - files remain shared among all authenticated users
- Persistent user database - sessions stored in memory (stateless JWT)

## Decisions

### 1. OAuth Flow: Authorization Code Flow
**Decision**: Use OAuth 2.0 authorization code flow with PKCE optional (Discord supports it but doesn't require it for server-side apps).

**Why**:
- Standard OAuth flow for web applications
- More secure than implicit flow (tokens not exposed in URL)
- Discord's recommended approach

**Alternatives considered**:
- Implicit flow: Less secure, deprecated in OAuth 2.1
- Device code flow: Not suitable for web apps

### 2. Session Storage: Stateless JWT
**Decision**: Use JWT tokens for session management, signed with server secret. Store tokens in httpOnly cookies or localStorage.

**Why**:
- No database required for session storage
- Stateless - server doesn't need to track sessions
- Easy to implement with Go's standard library or lightweight JWT library
- Scales horizontally without session store

**Alternatives considered**:
- Server-side sessions with in-memory store: Requires sticky sessions or shared session store
- Redis session store: Adds infrastructure dependency
- Database sessions: Overkill for simple use case

**Trade-off**: JWT tokens cannot be invalidated server-side without maintaining a blacklist. For MVP, token expiration (1-7 days) is acceptable.

### 3. Token Storage: httpOnly Cookies
**Decision**: Store JWT in httpOnly, secure, SameSite=Lax cookie.

**Why**:
- Protection against XSS attacks (JavaScript cannot access)
- Automatic inclusion in requests (no manual header management)
- Secure flag ensures HTTPS-only transmission

**Alternatives considered**:
- localStorage: Vulnerable to XSS attacks
- sessionStorage: Same XSS vulnerability, also lost on tab close

### 4. Frontend State Management: React Context
**Decision**: Use React Context API for authentication state (logged in user, token status).

**Why**:
- Simple, built-in React solution
- No additional dependencies
- Sufficient for single authentication state

**Alternatives considered**:
- Redux/Zustand: Overkill for simple auth state
- react-mirt might have state management - could leverage existing dependency

### 5. Backend Middleware Architecture
**Decision**: Create authentication middleware that validates JWT and injects user info into request context.

**Pattern**:
```go
func authRequired(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token, err := extractAndValidateJWT(r)
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        // Inject user info into request context
        ctx := context.WithValue(r.Context(), "user", token.Claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}
```

**Why**:
- Clean separation of concerns
- Reusable across all protected endpoints
- Idiomatic Go HTTP middleware pattern

### 6. External Dependencies
**Decision**: Add minimal dependencies:
- Backend: `github.com/golang-jwt/jwt/v5` for JWT signing/validation (or use standard crypto library)
- No database library initially

**Why**:
- Project currently has zero external dependencies
- Keep dependency footprint minimal
- JWT library is well-maintained and widely used

**Alternative**: Implement JWT manually using Go's crypto/hmac and encoding/base64 - more code but zero dependencies.

## Risks / Trade-offs

### Risk: Token Invalidation
**Issue**: JWTs cannot be revoked without server-side blacklist.
**Mitigation**:
- Short token expiration (24 hours)
- Implement refresh token flow in future iteration if needed
- For MVP, logout only clears client-side token

### Risk: CORS Configuration
**Issue**: Frontend and backend may run on different ports during development.
**Mitigation**:
- Configure CORS middleware to allow credentials
- Document CORS settings in .env configuration
- Use `SameSite=Lax` cookies to allow cross-origin OAuth callback

### Risk: Discord API Rate Limits
**Issue**: Discord API has rate limits for OAuth token exchange.
**Mitigation**:
- Cache user profile data in JWT (no repeated API calls)
- Implement exponential backoff for token exchange if needed
- Monitor rate limit headers in responses

### Trade-off: No File Ownership
**Decision**: Files remain shared among all authenticated users (no per-user isolation).
**Reason**: Simpler implementation for MVP. File ownership can be added in future iteration with database.

## Migration Plan

### Phase 1: Add Authentication (Non-Breaking)
1. Implement OAuth handlers without protecting existing endpoints
2. Deploy login UI alongside existing functionality
3. Test authentication flow in production

### Phase 2: Protect Endpoints (Breaking)
1. Add authentication middleware to file management endpoints
2. Update frontend to handle 401 responses
3. Communicate breaking change to users
4. Deploy with feature flag to enable/disable auth requirement

### Rollback
If issues occur:
1. Remove middleware from protected endpoints
2. Keep OAuth handlers live for debugging
3. Revert frontend auth checks

## Open Questions

### Q1: Token Expiration Duration
**Options**: 1 hour, 24 hours, 7 days
**Recommendation**: Start with 24 hours. Adjust based on user feedback.

### Q2: Cookie vs Authorization Header
**Decision**: Use httpOnly cookies for automatic inclusion
**Alternative**: If mobile app or API clients needed, support both cookie and `Authorization: Bearer` header

### Q3: Development vs Production OAuth Apps
**Recommendation**: Use separate Discord OAuth applications for development (localhost redirect) and production (domain redirect). Document setup for both environments.

### Q4: Spanish vs English Error Messages
**Observation**: Codebase currently uses Spanish.
**Recommendation**: Keep authentication errors in Spanish for consistency, or internationalize if targeting broader audience.
