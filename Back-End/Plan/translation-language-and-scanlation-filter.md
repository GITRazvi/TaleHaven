# Translation language and scanlation-group chapter filtering

## Goal

Let each reader save a preferred MangaDex translation language in Settings, then
offer a per-manga scanlation-group filter that keeps the default chapter list
useful without hiding a reader's ability to choose a specific translation.

## Product decisions

- The profile setting defaults to `all`, meaning all translated languages are
  available rather than the current implicit English-only server default.
- The chapter-page language control initially uses the saved profile preference
  but remains adjustable for that viewing session.
- The scanlation-group control defaults to **All groups**.
  - In this mode, return one selected release for each unique
    `language + volume + chapter` key.
  - Per language, favor the group with the greatest number of distinct numbered
    chapters and fill its gaps from the remaining groups.
  - A reader selecting a specific group sees only that group's releases,
    stopping reader next-chapter navigation at its first missing numbered
    chapter rather than falling back to another group.
- Scanlation groups are language-scoped. When **All languages** is selected,
  keep the merged list globally ordered by chapter number and require a
  specific language before choosing an individual group. This prevents an
  English group choice from unexpectedly filtering another language.
- Special, unnumbered, and non-standard chapters must retain a stable unique
  key instead of collapsing into one duplicate entry.

## Scope

- Add a persistent, authenticated user setting for preferred translation
  language; use `all` as the default for existing and new users.
- Expose the saved preference through a small Settings API and validate it
  against language codes returned by MangaDex rather than trusting arbitrary
  client input.
- Extend simplified chapter data with the stable scanlation-group ID in addition
  to its display name.
- Replace the chapter route's implicit English fallback with an explicit
  language/group-aware chapter listing response that also provides available
  languages and group coverage counts for the active language.
- Fetch enough MangaDex chapter metadata to calculate group coverage and
  deduplicate before applying API pagination; retain bounded, predictable
  upstream pagination and error handling.
- Keep duplicate selection deterministic: highest distinct numbered-chapter
  coverage first, then a stable group/name/ID ordering, then a stable release
  ordering for same-key releases.
- Persist the reader's named scanlation-group selection per MangaDex manga and
  language, restoring it on later visits to that manga's chapter list.
- Update reader next-chapter/preload behavior so it follows the same selected
  language and group path instead of the current hard-coded English lookup.
- Sort all-language results by chapter number before language, so a translated
  Chapter 50 never appears ahead of Chapter 1 merely due to its language.
- Persist a reader display mode (`vertical`, left-to-right, or right-to-left)
  alongside the existing language preference so the reader opens in the
  reader's chosen layout.

## API and data changes

- Add a nullable/defaulted `preferred_translation_language` user field (stored
  as `all` when no specific language is selected).
- Add a defaulted `preferred_reading_mode` user field. Expose it with the
  existing reader-settings payload and accept the three supported values only.
- Add authenticated read/update endpoints for reader settings, for example
  `GET/PATCH /settings/reading`.
- Evolve `GET /manga/:mangaId/chapters` with explicit filter inputs such as
  `language=all|<code>` and `scanlationGroup=all|<group-id>`; preserve support
  for existing callers during migration.
- Return chapter metadata plus:
  - available translation languages;
  - group options for a selected language, each with stable ID, display name,
    and distinct chapter count;
  - the effective language and group mode used to produce the list.
- Keep reading progress keyed by MangaDex chapter ID. No progress migration is
  required because each translation release already has its own chapter ID.
- Add authenticated read/write endpoints for a manga-specific scanlation-group
  preference. A missing preference continues to mean **All groups**.
- Add a bounded single-page chapter endpoint for paged readers. It returns one
  locally served page URL and the chapter's total page count, without starting
  a whole-chapter download or background page prefetch.
- Add a lightweight page-count endpoint for reader resume validation. It
  obtains metadata only; it does not download an image or preload any page.

## Implementation steps

1. Inspect and extend the user/profile settings storage and authenticated route
   conventions; add a reversible migration and database operations/tests.
2. Add the backend settings controller and validation. Decide the language
   option list from MangaDex response data, with `all` always allowed.
3. Preserve scanlation-group IDs in MangaDex transformations and add pure
   helpers for chapter keys, group coverage, group ordering, merged selection,
   and explicit-group filtering.
4. Update the chapter controller/client to request all required source pages,
   build a filtered result, and paginate only the final selected list. Keep
   request limits and upstream errors explicit.
5. Thread language/group selection through reader navigation and next-chapter
   prefetch so auto-navigation never changes translation source unexpectedly.
6. Update frontend request types after the response contract is stable.
7. Extend the reader-settings model with a persisted display mode while
   preserving vertical reading as the safe default for existing users.
8. Keep the existing progressive SSE route for continuous reading, but add a
   single-page downloader for left-to-right and right-to-left reader modes.
9. Return the real chapter page count before trusting a saved resume position,
   so corrupt progress can never produce unbounded page requests.

## Validation

- Added pure unit coverage for merge-gap fallback, explicit-group filtering,
  unnumbered chapters, coverage counts, and deterministic selection.
- Ran `gofmt`, focused MangaDex/controller tests, and
  `go test ./modules/database` successfully before restart.
- Rebuilt and restarted the backend. Startup logs confirm the
  `user_preferred_translation_language` migration completed successfully and
  the API is listening on port 3001.
- Added database coverage for per-user, per-manga, per-language isolation and
  the All groups reset behavior. `go test ./modules/database ./controllers`
  passes; the rebuilt backend ran the new preference-table migration and is
  listening on port 3001.
- Added all-language ordering coverage. `go test ./modules/mangadex-api
  ./controllers` and the required `go test ./modules/database` pass; the
  rebuilt backend starts successfully on port 3001.
- Added a defaulted, validated `preferred_reading_mode` setting and database
  coverage for its vertical default and per-user isolation. `go test
  ./modules/database ./controllers` passes before restart.
- Added `GET /chapter/:chapterId/page?page=N`, which validates the page number,
  downloads only that one image, and returns its local public URL plus the
  total page count. Focused MangaDex/controller tests and the required
  database test pass before restart.
- Added `GET /chapter/:chapterId/page-count`, a metadata-only endpoint used to
  validate a paged reader's saved position before it requests an image.
  Focused MangaDex/controller tests and the required database test pass.


## Follow-up

- Consider an advanced toggle to display every alternate release rather than
  the deduplicated **All groups** view.
