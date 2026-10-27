# Translation language and scanlation-group chapter filtering

## Goal

Give readers a saved language preference and clear chapter-list controls for
translation language and scanlation group, while keeping the default list free
of duplicate releases.

## Scope

- Add a Reading section to Settings with an **All languages** default and a
  selectable preferred MangaDex translation language.
- Use the saved preference as the initial language selection when opening a
  MangaDex chapter list; allow the reader to change it without changing their
  global Settings value.
- Add a scanlation-group dropdown below/next to the language selector.
  **All groups** is the default and displays the backend's merged,
  deduplicated chapter list.
- Display each selectable group as `Group name — N chapters` using distinct
  chapter totals supplied by the backend.
- When the language is **All languages**, keep chapter rows globally ordered
  by chapter number, show each row's language label, and keep the group
  control at **All groups** until a reader chooses one language.
- Preserve selected language/group while opening a chapter, changing chapters,
  and using next/previous navigation, so the reader does not silently switch
  translations.
- Restore a signed-in reader's saved scanlation group when they return to the
  same manga and language, and save a new choice when they change the group.
- Add a saved reader layout preference: continuous vertical, left-to-right
  paged, or right-to-left paged.

## API and client data

- Add typed Settings read/update client methods for the reader's preferred
  language.
- Extend chapter-list types with effective filters, available languages, group
  options (ID/name/distinct count), and scanlation-group ID on each chapter.
- Send explicit language and group query values on chapter-list requests;
  migrate existing `translatedLanguage[]` callers after the backend supports
  the new contract.
- Read and save the reader-mode field through the existing Settings client.

## UI behavior

- Settings labels should use human-readable language names while sending the
  backend's language code.
- The manga-details chapter toolbar shows loading and empty states separately
  from the list, so changing a filter does not look like a missing manga.
- **All groups** means one chosen release per chapter/language; selecting a
  named group means its exact release list. The reader treats the first
  missing numbered chapter as the end of that group's readable run and
  disables Next Chapter there.
- Retain existing read/unread color and last-read-page metadata after filtering.
  Progress remains attached to the individual MangaDex chapter ID.
- Keep **All groups** as the explicit saved fallback if a previous group is no
  longer available for that manga/language.
- In paged layouts, tapping the direction of travel advances one page (right
  half for left-to-right, left half for right-to-left); the opposite half goes
  back. In vertical layout, the lower half advances and the upper half goes to
  the previous loaded page. Provide matching arrow controls and keyboard
  arrows without intercepting buttons or links.
- Paged layouts must keep one clear navigation bar and render an explicit
  in-place loading state for the selected page while its image is still
  streaming. A slow or failed active image must never look like an empty
  reader.
- Paged layouts request exactly the selected page through a single-page API.
  They do not begin the continuous reader's chapter-wide SSE stream or fetch
  adjacent pages in the background.
- Paged readers offer a page dropdown centered between Previous Page and Next
  Page. Selecting an option immediately requests that exact image.
- A completed chapter opens on page 1. Any saved position outside the actual
  page range is clamped to page 1 before its image request is made.
- Chapter rows show a last-read-page tracker only for unfinished chapters. A
  final-page position is represented solely by the existing read styling.

## Implementation steps

1. Add client types/methods and load the saved preference in Settings and the
   chapter-details page.
2. Build the Settings selector, save/error feedback, and optimistic refresh of
   the displayed preference.
3. Replace the fixed English chapter request with filter-aware state and add
   language/group selectors to the chapter toolbar.
4. Update chapter rendering/navigation state to use the backend's effective
   source selection, including chapter reader route state.
5. Add language-section rendering for **All languages** and accessible labels
   for group coverage counts.
6. Add the reader-mode selector to Settings, then apply it in the reader while
   retaining progressive page loading and per-page progress persistence.
7. Refine paged-mode loading feedback and remove duplicated navigation so the
   active page remains understandable during progressive loading.
8. Route paged modes through the single-page endpoint, preserving the stream
   only for continuous vertical reading.
9. Fetch page-count metadata for paged resume validation and add direct page
   selection to the paged navigation controls.
10. Present direct page selection as an in-between dropdown rather than a
    separate numeric field and submit button.
11. Hide final-page progress from the chapter list while retaining its read
    completion state.

## Validation

- Ran `npm test -- --watchAll=false --runInBand` successfully (6 suites,
  15 tests). Existing React `act`, open-handle, and Create React App Babel
  warnings remain outside this feature's scope.
- The backend's pure chapter-selection tests cover merged deduplication; the
  frontend renders the response without performing a second deduplication.
- Rebuilt and restarted the frontend. The production bundle compiled and the
  application returned HTTP 200 on port 3000. Existing unrelated lint and
  Create React App maintenance warnings remain documented above.
- Added coverage for restoring/saving the per-manga group selection and for
  disabling Next Chapter at a numbered group gap. The complete frontend suite
  passes (7 suites, 26 tests); both Docker services were rebuilt and
  restarted, and the frontend serves HTTP 200. The production build retains
  its unrelated existing unused-type warning.
- Added the reader-layout selector and applied saved vertical, left-to-right,
  and right-to-left modes without interrupting streamed page loading. Paged
  navigation is tested for right-to-left left-side taps; the full frontend
  suite passes (7 suites, 27 tests), and the production build compiles with
  only the existing unused backend-client type and Create React App warnings.
- Changed left-to-right/right-to-left reading to fetch only the selected page.
  No SSE stream or adjacent-page preload runs in paged layouts; the reader test
  verifies RTL navigation requests page 2 only after the reader taps page 1.
  The full suite passes (7 suites, 27 tests) and the production build compiles
  with the same unrelated warnings.
- Added a direct page-number field and **Go** control to paged reading. Before
  resuming, it loads metadata only, sends completed or invalid saved positions
  to page 1, and then requests just that image. Reader coverage verifies the
  completed-page fallback, RTL navigation, and direct page selection; all 27
  frontend tests and the production build pass.
- Replaced the numeric field and Go button with an immediate page dropdown
  centered between Previous Page and Next Page. The focused reader test, full
  suite (7 suites, 27 tests), and production build pass with only existing
  project warnings.
- Completed chapters now retain their read styling without displaying a
  final-page "Last read" tracker. Focused detail-page coverage and the full
  frontend suite pass (7 suites, 28 tests); the production build completes
  with the existing unused backend-client type and Create React App warnings.

## Follow-up

- Offer a reader-only advanced mode for showing all alternate releases.
