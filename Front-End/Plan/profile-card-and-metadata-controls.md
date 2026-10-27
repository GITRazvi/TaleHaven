# Profile Cards and Metadata Controls

## Goal

Make My Profile manga tiles cover-focused and move personal reading controls to
the AniList manga metadata page.

## Scope

- Hide My Profile tile titles until hover or keyboard focus, using an overlay
  on the cover image.
- Remove reading-status and personal-rating selectors from profile tiles.
- Show those selectors in the AniList details page for titles already saved in
  My Library.
- Preserve library filtering, favorites, and existing metadata behavior.

## API and data changes

None. The details page will reuse the existing authenticated AniList library,
status, and personal-rating endpoints.

## Implementation steps

1. Replace profile tile text and selectors with accessible title overlays.
2. Load the selected title's saved-library entry on its AniList details page.
3. Add a compact status and rating controls panel to saved title metadata.
4. Update focused frontend tests and validate the development build.

## Validation

- Passed `npm test -- --watchAll=false --runInBand` (6 suites, 12 tests),
  including details-page status and rating updates for a saved manga.
- The active development server recompiled successfully with no type-check
  issues.

## Follow-up work

Consider a dedicated mobile interaction if reader feedback shows the title
overlay needs a different touch-device behavior.
