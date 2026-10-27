# AniList Manga Characters

## Goal

Return public character portraits, names, and story roles with AniList manga
details.

## Scope

- Extend the existing manga-details GraphQL query with the first 20
  characters (the equivalent of the first two original 10-character pages).
- Return character name, image, and relation role through the existing details
  endpoint.
- Do not persist characters or implement character favorites in this change.

## API and data changes

- `GET /anilist/manga/details/:anilistId` will include a `characters`
  connection inside its existing manga payload.

## Implementation steps

1. Add character connection models and request fields to the AniList details
   query.
2. Increase the initial details response to 20 characters while preserving
   AniList pagination metadata for a future Load more control.
3. Run focused API/client tests, rebuild, and restart the backend.

## Validation

- Passed `go test ./modules/anilist-api ./controllers`.
- Passed the required `go test ./modules/database` before restart.
- Rebuilt the backend and restarted it; it is listening on port 3001.
- Passed `go test ./modules/anilist-api ./controllers` and the required
  `go test ./modules/database` before restart. The backend was rebuilt and
  restarted successfully. A local details request for AniList manga `105398`
  confirmed `perPage: 20` and returned 20 character edges (of 500 available).

## Follow-up work

Add character pagination and character favorites separately.
