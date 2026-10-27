# Homepage Continue Reading

## Goal

Add a useful signed-in home-page section for manga that the reader has started
and saved, plus titles saved as Plan to Read.

## Scope

- Load the current user's MangaDex library, AniList library, and merged
  reading-progress snapshot through existing protected APIs.
- A MangaDex title is eligible only when it is saved in the library and at
  least one chapter is complete (`currentPage >= totalPages`).
- Add AniList titles with the `PLAN_TO_READ` status even if they have no
  chapter progress.
- Prefer the newest unfinished chapter as the resume target. Opening it uses
  the reader's existing saved-page behavior; completed-only titles open their
  chapter list to select the next release.
- Resolve each displayed started title's actual last-read chapter metadata,
  then label a new chapter only if the same scanlation group has a numeric
  chapter number greater than the last-read chapter number. Do not count a
  different group's release.
- Remove a fully completed title from Continue Reading when its last-read
  scanlation group has no unread chapter; retain titles with an unfinished
  chapter in progress and keep Plan to Read separate. An unread earlier gap
  is as valid as a newer release.

## API and Data Changes

No new API contract. Use `GET /library`, `GET /library/anilist`,
`GET /reading-progress`, `GET /chapter/:chapterId`, and a filtered
`GET /manga/:mangaId/chapters` request after authentication.

## Implementation Steps

1. Add small view-model helpers that safely interpret progress timestamps and
   completed chapter records.
2. Fetch and merge the three authenticated sources only for signed-in users.
3. Resolve the last-read chapter's group/number and check only that group for
   a higher-numbered chapter, limiting these extra checks to the small Home
   section rather than the whole library.
4. Render a clear Continue Reading section with resume actions, same-group
   release state, and a Plan to
   Read subsection; preserve the public discovery page for signed-out users.
5. Add focused component coverage for eligibility, same-group detection, and
   navigation.

## Validation

- Unauthenticated visitors make no private-library/progress requests.
- Only saved MangaDex titles with a completed chapter appear as started.
- Plan to Read titles appear even without reading progress.
- Resume opens the selected chapter, while completed-only titles open their
  chapter list.
- A release from another scanlation group never produces a new-chapter label.

## Implementation Status

- Added the signed-in home-page Continue Reading section.
- It merges the existing MangaDex library and progress snapshot, admitting a
  MangaDex title only after at least one completed chapter, then selecting the
  most recently read unfinished chapter as its resume target.
- Added AniList `PLAN_TO_READ` titles as a clearly labelled Plan to Read
  subsection.
- Added focused coverage for eligibility and the complete frontend suite now
  passes (8 suites, 33 tests); the production build succeeds with existing
  Create React App and unused-type warnings only.
- For up to six recent started titles, the page now resolves the actual
  last-read chapter's language, scanlation-group ID, and chapter number. It
  labels a new chapter and offers a direct link only when that same group has
  a higher-numbered chapter.
- Removed the manual-completion sentinel from local page-tracker state, so a
  completed chapter cannot render `2147483647` as its last-read page.
- Saved scanlation groups are no longer silently replaced with All groups when
  a response currently has no releases for that group. The selector retains a
  clear unavailable-group option and waits for preference reads/writes before
  fetching the filtered chapter list.
- Added regression coverage for both cases; the frontend suite now passes
  (8 suites, 35 tests).
- The reader now recognizes completed progress before page loading, including
  the legacy manual-completion marker. Replays start safely at page 1 and do
  not write that page back as unfinished progress. Regression coverage now
  verifies both streamed and left/right replays (8 suites, 36 tests).
- Fully completed titles now leave Continue Reading unless their exact
  last-read scanlation group has an unread chapter. This includes earlier
  skipped chapters as well as newer releases; regression coverage now covers
  both cases (8 suites, 38 tests).

## Follow-up

- Add “new chapters available” once a dependable per-title release comparison
  is available.

## Regression Follow-up

- Do not retain or render the manual-completion sentinel as a last-read page.
- Await restoring the saved scanlation group before fetching chapters after a
  refresh, and keep the selected language/group controls in sync with the
  server response.
- Replaying an already completed chapter must not overwrite completion with
  page 1. Detect completion before page loading, use page 1 safely (including
  for historical manual-completion values), and skip progress writes for that
  replay session.
