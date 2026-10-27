# AniList External Reading Links

## Goal

Give readers a clear **Read on** fallback on manga metadata pages when AniList
supplies public external reading-platform links.

## Scope

- Consume the existing AniList external-link data already returned by the
  backend.
- Show only `STREAMING` links with valid HTTP(S) URLs.
- Label links with the provider and language, and open them in a new tab.
- Keep TaleHaven's Start reading action when local chapters exist, but do not
  show an unavailable-chapters message when readers can use external links.

## Validation

- Added focused component coverage for streaming-link rendering and exclusion
  of non-streaming or unsafe links.
- `npm test -- --watchAll=false --runInBand` passes with 6 suites and 20
  tests. The production frontend rebuild succeeded; existing React test and
  Create React App maintenance warnings remain outside this change's scope.
- Added coverage confirming the unavailable-chapters message is absent when
  streaming links are shown. The frontend suite still passes with 20 tests,
  and the production frontend was rebuilt and restarted successfully.

