# Notification Center

## Goal

Provide each signed-in reader with a persistent, read/unread notification feed.

## Scope

- Store notification type, title, concise body, destination path, timestamp,
  and read state per user.
- Create immediate notifications for profile comments, chat messages, room
  moderation changes, and bans.
- Expose protected list, unread-count, mark-read, and mark-all-read routes.
- Manga release polling is a follow-up: it requires a rate-limited MangaDex
  scheduler and per-title release deduplication.

## Validation

- A user can only read or mark their own notifications.
- Event recipients exclude the actor where appropriate.
- Existing chat and profile actions continue to succeed when notification
  delivery is unavailable.

## Implementation Status

- Added persistent per-user notification storage, unread indexes, and protected
  list/count/read/read-all routes.
- Profile comments notify the profile owner; chat messages notify other room
  members; moderator assignments, removals, and bans notify the affected user.
- The standalone database test is blocked by this Windows host's forwarded
  PostgreSQL port restriction; controller tests pass before the containerized
  restart. The rebuilt backend applied `user_notifications` and is running on
  port 3001.
