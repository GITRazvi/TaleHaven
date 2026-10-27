# Favorites and Library Controls

## Goal

Make the library action visually distinct from favorites: `+` adds a title to
My Library, `−` removes it, and the heart appears only after the title is in
the library.

## Scope

- Apply the two controls to AniList discovery/genre cards, MangaDex readable
  cards, MangaDex details, and saved My Profile cards.
- Load each signed-in user's favorite IDs for the relevant source and update
  controls optimistically after a successful mutation.
- Keep signed-out action behavior consistent by directing the reader to login.
- Hide every favorite heart until its AniList or MangaDex title is saved in My
  Library; removing the title hides the heart again.
- Show every loaded AniList and MangaDex favorite together in one My Profile
  Favorites section, while keeping AniList reading status and rating controls
  on saved-library titles.
- Do not add a standalone favorites-list page in this change.

## API and data changes

- Consume the protected favorite-ID list and toggle endpoints for AniList and
  MangaDex titles.
- Combine saved AniList favorites and display-ready MangaDex favorites in the
  My Profile Favorites section.
- Continue using the existing My Library APIs for the `+`/`−` control.

## Implementation steps

1. Add typed frontend client methods for favorite IDs and mutations.
2. Replace library hearts with explicit library plus/minus controls.
3. Render favorite hearts only for saved titles and update focused UI tests.
4. Render AniList and MangaDex favorite covers, titles, detail navigation,
   and filled hearts that remove favorites from My Profile.
5. Run frontend tests/build and confirm the development server recompiles.

## Validation

- Passed `npm test -- --watchAll=false --runInBand`: 6 suites and 10 tests,
  including independent MangaDex favorite and library controls.
- Passed the production build with only the existing unused
  `SyncReadingProgressBody` warning.
- Restarted frontend compiled successfully.
- Added My Profile coverage for rendering and removing a stored MangaDex
  favorite; the full frontend test suite still passes (6 suites, 10 tests).
- The production build passes with the existing unused
  `SyncReadingProgressBody` warning, and the restarted development server
  compiled successfully.
- Added coverage proving an unsaved MangaDex title has no favorite heart; the
  full test suite passes with 6 suites and 11 tests. All card and details
  views now show the heart only for saved titles.
- Replaced the MangaDex-only profile block with one Favorites container that
  displays both saved AniList favorites and MangaDex favorites; the full
  frontend test suite continues to pass (6 suites, 11 tests).

## Follow-up work

- A dedicated Favorites page can reuse these stored records later.
