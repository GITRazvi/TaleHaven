# AniList Related Manga

## Goal

Show connected manga on a manga metadata page to make sequels, prequels,
spin-offs, and alternate versions discoverable.

## Scope

- Render only relations whose AniList media type is `MANGA`.
- Show each related title's cover when available, title, and relationship
  label; link it to its own TaleHaven metadata page.
- Do not filter related entries based on their chapter or volume totals.
- Do not add unrelated anime or external link cards in this change.

## Validation

- Added component coverage confirming a manga relation is shown and an anime
  relation is excluded.
- `npm test -- --watchAll=false --runInBand` passes with 6 suites and 18
  tests. The frontend was rebuilt and restarted successfully; it returns HTTP
  200 locally. Existing React test and Create React App warnings are outside
  this change's scope.

