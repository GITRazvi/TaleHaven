# To Implement

## Goal

Track the next user-facing TaleHaven features before their individual
implementation plans are created.

## Backlog

1. **Chat bot**
   - Provide an in-app assistant for actions such as changing settings.
   - Define its supported actions, confirmation rules, and access boundaries
     before implementation.

2. **Authentication sessions**
   - Replace the current login flow with proper login/logout sessions backed by
     secure cookies.
   - Include session expiry, logout invalidation, protected routes, and clear
     unauthenticated states.
   - See `authentication-sessions.md` for the approved browser/session design.

3. **Profile and overall interface refresh**
   - Build out a proper My Profile experience and establish the overall visual
     interface theme.
   - Include profile picture, display name, and the related editable account
     presentation settings.

4. **Homepage Continue Reading**
   - Add a main-page section for manga the user has started but not finished.
   - Each entry should resume the most recently read unfinished chapter at its
     saved page, and distinguish series with newly available chapters.

5. **Admin tools and chatrooms**
   - Add an administrator area for moderation and management tools.
   - Add chatrooms that can be open to everyone or protected by a room
     password, with clear join and access states.

## API and Data Changes

To be decided in dedicated feature plans. Authentication will require session
and profile data contracts; Continue Reading needs a recent-progress/new
chapter contract; chatrooms need room, membership, and message contracts. The
chat bot will require an explicitly limited action API.

## Implementation Steps

Create a separate, detailed plan for each backlog item before starting its
implementation.

## Validation

Documentation-only backlog; no application behavior changed.

## Follow-up

Prioritize the authentication/session design before allowing any chat-bot
setting changes, chatroom access, or administrator capabilities.
