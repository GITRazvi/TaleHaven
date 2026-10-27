# Admin dashboard

## Goal

Add a protected administrator dashboard that establishes the navigation and
overview surface for future TaleHaven administration tools.

## Scope

- Show a dashboard only to users whose authenticated profile has the `admin`
  role.
- Display safe aggregate totals supplied by the server.
- Keep the first version read-only; it does not manage users, sessions, chat,
  or content.

## API and data changes

- Consume `GET /admin/dashboard`.
- Reuse the authenticated user role already returned by the session endpoint.

## Implementation steps

1. Review existing routing, auth context, and shared dashboard styling.
2. Add the API client method and an admin-only route guard.
3. Add the dashboard cards and empty-state/error handling.
4. Expose an Admin navigation link only for administrators.

## Validation

- Add or update frontend tests for role-gated routing and dashboard states.
- Run the full frontend test suite and production build.
- Rebuild/restart the frontend and verify it is serving successfully.

## Follow-up work

- Add moderation queues, audit logs, chat-room management, and permissions once
  their workflows are agreed.

## Implementation status

- Added the `/admin` route, role-aware navigation item, and an administrator
  route guard.
- Added a read-only dashboard with aggregate statistic cards, loading, retry,
  and error states.
- Added tests for the route guard and dashboard states.
- Validation passed: `npm test -- --watchAll=false --runInBand` (43 tests),
  production build, and a rebuilt frontend container serving `/admin`.
