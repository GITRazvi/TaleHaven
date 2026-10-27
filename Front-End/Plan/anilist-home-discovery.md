# AniList Home Discovery Sections

## Goal

Replace the single MangaDex Popular Manga block on the home page with several
AniList-backed public-metadata discovery sections.

## Scope

- Render Trending Now, All-Time Popular, Top Rated, and New Releases manga.
- Fetch each section from TaleHaven's backend, not from AniList in the browser.
- Show cover images and compact title cards with independent loading and error
  states for each section.
- Open a dedicated AniList details page only when a user selects a card.
- Keep the existing MangaDex search page, reader, and library controls
  unchanged until title matching exists.
- Add a separate MangaDex-backed “Read on TaleHaven” row for titles that
  advertise English chapter availability, with direct links to their chapter
  pages and independent pagination.
- Give each MangaDex reading card the same signed-in heart control as other
  home-page cards, using the existing MangaDex-specific library endpoints.

## API and data changes

- Consume `GET /anilist/manga/:section?page=&perPage=`.
- Consume the existing `GET /manga/popular?limit=&offset=` endpoint for the
  readable MangaDex row; it already defaults to English availability.
- Use the existing protected MangaDex library list/add/remove endpoints to
  render and toggle saved-heart state for readable cards.
- No authentication or database changes.

## Implementation steps

1. Add typed frontend client methods and AniList metadata response types.
2. Replace the single popular-list state with per-section list state.
3. Render the four sections as responsive rows of metadata cards.
4. Update home-page tests and run frontend tests/build checks.
5. Add independently loaded MangaDex reading cards and a Load more control
   that requests only the next batch when the reader asks for it.
6. Load MangaDex saved IDs for signed-in readers and toggle the card heart
   without navigating away from the home page.

## Validation

- Added home-page tests that verify all four sections request their own first
  page and that loading more affects only the selected section.
- Passed `npm test -- --watchAll=false --runInBand`: 4 suites and 5 tests.
- Passed a production frontend build. It retains the existing unused
  `SyncReadingProgressBody` warning.
- Confirmed the running frontend development server recompiled successfully
  after the home-page update.
- Prevented React development strict-mode effect replay from issuing a second
  initial request for every discovery section.
- Added the independent Read on TaleHaven MangaDex row with English-available
  popular titles, direct MangaDex-detail navigation, and demand-driven
  pagination.
- Passed the full frontend suite (6 suites, 8 tests) and production build;
  the running frontend hot-recompiled successfully. Confirmed the local
  MangaDex endpoint returned a readable title card.
- Added signed-in MangaDex library-heart coverage; the full frontend suite
  passes with 6 suites and 9 tests, the production build passes, and the
  running frontend hot-recompiled successfully.

## Remaining follow-up work

- Add AniList details-page navigation.
- AniList cards remain metadata-first; matching individual AniList titles to
  MangaDex reader entries is still intentionally out of scope.
