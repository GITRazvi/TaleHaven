# Profile Reading Status

## Goal

Present the existing saved AniList library as a reader's My Profile page with
book-level reading lists, inspired by MAL-style organization.

## Scope

- Rename the current navigation labels to Settings and My Profile without
  changing their URLs.
- Replace the My Profile content with an initials-based profile-photo
  placeholder, saved-title count, and status sections.
- Let the reader change a saved title's status from My Profile.
- Label the in-progress state as Reading in the profile UI.
- Retain the filled heart on each saved profile card as the remove-from-library
  action.
- Show status statistics at the top of My Profile and present saved titles in
  one filterable card grid rather than separate status containers.
- Keep the profile statistics in one compact responsive summary container.
- Present the profile statistics as a colorful Manga Stats panel with a
  status-distribution bar, color-coded status legend, and quick details.
- Show the reader's top completed-book genres in the profile summary.
- Show the number of chapters completed through TaleHaven's MangaDex reader.
- Show the most recently completed AniList title in the profile statistics
  panel.
- Show a compact cover thumbnail beside the last-completed title when
  available.
- Let readers set or clear an optional personal 1–10 rating directly below a
  saved book's status selector.
- Keep these book-level statuses separate from chapter progress.

## API and data changes

- Consume `readingStatus` from the AniList library response.
- Consume aggregate profile statistics from the AniList library response.
- Consume `chaptersRead` from the aggregate profile statistics response.
- Consume optional `lastMangaCompleted` from the aggregate profile statistics
  response, including its optional cover URL.
- Send a protected status-update request from My Profile.
- Send protected personal-rating updates from My Profile.

## Implementation steps

1. Add frontend types and backend-client methods.
2. Redesign the current library page as My Profile with statistics and a
   status filter.
3. Style the statistics as a responsive, color-coded Manga Stats panel.
4. Add focused UI tests for loading, status updates, personal ratings, and
   top-genre and completed-chapter display.
5. Rename the latest-title statistic to Last Manga Completed and render the
   AniList title and cover supplied for the latest completed status change.

## Validation

- Passed `npm test -- --watchAll=false --runInBand`: 6 suites and 7 tests.
- Passed the production build; the existing unused `SyncReadingProgressBody`
  warning remains.
- Passed the focused My Profile status and heart-removal test and confirmed
  the frontend recompiled successfully.
- The in-progress status is labeled Reading in the profile UI.
- Passed statistics/filter coverage in the focused My Profile test; the full
  frontend suite and production build pass.
- Passed the focused profile test and confirmed hot recompilation after the
  compact statistics layout update.
- Passed the full frontend suite with the personal-rating and top-genre test
  coverage (6 suites, 7 tests), and the production build. The pre-existing
  unused `SyncReadingProgressBody` lint warning remains.
- Passed the full frontend suite and production build after the Manga Stats
  redesign; the running frontend hot-recompiled successfully.
- Passed the focused and full frontend test suites plus the production build
  for Chapters Read and Last Manga Read; the restarted frontend compiled
  successfully.
- Passed the full frontend suite (6 suites, 7 tests) and production build
  after adding the compact Last Manga Read cover; the restarted frontend
  compiled successfully.
- Passed the full frontend suite (6 suites, 9 tests) and production build
  after replacing Last Manga Read with Last Manga Completed; the restarted
  frontend compiled successfully.

## Follow-up work

- Actual profile-photo upload and user-editable profile fields are out of
  scope for this first profile pass.
