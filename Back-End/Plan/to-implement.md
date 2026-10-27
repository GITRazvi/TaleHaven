# To Implement

## Goal

Track the next server-side TaleHaven features before their individual
implementation plans are created.

## Backlog

1. **Chat bot**
   - Support an in-app assistant capable of approved actions, including
     settings changes.
   - Define allowed commands, authentication requirements, audit behavior, and
     confirmation safeguards before implementation.

2. **Authentication sessions**
   - Implement proper login/logout sessions using secure cookies.
   - Cover session creation, expiry, renewal policy, logout invalidation,
     authorization middleware, and CSRF protection where applicable.
   - See `authentication-sessions.md` for the approved server/session design.

3. **Profile and overall interface refresh**
   - Provide profile data and update APIs for profile pictures, display names,
     and the settings needed by the new interface theme.

4. **Homepage Continue Reading**
   - Provide the user's started-but-unfinished manga, their most recent
     chapter/page position, and whether new matching releases are available.
   - Define a predictable ordering, such as most recently read first.

5. **Admin tools and chatrooms**
   - Add role-based administrator tools for moderation and site management.
   - Support public chatrooms and password-protected rooms, including secure
     password handling, membership/access checks, and message moderation.

## API and Data Changes

To be decided in dedicated feature plans. Expected work includes session
storage/cookie contracts, authenticated profile endpoints, reading-progress and
release-availability queries, room/message storage, role authorization, and a
strictly scoped chat-bot action interface.

## Implementation Steps

Create a separate, detailed plan for each backlog item before starting its
implementation.

## Validation

Documentation-only backlog; no server behavior changed.

## Follow-up

Design and implement authentication sessions before exposing chat-bot actions,
chatroom access, or administrator capabilities.
