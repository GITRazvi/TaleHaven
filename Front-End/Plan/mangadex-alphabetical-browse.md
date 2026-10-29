# MangaDex alphabetical browse

## Goal

Show a paginated A-to-Z MangaDex catalog list on the Manga Search page while
preserving the separate AniList metadata search bar and filters.

## Scope

- Load an initial MangaDex browse page when the Manga Search page opens.
- Render compact cover, title, and description cards that open the existing
  MangaDex title page.
- Offer accessible Previous and Next controls plus compact direct page buttons:
  first/last pages, two pages on either side of the active page, and ellipses
  for skipped ranges.
- Keep MangaDex browse loading and error states separate from AniList search
  results, so a temporary failure from one catalog does not hide the other.

## API and data

- Consume `GET /manga/browse?limit=&offset=` through the typed backend client.
- Use the returned total, limit, and offset to calculate page controls.

## Implementation

1. Add typed MangaDex browse models and an API-client method.
2. Add independent browse state and lifecycle loading to Manga Search.
3. Render the A-to-Z list and compact direct-page navigation below AniList
   search results.
4. Add component coverage for the initial browse request, pagination, and
   ellipsis behavior for large catalog totals.

## Validation

- Run the frontend test suite and production build.
- Restart/rebuild the frontend after the implementation and check its startup
  logs.
- `npm test -- --watchAll=false --runInBand` passes: 16 suites and 64 tests.
- `npm run build` succeeds. It retains the pre-existing unused
  `SyncReadingProgressBody` warning and Create React App maintenance warnings.
- The Docker frontend rebuilt and started successfully; `/manga/search`
  returns HTTP 200.
- Added compact direct-page controls. For a large catalog, page 6 renders
  `1 … 4 5 6 7 8 … last`, with the active page clearly identified. The full
  frontend suite still passes (16 suites, 64 tests), and the Docker frontend
  was rebuilt and restarted successfully.

## Follow-up

- Consider letter jumps or a browse-only title filter if readers request faster
  navigation through the full catalog.
