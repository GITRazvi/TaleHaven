# Containerized Local Stack

## Goal

Run the TaleHaven backend as a Docker service alongside its existing Redis and
PostgreSQL dependencies, with a browser-accessible frontend container.

## Scope

- Add a production Dockerfile for the Go backend.
- Extend the existing Compose stack with backend and frontend services.
- Configure the backend container to use Compose service hostnames for Redis
  and PostgreSQL.
- Persist request-generated backend images in a named Docker volume.

## API or data changes

None. The backend remains available on host port 3001 and keeps its existing
database schema and routes.

## Implementation steps

1. Add a multi-stage backend Dockerfile that builds only the application entry
   point, avoiding the repository's standalone utility programs.
2. Add a `backend` Compose service, dependency health ordering, configuration,
   port mapping, and image volume.
3. Add the frontend service to the same Compose stack and validate the running
   containers and backend startup logs.

## Validation

- Passed `go test ./modules/database` before starting the backend container.
- `docker compose config --quiet` passed and the backend image built
  successfully.
- Compose reports the backend running on port 3001, Redis running, and
  PostgreSQL healthy.
- Backend startup logs confirm Redis and PostgreSQL connections, migrations,
  and the port-3001 listener. A request to the backend root returned the
  expected 404 from its no-route handler.

## Follow-up work

None.
