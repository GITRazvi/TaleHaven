# AniList Metadata Client

## Goal

Add a backend-only AniList GraphQL client that can retrieve public manga
metadata without exposing the upstream API directly to the frontend.

## Scope

- Add reusable GraphQL query documents for manga search and manga details by
  AniList media ID.
- Add a typed Go client that sends JSON `POST` requests with query variables.
- Model the public metadata needed for future search and details pages.
- Include public AniList favourites and all-time rating/popularity rankings in
  the manga-details response for reader-facing statistics.
- Handle transport, HTTP, JSON, and GraphQL errors consistently.
- Add unit tests using a local HTTP test server.

## API and data changes

- No TaleHaven routes, database migrations, or frontend changes in this step.
- Public AniList requests require no account, API key, or OAuth token.
- The client is intentionally uncached in this initial step; any later cache
  must be short-lived and explicitly scoped to public response reuse.

## Implementation steps

1. Create `modules/anilist-api` with query documents and typed response models.
2. Implement `SearchManga` and `GetMangaByID` against AniList's GraphQL
   endpoint using request contexts and configurable HTTP clients for tests.
3. Test request shape, response decoding, and GraphQL error handling.
4. Run focused tests, the required database suite, rebuild, and restart the
   backend before handoff.

## Validation

- Formatted the new client package with `gofmt`.
- Passed `go test ./modules/anilist-api`, including request method, JSON query
  variables, metadata decoding, input validation, and GraphQL rate-limit error
  handling. The test uses a local HTTP server and does not contact AniList.
- Passed the required `go test ./modules/database` suite.
- Passed `go test ./controllers ./modules/anilist-api`.
- Rebuilt the backend from `main.go` and restarted it successfully on port
  3001. A root `go build .` remains unsuitable because existing standalone
  backend utility files each declare their own `main` function.
- Added favourites and ranking decoding coverage. `go test
  ./modules/anilist-api ./controllers` and the required `go test
  ./modules/database` pass; the rebuilt backend starts successfully on port
  3001.

## Remaining follow-up work

- Add TaleHaven REST endpoints that expose AniList search and details.
- Define the safe matching workflow between AniList media IDs and MangaDex
  manga IDs before enabling reader controls for AniList-backed titles.
