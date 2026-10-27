# AniList Character Favorites

## Goal

Let readers favorite a character from an AniList manga's details page and see
those characters in a dedicated subsection of My Profile's Favorites area.

## Scope

- Add a favorite-heart control to displayed characters for signed-in readers.
- Load favorite state for the current manga details page.
- Organize My Profile Favorites into Manga and Characters subsections.
- Show favorite character portrait, name, contextual role, and a removal
  control.
- Open a character-details modal from a character portrait or name, showing
  its public biography, fields, and associated manga/anime.
- Allow the portrait and name in the Profile Favorites character cards to open
  that same details view.
- Use one shared character-details modal from both the manga page and Profile
  Favorites so the two entry points cannot diverge.
- Display Profile Favorites characters as portrait tiles, revealing their name
  only on hover while keeping the portrait as the details trigger.
- Keep Profile Favorites character portraits compact at half their prior tile
  height without changing their hover or details behavior.
- Make the full character tile compact in both dimensions, not only its
  portrait height.
- Reveal the Profile Favorites removal heart only while interacting with its
  character tile.

## API / data changes

- Use protected AniList-character favorite list/add/remove endpoints.
- Use typed character favorite display data returned by the backend.

## Implementation steps

1. Add API client methods and types.
2. Implement the details-page toggle and profile character collection.
3. Cover the controls and profile rendering with component tests.
4. Load character details only after a character is selected.
5. Reuse the on-demand details endpoint from Profile Favorites.
6. Extract the character-details panel into a shared component and keep the
   Profile Favorites card's removal control independent from its details
   trigger.
7. Match the manga-tile hover treatment for Profile Favorites characters.
8. Reduce the Profile Favorites character tile height.
9. Use auto-width grid cells and fixed portrait dimensions for compact tiles.
10. Match the manga-control hover and focus behavior for character hearts.

## Validation

- Passed `npm test -- --watchAll=false --runInBand` (6 suites, 12 tests).
- Passed `npm run build`; it retains the existing unused
  `SyncReadingProgressBody` warning and Create React App dependency warnings.
- Restarted the development server and confirmed its final hot reload compiled
  with no type-check issues (aside from that existing lint warning).
- Extended the details-page test to cover opening a character modal and its
  biography plus associated-work content; the full suite still passes.
- Restarted the frontend after the character-details extension; it compiled
  successfully with no type-check issues.
- Passed the full frontend suite after adding the Profile Favorites character
  details trigger and modal.
- Cleanly restarted the frontend; it compiled successfully and is listening on
  port 3000.
- Corrected a Profile character-details type mismatch that blocked development
  compilation; the frontend test suite passes and the restarted server reports
  no type-check issues.
- Profile Favorites character cards now use the shared character-details modal
  and have a full-height clickable trigger; the full frontend suite passes
  (6 suites, 12 tests).
- A clean frontend restart was attempted after validation, but Windows reports
  port 3000 as unavailable without an owning listener; this needs the local
  port reservation to be released before the development server can bind.
- Profile Favorites character portraits now use the manga-style hover name
  overlay and remain the character-details trigger; the full frontend suite
  passes (6 suites, 12 tests).
- Rebuilt the Docker frontend that serves port 3000 after confirming it was
  serving a stale production bundle. The replacement container is healthy,
  deploys the character-tile CSS, and returns HTTP 200 locally.
- Reduced Profile Favorites character portrait tiles from 240px to 120px;
  tests pass and the rebuilt Docker frontend serves the compact CSS over
  HTTP 200.
- Reduced the full Profile Favorites character tile to 120px by 120px rather
  than leaving it stretched across a grid column; tests pass and the rebuilt
  frontend container is healthy on port 3000.
- Character favorite hearts now reveal only while hovering or keyboard-focusing
  their tile; tests pass and the rebuilt frontend container returns HTTP 200.

## Follow-up

Character profiles and a separate all-favorites page remain out of scope.
