# User Library Frontend Plan

## Goal

Let signed-in users save manga and revisit their private library.

## Implemented scope

- Added a protected `/library` page and navigation entry.
- Added a library client for listing, adding, and removing saved manga.
- Replaced the details-page text action with an accessible heart toggle.
  - `♡` means the manga is not saved.
  - `♥` in red means the manga is saved.
- Added the same heart toggle to Popular Manga covers. It appears in the
  upper-right corner while a card is hovered and does not open the manga when
  clicked.
- Redirected unsigned users to login when they try to save a manga.
- Added library-card reading controls:
  - **Read from page 1** selects the first available chapter and overrides
    saved progress.
  - **Continue reading** opens the most recently read chapter and lets the
    reader resume from its saved page.
- Added a reader header that displays the manga title and current chapter
  number, including when reading starts from a library card.
- Added completed-chapter state to the details page and placed chapters in a
  dedicated scrollable panel.
- Updated the library reading actions to hand the reader the ordered chapter
  sequence it needs for previous/next chapter navigation.
  - **Read from chapter 1** starts at chapter 1 when available and overrides
    saved page progress.
  - **Continue reading** supplies the same sequence while retaining the saved
    resume page.
- Retrieve that sequence in MangaDex-compatible 100-chapter batches, rather
  than making an invalid 500-chapter request.
- Add **Read from chapter 1** and **Continue reading** controls above the
  chapter list on each manga details page.
- Show each chapter's saved reading position as a smaller `Page X` label in
  the chapter list, matching the page the reader resumes from.
  - Keep the label visibly right-aligned beside its chapter title.
  - Load every chapter in supported 100-item batches, so progress is visible
    even when the saved chapter falls after the first 100 results.
  - Save the current page for the active chapter before navigating away, so
    rapidly switching chapters cannot lose that chapter's separate position.
  - Resolve and restore a chapter's saved position when streaming starts, not
    after every page has arrived; do not overwrite that position while the
    stream is still catching up to it.
  - Ask the chapter stream to begin at the saved page, so Continue Reading
    does not download earlier pages before displaying the saved position.
  - Keep chapter pages in their natural visual order while earlier pages arrive
    later in the progressive stream.
  - Persist the initial or resumed page when it is inserted into the reader,
    rather than waiting for a browser visibility callback.
  - Track the loaded page closest to the viewport center on every scroll and
    as streamed images finish loading.
  - Register streamed-image load handlers before assigning their source, so
    cached pages also enter the progress tracker and can trigger completion.
  - Keep the active viewport anchored while earlier streamed pages are inserted
    and expand, preventing progressive loading from returning the reader to
    the first page.
  - Persist and synchronize each newly visited page through a serialized queue,
    so chapter-list progress updates before the page stream completes and
    cannot be overwritten by an older in-flight request.
  - Add frontend automated tests for home-page library heart toggles and for
    per-page reader progress persistence before a stream completes.
  - Cover failed add and remove requests so the heart remains in its prior
    state and the user receives an error message.
  - Load My Library in 24-title batches and show a **Load more** action only
    when the backend reports further saved titles.
  - Load Popular Manga in 20-title batches using the existing MangaDex offset
    API, appending a requested batch only when the reader selects **Load more**.
    Preserve saved-heart state for newly appended cards.
  - Render the cover URL already returned by Manga search on each result card,
    with a visible fallback when a title has no cover or its image cannot load.
  - Arrange Manga Search results as horizontal cards: a fixed cover column on
    the left and title/description content on the right.
  - Show three horizontal search-result cards per desktop row and clamp long
    descriptions to the available card height with an ellipsis.
- Make **Continue reading** start at chapter 1, page 1 when the manga has no
  saved progress yet.
- Mark a chapter as read only when its saved current page reaches the actual
  total page count reported by the chapter stream.
  - Immediately persist and sync completion when the reader displays the final
    page, so the chapter becomes read without requiring navigation away.
- Add multi-select chapter controls on manga details pages to mark selected
  chapters read or unread.
  - Use a checkbox-free selection mode; chapter color remains the read-state
    indicator (grey for read, white for unread).
  - Marking read does not query MangaDex once per selected chapter, avoiding
    rate limits during large selections.
  - Marking unread clears the selected chapters' saved progress.

