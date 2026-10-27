# Authentication Sessions and Protected Routes

## Goal

Move TaleHaven to secure, cookie-based authenticated sessions and make the
frontend clearly handle signed-in, signed-out, and expired-session states.

## Scope

- Use an opaque session cookie; JavaScript must never read or store an access
  token, JWT, or refresh token in `localStorage` or `sessionStorage`.
- Configure API calls to include cookies.
- Add a session bootstrap request and client-side protected-route guard.
- Provide login/logout flows and friendly expired-session handling.
- Keep backend authorization as the source of truth; route guards are only a
  user-experience layer.

## API Contract

The frontend will use these authenticated endpoints:

- `POST /auth/login` — submits credentials and receives only safe user data;
  the browser receives the session cookie through `Set-Cookie`.
- `POST /auth/logout` — invalidates the session and clears the cookie.
- `GET /auth/me` — returns the current safe user/profile summary or an
  unauthenticated response.

All cookie-based requests must use Axios/fetch credentials. Local development
CORS must use the exact frontend origin and allow credentials; production
should prefer frontend and API under the same site.

## Implementation Steps

1. **Next implementation: add an authentication client and app-level session
   bootstrap** based on `GET /auth/me`, while the application shows a neutral
   loading state.
2. Store only the returned safe user/profile data in React state; never retain
   the opaque cookie credential, JWT, or refresh token in browser storage.
3. Configure a single global `401` handler that clears local user state and
   routes users to Login when an existing session expires or is revoked.
4. Add a protected-route wrapper for Profile, Settings, library mutations,
   reading-progress writes, favorites, future chatrooms, and administrator UI.
5. Update login to call the session endpoint rather than retain credentials or
   tokens in browser storage.
6. Add logout to the main navigation and reset user-specific cached state
   after successful logout.
7. Add clear loading, unauthenticated, and expired-session screens.

## Backend Prerequisite

Completed on October 26, 2026: the backend now provides cookie-backed,
revocable `POST /auth/login`, `POST /auth/logout`, and protected `GET
/auth/me` routes. The frontend is the remaining half of this feature.

## Validation

- A signed-out visitor cannot access protected UI or successfully mutate a
  protected API resource.
- Login restores the correct profile through `/auth/me` after a page refresh.
- Logout removes access immediately and a stale/expired session is handled
  gracefully.
- Tests confirm API calls include credentials and protected routes behave for
  signed-in and signed-out users.

## Implementation Status

- Retained the existing app-level `/auth/me` bootstrap and strengthened it for
  opaque server sessions: the safe user response is held only in React state.
- Added a global `401` session-expiry event for protected requests. The auth
  provider clears the user state, and protected pages redirect to Login with
  a clear expired-session message.
- Login, registration, and the expected initial `/auth/me` unauthenticated
  check do not produce a false expired-session notification.
- Logout clears local user state even if the server request fails. Password
  changes now immediately send the user to Login because the backend revokes
  every session after that operation.
- Added protected-route coverage for bootstrap loading, signed-out redirect,
  expired-session redirect, and authenticated rendering. The full frontend
  suite passes (8 suites, 32 tests) and the production build completes with
  only the existing unused-type and Create React App maintenance warnings.

## Follow-up

Use the server-issued role returned by `/auth/me` to conditionally show future
administrator features; the backend must still enforce that role.
