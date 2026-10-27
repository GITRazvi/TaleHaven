# Remove MangaHook integration

## Goal

Remove the deleted MangaHook service's backend integration and restore
MangaDex as the only TaleHaven manga-search provider.

## Scope

- Delete the MangaHook proxy controller, routes, tests, environment-based
  upstream configuration, and obsolete proxy plan.
- Restore `GET /manga/search` to the MangaDex search API, including its title,
  pagination, tag, and language filters.
- Preserve the existing MangaDex alternative-title fields and display-title
  selection; they do not depend on MangaHook.

## API and data changes

- Remove all `/mangahook/*` routes.
- `GET /manga/search` again returns the MangaDex paged response without
  provider-specific raw payload fields.
- No database changes.

## Implementation steps

1. Remove controller registration and the proxy controller/test files.
2. Restore the MangaDex search handler and remove MangaHook-only tests.
3. Update the scraper trial documentation so it has no MangaHook references.
4. Run backend tests, restart the backend, and verify removed routes return
   404 while the restored search handler is registered.

## Validation

- Passed `go test ./controllers ./modules/database -count=1`.
- Restarted the backend and confirmed it is listening on port 3001.
- Confirmed a former `/mangahook/search/test` route now returns HTTP 404.
- Confirmed `GET /manga/search` invokes the MangaDex handler. Its live request
  currently receives MangaDex's maintenance HTTP 503, which the backend
  surfaces as HTTP 500.

## Remaining follow-up work

- Evaluate the separate MangaKakalot scraper trial independently; it is not a
  MangaHook replacement until it can pass the source site's Cloudflare check.
