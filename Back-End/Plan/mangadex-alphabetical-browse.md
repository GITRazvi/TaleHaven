# MangaDex alphabetical browse

## Goal

Provide a stable, paged MangaDex catalog browse endpoint ordered from A to Z,
without changing the AniList-first metadata search experience.

## Scope

- Add a public `GET /manga/browse` route backed by MangaDex's title-ascending
  catalog ordering.
- Keep the existing cover, author, artist, and safe-content request options so
  browse cards have the same useful metadata as other MangaDex catalog cards.
- Return the existing simplified manga payload with total, limit, and offset
  so the frontend can render ordinary page controls.
- Cache page responses using the existing short-lived MangaDex catalog cache.

## API and data

- `GET /manga/browse?limit=&offset=` returns a paged MangaDex catalog page
  ordered by title ascending.
- No account data or library state is added to this public endpoint.

## Implementation

1. Add a MangaDex client request with `order[title]=asc`.
2. Expose it from the existing manga controller before the parameterized manga
   detail route.
3. Transform entries through the existing `SimpleManga` mapper and include
   pagination metadata.

## Validation

- Run Go formatting and focused backend tests.
- Run the required database test before rebuilding and restarting the backend.
- `go test ./controllers` passes, including invalid browse pagination coverage.
- The required `go test ./modules/database` was attempted before restart but
  this Windows host cannot access its forwarded PostgreSQL port (`localhost:5433`).
- The Docker backend rebuilt and started successfully; a live
  `GET /manga/browse?limit=2&offset=0` returned MangaDex entries and paging
  metadata.

## Follow-up

- Consider an optional letter jump control only if ordinary page navigation
  proves cumbersome for the full MangaDex catalog.
