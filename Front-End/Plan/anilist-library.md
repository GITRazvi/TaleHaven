# AniList ID Library

## Goal

Allow home discovery cards to be saved to My Library by AniList ID and render
their current AniList metadata when the library is opened.

## Scope

- Add heart toggles to AniList home cards.
- Keep those hearts visible on cover hover/focus by using the established
  `manga-library-card` cover-card styling hook.
- Use the AniList library endpoint to render current library title/cover cards.
- Open the on-demand AniList details page when a saved library card is clicked.
- Do not expose reader controls until MangaDex matching is implemented.

## Validation

- Passed `npm test -- --watchAll=false --runInBand`: 4 suites and 5 tests.
- Passed the production build; the existing unused `SyncReadingProgressBody`
  warning remains.
- Confirmed the development server recompiled the updated UI, including saved
  library cards opening the on-demand details page.
