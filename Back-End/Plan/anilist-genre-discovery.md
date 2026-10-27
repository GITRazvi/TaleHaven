# AniList Genre Discovery

## Goal

Let users open a public, paginated AniList manga listing for a genre selected
on a manga's details page.

## Scope

- Add a typed AniList GraphQL query that filters manga by one genre.
- Add `GET /anilist/manga/genre/:genre?page=&perPage=` with existing safe
  pagination rules and non-adult results.
- Keep query documents and variables server-controlled; do not accept raw
  GraphQL from clients.

## API and data changes

- The new route returns the existing paged AniList discovery response shape.
- `genre` is decoded from the path and forwarded as a GraphQL variable.

## Implementation steps

1. Add the genre-filtered query and client method.
2. Expose the controller route and validate the requested genre.
3. Cover the client and controller behavior with focused tests.

## Validation

- Passed `go test ./modules/anilist-api ./controllers`.
- Passed the required `go test ./modules/database` before the backend restart.
- Rebuilt and restarted the backend; a live Fantasy request returned HTTP 200
  with one result and `hasNextPage: true`.

## Follow-up work

- Multiple-genre filters and a dedicated genre browser are out of scope.
