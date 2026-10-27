# Profile Reading Status

## Goal

Turn the signed-in user's saved AniList library into a profile reading list
with per-title reading statuses, while keeping chapter progress separate.

## Scope

- Persist one reading status for each saved AniList manga: `PLAN_TO_READ`,
  `CURRENTLY_READING`, `ON_HOLD`, `READ`, or `DROPPED`.
- Default newly saved titles to `PLAN_TO_READ` without changing existing
  chapter-level progress behavior.
- Return the status with the protected AniList library list.
- Add a protected endpoint for changing the status of an already saved title.
- Permit the endpoint's `PATCH` preflight request through the local CORS
  policy.
- Return per-status and total saved-title counts with the AniList library
  response for profile statistics.
- Persist an optional per-user 1–10 personal rating for saved AniList titles.
- Derive tied top genres from the user's Completed titles using AniList genre
  metadata.
- Count a user's completed MangaDex chapters from TaleHaven reading progress.
- Return the most recently completed AniList title, based on when its profile
  status changed to Completed.
- Include its AniList cover URL in the last-completed statistic when one is
  available.

## API and data changes

- Add `reading_status` to `user_library_anilist_manga` with a constrained
  default of `PLAN_TO_READ`.
- Add `completed_at` to record when a title most recently entered Completed.
- `GET /library/anilist` includes `readingStatus` per item.
- `GET /library/anilist` includes aggregate `stats` counts for every reading
  status and the total saved-title count.
- `GET /library/anilist` includes optional `personalRating` per title and
  `topGenres`, `chaptersRead`, and `lastMangaCompleted` in its aggregate stats.
- `lastMangaCompleted` includes an AniList ID, title, and optional `coverUrl`.
- `PATCH /library/anilist/:anilistId/status` accepts `{ "status": "..." }`.
- `PATCH /library/anilist/:anilistId/rating` accepts an optional integer
  `personalRating` from 1 to 10; `null` clears it.

## Implementation steps

1. Add the idempotent migration and database operations.
2. Validate status and personal-rating updates in the library controller and
   preserve user isolation.
3. Fetch completed-title metadata in AniList-safe batches and calculate tied
   top genres.
4. Add focused database, controller, and AniList-client tests.
5. Merge persisted and active reader progress to count only chapters whose
   final page was reached.
6. Record a completion timestamp when a saved AniList title enters Completed,
   and select the latest completed title for the profile statistics.
7. Resolve that AniList title from the existing batched metadata response and
   return its public cover URL.

## Validation

- Passed `go test ./modules/database ./controllers`, including the user-scoped
  AniList status test.
- Passed the required `go test ./modules/database` before restarting the
  backend.
- Rebuilt and restarted the backend; startup logs confirm the
  `user_library_anilist_reading_status` migration completed.
- Passed the CORS preflight test and verified the live local preflight returns
  HTTP 204 with `PATCH` in `Access-Control-Allow-Methods`.
- Passed the database status-count test for the profile statistics response.
- Passed `go test ./modules/database ./modules/anilist-api ./controllers`,
  including personal-rating validation/user isolation and 25-ID AniList query
  batching coverage.
- Rebuilt and restarted the backend; startup logs confirm the idempotent
  `user_library_anilist_personal_rating` migration completed.
- Passed `go test ./modules/sessions ./modules/database ./controllers`,
  including the completed-chapter and latest-manga progress calculations.
- Rebuilt and restarted the backend; the signed-in profile request returned
  HTTP 200 with the new reading statistics available.
- Passed the required database test and controller/session coverage after
  adding the optional last-read cover URL; the rebuilt backend returned the
  signed-in profile response with HTTP 200 after restart.
- Passed database, controller, and session tests after replacing the reader
  activity field with Last Manga Completed. The new idempotent completion-time
  migration ran during the rebuilt backend restart, and the signed-in profile
  response returned HTTP 200.

## Follow-up work

- MangaDex-only library records are not migrated in this change.
- Status history and automatic book completion remain out of scope.
- Top genres depend on the public AniList metadata request being available.
