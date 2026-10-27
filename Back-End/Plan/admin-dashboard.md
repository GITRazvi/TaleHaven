# Admin dashboard

## Goal

Provide a protected administrator-only API that supplies the initial TaleHaven
admin dashboard with safe, aggregate information.

## Scope

- Add an admin-only dashboard endpoint.
- Return aggregate account and library counts only; do not expose user records,
  credentials, session data, or moderation actions yet.
- Reuse the existing authenticated opaque-session role check.

## API and data changes

- `GET /admin/dashboard` requires an authenticated user with the `admin` role.
- The response contains aggregate totals for users, library entries, favourites,
  and completed/read library entries.
- No schema migration is required.

## Implementation steps

1. Review the current router, authentication middleware, and database access
   patterns.
2. Add the aggregate database query and admin controller route.
3. Add an admin dashboard page and protected route in the frontend.
4. Validate role rejection, role access, and the dashboard UI.

## Validation

- Run the required database test before a backend restart.
- Run focused Go tests plus frontend tests and a production build.
- Rebuild/restart both services and confirm their startup logs.

## Follow-up work

- Add audited moderation actions, user lookup, reports, and chat administration
  only after their requirements are defined.

## Implementation status

- Added `GET /admin/dashboard`, protected by the existing authenticated session
  middleware followed by a server-side `admin` role requirement.
- Added aggregate-only totals for accounts, AniList and MangaDex library entries,
  favorites, and stored chapter-progress records.
- Added unit coverage for the administrator role middleware.
- Validation passed: `go test ./modules/database`, `go test ./middlewares ./controllers`,
  and a rebuilt backend container with the anonymous endpoint correctly returning
  `401`.
