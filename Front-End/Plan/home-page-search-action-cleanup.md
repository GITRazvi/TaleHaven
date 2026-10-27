# Home Page Search Action Cleanup

## Goal

Remove the redundant MangaDex chapter-search button from the home page.

## Scope

- Remove only the extra `Search MangaDex Chapters` action from the home page.
- Preserve the existing navigation and search routes.

## API and data changes

None.

## Implementation steps

1. Remove the extra home-page button markup.
2. Run frontend tests and confirm the development server recompiles.

## Validation

- Passed `npm test -- --watchAll=false --runInBand` (6 suites, 11 tests).
- The active development server recompiled without type-check issues.

## Follow-up work

None.
