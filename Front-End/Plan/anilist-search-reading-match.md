# AniList search with MangaDex reading matches

## Goal

Show AniList metadata as the single main search result source, then locate MangaDex chapters only after the reader requests them.

## Scope

- Replace the MangaDex search-results request and incompatible MangaDex tag filters with AniList title search.
- Add optional genre, release-year, format, publication-status, and sort
  controls to the AniList search form. Submit the selected filters to the
  backend rather than filtering already-loaded cards in the browser.
- Keep the title input visible on the search page and place the remaining
  filters in a Filters modal overlay, so the primary search experience stays
  compact without changing the upstream request behavior.
- Map every selector option to the appropriate AniList constraint: standard
  genres use `genre`, while catalog concepts such as Isekai, Magic, and Martial
  Arts use `tag`.
- Merge AniList genres with searchable, non-spoiler metadata tags in one
  clickable `Genres` section on manga details. Standard genres open genre
  results and metadata tags open tag-filtered results.
- Hide non-searchable and spoiler-marked AniList metadata tags from the
  details page, keeping the genre list focused on filters readers can use.
- Place the manga title, alternative titles, AniList release date, and unified
  genre badges directly below the cover on the details page.
- Expand the cover-side information with type, known/unknown volumes and
  chapters, publication status, and credited AniList staff authors.
- Add a cover-side Statistics block using public AniList score, all-time
  rating and popularity ranks, member count, and favourites.
- Order the cover-side blocks so Genres appear before Statistics.
- Preserve selected filters in the search URL so refreshes and shared URLs run
  the same upstream-filtered search.
- Route search cards to AniList details.
- Add a Start reading action that resolves a confident MangaDex source and opens its chapter page, or clearly reports that chapters are unavailable.
- Render AniList CDN covers without forcing CORS mode, since the image host does not return browser CORS headers.
- Show whether a verified MangaDex chapter source is available on AniList details pages and let readers open it there.
- Show distinct alternative manga titles from AniList (romaji, native, and
  supplied synonyms) on the details page without repeating the primary title.
- Remove the obsolete MangaDex tag filters from the AniList-only search form.
- Add compact library controls on AniList manga details, and reveal the favorite control only after the title is saved.
- Keep the same library-first control order on the MangaDex chapter-details page: always show library add/remove, and only show favorites for saved titles.
- Keep the latest successful AniList home lists and signed-in profile-library response in short-lived browser storage, so readers can keep browsing saved results during a temporary AniList rate limit or outage.
- Clearly mark stale saved data, provide a retry action, and keep profile cache keys scoped to the signed-in user.

## API and data

- Use the paged AniList search and reading-source APIs.
- Send optional AniList search filters as query parameters on the paged search
  request.
- Cache only public home discovery responses and the current user's already-rendered profile response in local browser storage; cached data is a display fallback, never a substitute for a write request.

## Validation

- Update search and details component tests, including the details-page library and favorite controls.
- Run the frontend suite and rebuild the Docker frontend serving port 3000.
- Passed the frontend suite (6 suites, 12 tests). The rebuilt Docker frontend
  production build succeeded with the existing unused-type warning and now
  returns HTTP 200 on port 3000.
- Removed the forced anonymous CORS mode from AniList cover images; the full
  frontend suite passes and the rebuilt frontend container is healthy.
- Add component coverage for rendering the cached fallback and retry affordance
  when AniList-backed requests fail.
- Passed the frontend suite (6 suites, 15 tests), including cached home and
  profile fallbacks. Rebuilt and restarted the Docker frontend; it is healthy
  and returns HTTP 200 on port 3000.
- Added focused alternative-title coverage. `npm test -- --watchAll=false
  --runInBand` passes with 6 suites and 17 tests; the production frontend was
  rebuilt and restarted. Existing React test and Create React App maintenance
  warnings remain outside this change's scope.
- Added coverage for filter-only URL restoration and submission. The complete
  frontend suite passes (6 suites, 23 tests); the Docker frontend was rebuilt,
  restarted, and serves HTTP 200 on port 3000.
- Added overlay interaction coverage. The complete frontend suite passes
  (6 suites, 23 tests); the Docker frontend was rebuilt, restarted, and serves
  HTTP 200 on port 3000.
- Added tag-selector coverage. The complete frontend suite passes (6 suites,
  23 tests); the rebuilt Docker frontend serves HTTP 200 on port 3000.
- Added non-spoiler tag rendering coverage. The complete frontend suite passes
  (6 suites, 24 tests); the Docker frontend was rebuilt, restarted, and serves
  HTTP 200 on port 3000.
- Added focused coverage for searchable tag badges. The complete frontend
  suite passes (6 suites, 24 tests); the Docker frontend was rebuilt,
  restarted, and serves HTTP 200 on port 3000.
- Added merged genre/tag coverage. The complete frontend suite passes (6
  suites, 24 tests); the Docker frontend was rebuilt, restarted, and serves
  HTTP 200 on port 3000. The production build retains its unrelated existing
  unused-type warning.
- Added cover-side metadata and release-date coverage. The complete frontend
  suite passes (6 suites, 24 tests); the Docker frontend was rebuilt,
  restarted, and serves HTTP 200 on port 3000. The production build retains
  its unrelated existing unused-type warning.
- Added author, publication metadata, and statistics coverage. The complete
  frontend suite passes (6 suites, 24 tests); both Docker services were
  rebuilt and restarted. The frontend returns HTTP 200 and the backend is
  listening on port 3001. The production frontend build retains its unrelated
  existing unused-type warning.
- Added coverage confirming Genres appear before Statistics. The complete
  frontend suite passes (6 suites, 24 tests); the Docker frontend was rebuilt,
  restarted, and serves HTTP 200 on port 3000. The production build retains
  its unrelated existing unused-type warning.

## Follow-up

Offer a manual MangaDex-source picker only if readers need an override for an ambiguous title.
