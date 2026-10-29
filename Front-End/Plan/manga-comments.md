# Manga Comments

## Goal

Add a focused discussion section to AniList manga description pages, with each
comment author linking to the existing read-only public profile.

## Scope

- Load and render comments beneath the description-page content.
- Let signed-in users submit a short text comment and remove their own entry.
- Make each author name and avatar link to `/profile/:userID`.
- Keep the UI limited to basic text comments; no replies, images, or reactions.

## API and Data Changes

- Add credentialed backend-client methods and types for list/create/delete
  comment requests.

## Implementation Steps

1. Add comment types and backend-client methods.
2. Build the comment form, loading/error/empty states, and comment list.
3. Link author identity to the safe public-profile route and expose delete only
   for the active user's own comments.
4. Add focused UI coverage and run the complete frontend test suite.
5. Match chat composer keyboard behavior: Enter posts, while Shift+Enter and
   Ctrl+Enter insert a newline.

## Validation

- New comments appear without a full page refresh.
- Profile links point at the correct author ID.
- Only the current user's comments show a delete control.

## Implementation Status

- Added a signed-in Comments section to AniList manga details pages, including
  loading, error, and empty states, a 1,000-character form, and own-comment
  deletion.
- Each commenter has a profile avatar and name link to `/profile/:userID`.
- Focused coverage verifies author links, ownership-only delete controls, and
  posting without a page reload. The full frontend suite passes (17 suites,
  66 tests), and the rebuilt frontend container is running on port 3000.
- The manga discussion composer follows chat keyboard behavior: Enter posts,
  while Shift+Enter and Ctrl+Enter insert a newline. Focused coverage verifies
  the Enter submission path.

## Follow-up

- Pagination, replies, reports, and reactions remain intentionally out of
  scope for the first release.
