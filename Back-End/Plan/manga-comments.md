# Manga Comments

## Goal

Let signed-in readers discuss an AniList manga on its description page while
keeping commenter identity limited to safe, viewable profile data.

## Scope

- Store short text comments against an AniList manga ID and authenticated user.
- List comments newest first and include only author ID, display name, profile
  image path, body, and creation time.
- Allow a user to delete only their own comments.
- Validate manga IDs and comment length; no image uploads, replies, reactions,
  or moderation workflow are included in this first version.

## API and Data Changes

- `GET /anilist/manga/:anilistId/comments` lists comments for a signed-in user.
- `POST /anilist/manga/:anilistId/comments` creates a comment with a text body.
- `DELETE /anilist/manga/:anilistId/comments/:commentId` deletes the caller's
  own comment.
- Add an indexed `anilist_manga_comments` table with user, AniList manga ID,
  body, and timestamps.

## Implementation Steps

1. Add an idempotent migration, comment model, and scoped database operations.
2. Add protected routes with request validation and ownership enforcement.
3. Return safe author identity data only, matching the public-profile route.
4. Add controller coverage for invalid inputs and safe response formatting.

## Validation

- Users can create/list comments only while authenticated.
- A user cannot delete someone else's comment.
- Comment responses never expose email, role, credentials, or session data.

## Implementation Status

- Added the idempotent `anilist_manga_comments` migration with an index for
  newest-first per-title retrieval.
- Added protected list/create/delete routes. Bodies are trimmed and limited to
  1,000 Unicode characters; delete queries scope by comment, title, and user.
- Comment responses carry only safe author identity fields used by the public
  profile route. Controller tests pass, and the rebuilt backend successfully
  applied the migration and started on port 3001.
- The required standalone database suite was invoked before restart but cannot
  connect to forwarded PostgreSQL port 5433 from this Windows host. The
  containerized backend connected to PostgreSQL and completed all migrations.

## Follow-up

- Reports, moderator controls, reactions, and comment replies can be designed
  separately once the basic discussion flow is in use.
