# AniList Character Pagination

## Goal

Display ten characters initially on a manga details page, then let a reader
append the next pages with an explicit Load more button.

## Scope

- Preserve role ordering across all loaded character pages.
- Show a Load more control only while AniList reports another page.
- Preserve existing favorite and character-details actions for appended items.
- Keep the chapter panel fully loaded and scrollable; this change is limited
  to character portraits.

## Validation

- Added component coverage for the Load more action, appended characters, and
  removal of the control on the final page.
- `npm test -- --watchAll=false --runInBand` passes with 6 suites and 19
  tests. The production frontend was rebuilt and restarted successfully; it
  returns HTTP 200 locally. Existing React test and Create React App warnings
  remain outside this change's scope.

