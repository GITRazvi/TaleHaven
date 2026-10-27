# Favorites and Library Controls

## Goal

Keep a reader's saved-library state separate from their favorites, while
allowing a title to be favorited only after it has been saved to My Library.

## Scope

- Persist per-user favorite records for AniList metadata titles and MangaDex
  readable titles, scoped to titles that are currently saved in My Library.
- Add protected list, add, and remove endpoints for each source.
- Preserve the existing library APIs and reading-status behavior.
- Show MangaDex favorites in a small My Profile section without adding a
  dedicated favorites page.

## API and data changes

- Add `user_favorite_anilist_manga` and `user_favorite_manga` tables with
  user-scoped, duplicate-safe records.
- Add protected `GET`, `POST`, and `DELETE` routes below `/favorites/anilist`
  and `/favorites/mangadex`.
- MangaDex favorite mutations save the source metadata locally before storing
  the favorite relation, matching existing MangaDex library behavior.
- Enrich `GET /favorites/mangadex` with display-ready MangaDex title and cover
  data while retaining its existing `data` ID list for current clients.
- Filter favorite lists to active library membership and reject attempts to
  favorite a title that has not been added to My Library.

## Implementation steps

1. Add idempotent migrations and database operations with user-isolation
   coverage.
2. Expose validated favorite routes alongside the protected library routes.
3. Filter favorites by library membership and enforce the prerequisite when
   a favorite is added.
4. Return stored MangaDex favorite metadata from the favorite-list route for
   the profile view.
5. Run focused backend tests, rebuild, and restart the backend.

## Validation

- Passed `go test ./modules/database ./controllers`, including favorite
  isolation and independent MangaDex favorite removal coverage.
- Passed the required `go test ./modules/database` before restart.
- Rebuilt and restarted the backend; startup logs confirm the
  `user_favorite_tables` migration completed and authenticated favorite-list
  requests return HTTP 200.
- Extended the database coverage for stored MangaDex favorite display data;
  `go test ./modules/database ./controllers` and the required pre-restart
  `go test ./modules/database` pass.
- The restarted backend returns HTTP 200 for authenticated
  `GET /favorites/mangadex` requests with profile-ready favorite metadata.
- Favorite IDs and profile-ready MangaDex favorites are now filtered to titles
  still in the same user's library; direct favorite requests for unsaved titles
  receive a clear rejection.
- Backend tests and the required pre-restart database test pass after this
  membership rule, and the backend restarted successfully.

## Follow-up work

- A filterable Favorites page can be added later; My Profile will provide the
  initial MangaDex favorites view.
