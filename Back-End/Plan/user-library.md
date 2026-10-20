# User Library Backend Plan

## Goal

Give every authenticated user a private saved-manga library.

## Implemented scope

- Added the `user_library_manga` relation between `users` and shared `manga` metadata.
- Added protected `GET /library`, `POST /library/:mangaId`, and `DELETE /library/:mangaId` endpoints.
- Made add and remove operations idempotent.
- Preserved locally cached cover paths while refreshing manga metadata.
- Corrected PostgreSQL array scanning for persisted manga metadata.
- Clamp chapter-list requests to MangaDex's supported range before forwarding
  them, preventing an oversized client request from becoming a server error.
- Avoid a chapter page-count lookup for bulk manual progress updates, because
  one upstream request per chapter causes MangaDex rate limiting.
- Include live Redis progress in manga-level progress responses so recently
  saved reading pages appear immediately in the details page chapter list.
- Support a validated `startPage` query parameter on the progressive chapter
  stream, allowing the reader to begin downloading at a saved page.
  - Prioritize the saved page, then forward pages, then gradually backfill
    preceding pages nearest-first.
- Add database-backed coverage for library add, list, remove, idempotency, and
  per-user isolation using uniquely created test records that are cleaned up
  after the test.
- Paginate `GET /library` with validated `limit` and `offset` parameters.
  Return the page metadata and total saved-title count so the frontend can load
  an initial batch and request more only when the reader asks for it.

## Validation

- Added and passed a database-backed pagination case that confirms two pages
  contain distinct saved titles and return the same total count.
- Rebuilt the backend server entry point and restarted it after adding
  paginated `GET /library` support.
- Added a database-backed test for library add, list, remove, idempotency, and
  per-user isolation. It creates uniquely named records and removes them on
  completion.
- Ran `go test ./modules/database` successfully against the local PostgreSQL
  database, rebuilt the backend server entry point, and restarted it.
- Ran focused Go package compilation with `go test ./controllers ./modules/database ./modules/mangadex-api`.
- Rebuilt and restarted `TaleHaven-Backend-local.exe`.
- Confirmed the startup log executed the `user_library_manga_table` migration and the server started on port 3001.
- Re-ran focused Go package tests and rebuilt the backend after protecting the
  chapter endpoint against invalid request limits.
- Restarted the backend and verified an incoming `limit=500` chapter request
  returns HTTP 200 with a clamped limit of 100.
- Re-ran focused Go package tests and rebuilt the backend after adding the
  page-count endpoint for manual chapter progress updates.
- Restarted the backend and confirmed the page-count endpoint returns HTTP 200
  and an accurate page total for a known chapter.
- Re-ran focused Go package tests and rebuilt the backend after removing the
  rate-limit-prone bulk page-count route.
- Restarted the backend after removing the unused page-count route.
- Re-ran focused Go package tests and rebuilt the backend after merging live
  Redis progress into progress-list responses.
- Restarted the backend and confirmed the manga progress endpoint returns
  HTTP 200 with the updated code running.
- Ran focused Go package tests, rebuilt and restarted the backend after adding
  `startPage` streaming support.
- Confirmed a local stream request for page 20 announced `startPage: 20` and
  emitted page 20 as its first page.
- Ran focused Go package tests, rebuilt and restarted the backend after adding
  next-page priority and slower nearest-first previous-page backfill.
- Confirmed a local stream from page 36 emitted pages 36, 37, 38, then 35.

## Follow-up work

- Add controller-level tests for invalid library pagination parameters.