## Validation

- Extended the Manga Search layout test to confirm desktop three-column cards.
- Ran the full frontend suite successfully: 4 suites and 8 tests passed.
- Re-ran the production build and restarted the frontend development server
  after adding three-column search results and description clamping.
- Extended the Manga Search cover test to verify its result card uses the
  horizontal layout.
- Ran the full frontend suite successfully: 4 suites and 8 tests passed.
- Re-ran the production build and restarted the frontend development server
  after changing the search-result layout.
- Added a Manga Search result test confirming the API cover URL renders as the
  result card image.
- Ran the full frontend suite successfully: 4 suites and 8 tests passed.
- Re-ran the production build and restarted the frontend development server
  after adding search-result covers.
- Added a Popular Manga **Load more** test that confirms the second request
  uses the next offset and appends the returned manga.
- Ran the full frontend suite successfully: 3 suites and 7 tests passed.
- Re-ran the production build and restarted the frontend development server
  after adding Popular Manga pagination.
- Added a My Library pagination test that confirms the next request is made
  only after **Load more** is clicked.
- Ran the full frontend suite successfully: 3 suites and 6 tests passed.
- Re-ran the production build and restarted the frontend development server
  after adding the Load more flow.
- Added frontend coverage for failed library additions and removals; each test
  confirms the heart state remains unchanged and an error is displayed.
- Ran the full frontend suite successfully: 2 suites and 5 tests passed.
- Re-ran the production build and restarted the frontend development server.
- Ran a production frontend build successfully.
- Re-ran the production build after adding the library reading controls.
- Re-ran the production build after adding the reader header and chapter-state UI.
- Re-ran the production build after fixing library reader navigation and the
  chapter-1 control label.
- Re-ran the production build after batching library chapter retrieval within
  MangaDex's supported request limit.
- Re-ran the production build after adding reader controls to manga details
  pages.
- Re-ran the production build after correcting chapter completion tracking.
- Re-ran the production build after adding multi-select read/unread actions.
- Confirmed the restarted backend returns the actual page count used by the
  multi-select read action.
- Re-ran the production build after replacing checkboxes with selection mode
  and removing per-chapter MangaDex requests from bulk updates.
- Re-ran the production build after adding saved-page labels to chapter rows.
- Re-ran the production build after adding the Continue Reading chapter-1
  fallback.
- Re-ran the production build after surfacing live saved-page progress and
  making Continue Reading available before a manga has been started.
- Confirmed the restarted development server compiled the checkbox-free
  selection flow.
- Confirmed the development server recompiles the current source.
- Restarted the development server and confirmed it compiled successfully.
- Re-ran the production build and restarted the development server after
  making saved-page labels visible across the full chapter list.
- Re-ran the production build and confirmed the development server recompiled
  after persisting each chapter's last-read page on navigation.
- Re-ran the production build and restarted both services after moving resume
  and per-page progress persistence into the active streaming phase.
- Re-ran the production build and restarted both services after requesting the
  progressive stream directly at the saved page.
- Re-ran the production build and restarted both services after preserving
  visual chapter order during prioritized page streaming.
- Re-ran the production build and restarted both services after adding
  immediate final-page completion persistence.
- Re-ran the production build and restarted both services after saving the
  initial streamed page without waiting for a visibility callback.
- Re-ran the production build and restarted both services after adding
  continuous scroll-based tracking for rendered streamed pages.
- Re-ran the production build and restarted both services after ensuring
  cached streamed pages always register as loaded before their source is set.
- Re-ran the production build and restarted both services after preserving the
  reading viewport while prior pages stream into the document.
- Re-ran the production build and restarted both services after making each
  visited page save and synchronize through the serialized progress queue.
- Added frontend automated tests covering add/remove home-page heart actions
  and page-progress saving before a stream completes; all three tests pass.
- Confirmed the restarted backend accepts the chapter request behind both
  library reading controls without returning a 500.
- The project still reports its existing unused `SyncReadingProgressBody` lint warning.

## Follow-up work

- Consider library filtering and sorting, such as recently added or title A–Z.
