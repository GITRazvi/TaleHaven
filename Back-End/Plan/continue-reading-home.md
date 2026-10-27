# Homepage Continue Reading

## Goal

Give signed-in readers a home-page view of library titles that can be resumed
or that they explicitly plan to read.

## Scope

- Reuse existing authenticated library and reading-progress routes; no new
  server route or database table is required for this iteration.
- Treat a MangaDex library title as started only when at least one saved
  chapter progress record is complete (`currentPage >= totalPages`).
- Leave AniList `PLAN_TO_READ` titles eligible even when no MangaDex chapter
  progress exists.
- For the small recent set displayed on Home, reuse existing chapter detail
  and filtered chapter-list routes to resolve the actual last-read chapter's
  scanlation group and confirm whether that same group has a higher numbered
  release.
- Do not return a fully completed title as a Continue Reading recommendation
  unless the same scanlation group contains at least one chapter the user has
  not completed, whether it is a newer release or an earlier skipped chapter.

## API and Data Changes

None. The frontend consumes the existing `GET /library`,
`GET /library/anilist`, `GET /reading-progress`, `GET /chapter/:chapterId`,
and filtered `GET /manga/:mangaId/chapters` responses.

## Implementation Steps

1. Keep the existing routes private through session authentication.
2. Let the frontend combine the library and merged progress snapshot per
   signed-in user.
3. Use the last-read chapter's actual scanlation group and numeric chapter
   value to look for a later chapter in the same group only.
4. Add focused UI validation for empty, planned, resumable, and new-release
   states.

## Validation

- Existing database tests pass unchanged before any server restart.
- The section never exposes another user's library or progress because all
  source endpoints remain authenticated and user-scoped.

## Implementation Status

- No backend code or schema change was required. The existing session-protected
  library and progress routes supply the user-scoped data consumed by the
  frontend.
- Existing chapter detail and same-group filtered chapter routes also provide
  the metadata needed to distinguish genuinely newer releases from releases
  published by another group.
- The existing persisted per-user manga/language group preference remains the
  source of truth; the frontend no longer overwrites it with All groups when a
  filtered response is empty.
- Existing progress records remain the source of truth for completed chapters;
  the frontend no longer overwrites them when a completed chapter is replayed.
- Continue Reading now excludes fully completed titles unless the existing
  same-group availability check finds an unread chapter, whether newer or a
  skipped earlier release.

## Follow-up

- A later optimized aggregate endpoint may replace the client-side merge if
  library sizes or page-load costs make it necessary.

## Regression Follow-up

- Preserve scanlation-group preferences through a refresh while keeping the
  independently selected language intact. Existing settings routes must return
  the saved group before the chapter list is queried.
- The manual chapter-completion sentinel is an internal marker only; it must
  never be surfaced as a reader page number.
- A reopened completed chapter must retain its completed progress record. The
  reader may start it at page 1 for replay, but must not write that page back
  as unfinished progress.
