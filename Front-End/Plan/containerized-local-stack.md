# Containerized Local Stack

## Goal

Serve the TaleHaven React application from a Docker container at the existing
browser URL, `http://localhost:3000`.

## Scope

- Add a multi-stage frontend Dockerfile that builds the React application and
  serves it with Nginx.
- Add Nginx history-fallback configuration for client-side routes.
- Add the frontend service to the existing local Compose stack.

## API or data changes

None. The browser bundle continues using `http://localhost:3001` as the API
base URL, preserving the current frontend/backend contract.

## Implementation steps

1. Build the React bundle in a Node image with the API base URL supplied at
   build time.
2. Copy the build output into a minimal Nginx image configured for SPA routes.
3. Add the Compose service on host port 3000 and verify that it serves the
   built application.

## Validation

- Passed `npm test -- --watchAll=false --runInBand` (6 suites, 12 tests)
  before starting the frontend container.
- `docker compose config --quiet` passed and the frontend image built
  successfully.
- Compose reports the Nginx frontend running on port 3000, and a host request
  to `http://127.0.0.1:3000/` returned HTTP 200.

## Follow-up work

None.
