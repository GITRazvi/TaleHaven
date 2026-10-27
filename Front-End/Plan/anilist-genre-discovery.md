# AniList Genre Discovery

## Goal

Allow a reader to select a genre on an AniList manga details page and discover
more public AniList manga in that genre.

## Scope

- Make displayed AniList genre badges accessible links.
- Add a paginated genre-results route using the existing AniList card style
  and library heart actions.
- Add a six-card **Top-rated in [genre]** row to manga details, using the
  current title's first AniList genre, AniList score ordering, and excluding
  the title currently being viewed. Keep this recommendation row as the final
  content section on the page so later details sections appear above it.
- Do not add database state; genre discovery stays public and on demand.

## API and data changes

- Call `GET /anilist/manga/genre/:genre?page=&perPage=` through the backend
  client.
- Add `/anilist/genre/:genre` to the application router.

## Implementation steps

1. Add the backend client method and route.
2. Build the genre-results page with loading, empty, error, and load-more
   states.
3. Link genre badges from the details page and add focused tests.
4. Reuse the score-sorted genre endpoint for the details-page recommendation
   row without adding a separate recommendation source.

## Validation

- Passed `npm test -- --watchAll=false --runInBand`: 6 suites and 23 tests,
  including the recommendation row behavior and final-section placement.
- Passed the production build and restarted the frontend container; it is
  serving `http://localhost:3000` with HTTP 200.
- Existing build warnings remain for unused `SyncReadingProgressBody` and
  `COMMON_GENRES`, plus the Create React App Babel/Browserslist notices.

## Follow-up work

- Multiple selected genres and richer AniList tag filters are out of scope.
