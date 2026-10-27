# Authentication Sessions and Protected Routes

## Goal

Implement secure, revocable server-side sessions for TaleHaven and enforce
authentication/authorization on protected API routes.

## Architecture Decision

Use an opaque, high-entropy session credential in a cookie rather than a JWT
as the primary browser session. The cookie carries no user data; the server
stores the session state and can revoke it immediately for logout, password
changes, account restrictions, and future role changes.

Generate at least 128 bits of cryptographically random secret material. Store
a one-way verifier/hash of the secret rather than the raw credential. The
cookie may contain an identifier plus verifier, but the identifier alone must
never authenticate a request. Do not use browser storage for any
authentication credential.

## Data Changes

Add a session store/table with at least:

- session lookup identifier and hashed verifier;
- user ID;
- created, last-used, idle-expiry, and absolute-expiry timestamps;
- revoked/invalidated timestamp;
- optional user-agent/device metadata for future session management.

Add or reserve a user role field (`user`, `admin`) so authorization can grow
without redesigning sessions.

## API Changes

- `POST /auth/login`: verify credentials, create a new session, and set the
  cookie.
- `POST /auth/logout`: revoke the current server-side session and expire its
  cookie.
- `GET /auth/me`: return safe authenticated user/profile data, never a session
  credential or password-related field.
- Authentication middleware: validate cookie sessions, expiry, revocation, and
  optionally update last-used time.
- Authorization middleware/helpers: require authenticated users and enforce
  roles for administrator endpoints.

## Cookie and Request Security

- Production cookie: `HttpOnly`, `Secure`, `SameSite=Lax` initially,
  `Path=/`, and no broad `Domain` attribute. Evaluate `SameSite=Strict` once
  navigation flows are confirmed. Use the `__Host-` prefix in HTTPS production
  where deployment permits it.
- Use environment-specific local development settings because `Secure` cookies
  cannot be sent over ordinary HTTP localhost.
- Regenerate/create a fresh authenticated session at login; never accept an
  arbitrary client-supplied session identifier.
- Enforce server-side idle and absolute expirations (initial target: 30-minute
  idle timeout and 7–14-day absolute lifetime, subject to product decision).
- Require CSRF protection and/or strict Origin checking for state-changing
  cookie-authenticated routes. `SameSite` is defense in depth, not a complete
  CSRF solution.
- Configure CORS credentials only for the known frontend development origin;
  never use wildcard origin with credentials.

## Implementation Steps

1. Inspect and replace the existing authentication mechanism without breaking
   current account/library ownership data.
2. Add the session migration, cryptographic token generation, verifier
   handling, expiry, revocation, and cleanup policy.
3. Implement login, logout, and current-user endpoints with no sensitive data
   in responses.
4. Add authentication middleware to all user-specific data routes and role
   checks where administration is introduced.
5. Add CSRF/Origin validation and production/development cookie configuration.
6. Update CORS for credentialed frontend requests.
7. Test login, refresh persistence, unauthorized access, expiry, logout,
   revocation, per-user data isolation, and admin-role denial.

## Validation

- The raw session credential is never retained in the database or sent in API
  JSON.
- Cookie attributes match the environment and production requests use HTTPS.
- Protected endpoints reject absent, expired, revoked, malformed, and
  cross-user sessions.
- Logout invalidates the server session, not merely the local cookie.
- Backend database and integration tests cover session lifecycle and protected
  route authorization.

## Implementation Status

- Replaced JWT-cookie authentication with an opaque `talehaven_session`
  cookie backed by a server-side `sessions` record. Each record stores only a
  SHA-256 verifier hash plus idle/absolute expiry and revocation timestamps.
- Login and registration create sessions; logout revokes the server record;
  password changes revoke every session for that user; `/auth/me` is now an
  authenticated route.
- Reader progress now uses required authentication rather than optional
  authentication, matching its existing per-user storage model.
- Added `user.role` with `user`/`admin` values for future authorization.
- Added exact-origin protections for unsafe cookie requests, credentialed CORS
  support through `FRONTEND_ORIGIN`, and environment-driven `Secure` cookie
  behavior (`SESSION_COOKIE_SECURE`; production defaults to secure cookies).
- Tests cover opaque credential parsing, verifier validation, database-backed
  create/authenticate/revoke lifecycle, trusted origins, cookie attributes,
  controllers, and the existing database suite.
- Passed `go test ./modules/sessions ./middlewares ./controllers` and the
  required `go test ./modules/database`. The rebuilt backend completed the
  migration and served `/auth/me` with the expected unauthenticated `401`.
  A disposable live smoke test passed registration, `/auth/me`, logout
  revocation, a second login, and temporary-account deletion.

## Follow-up

JWTs may be introduced later for a non-browser/mobile or third-party API use
case. If used for the browser, design a separate short-lived access-token plus
rotating, server-revocable refresh-session flow; do not add that complexity to
the initial web session implementation.
