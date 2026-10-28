# Public and invite-only chatrooms

## Naming note

The UI calls invite-only rooms **Private**. The persisted/API compatibility
value remains `invite_only`, so existing rooms and invite links continue to
work without migration.

## Goal

Let signed-in TaleHaven users create and use real-time-style chatrooms without
passwords: rooms can be public or accessible only through an owner-generated
invite link.

## Scope

- Signed-in users can create public or invite-only rooms.
- Public rooms are discoverable and can be joined by any signed-in user.
- Invite-only rooms are hidden from discovery and require a secure invite link.
- Members can read and post messages; owners can create additional invite links.
- Room owners can switch a room between public and invite-only access without
  removing its existing members.
- Use short polling for new messages in this first version. WebSockets and
  moderation are deferred.

## API and data changes

- Add room, membership, invite, and message tables.
- Add protected routes under `/chat` for room listing/creation/joining,
  invitations, visibility updates, and paged messages.
- Invite secrets are random browser-safe tokens; only their hash is persisted.

## Implementation steps

1. Add migrations, models, database operations, and member/owner checks.
2. Add controller routes with validation and server-side authorization.
3. Add frontend client types and protected room list, create, invite, and
   messaging views.
4. Add focused access-control and UI tests, then restart both services.

### Visibility update extension

1. Add an owner-only visibility update operation and `PATCH` route.
2. Add a room-owner control to toggle the current access mode.
3. Verify a non-owner cannot change access and existing members retain access.

### Room and message images extension

1. Add optional room-image and message-image paths through additive migrations.
2. Add owner-only room image upload and member-only image message upload routes.
3. Store only server-generated image files under the existing static images
   directory; validate JPEG, PNG, and WebP contents and a 5 MB maximum.
4. Permit a message to contain text, an image, or both, while keeping message
   access member-only.

### Animated theme-background extension

1. Permit detected GIF content for room theme backgrounds only.
2. Keep room covers and message images restricted to their current static
   JPEG, PNG, and WebP formats.

### Message deletion extension

1. Add a protected endpoint that deletes only a message's author-owned row in
   its current chatroom.
2. Remove any attached server-owned message image after the database deletion.
3. Verify that a different room member cannot delete another user's message.

### Replies and message reporting extension

1. Persist an optional same-room parent-message reference so replies remain
   visible as contextual message links, and clear that reference if the parent
   is deleted.
2. Add a protected report endpoint that requires a reader-written reason,
   accepts reports only from room members about another member's message, and
   prevents duplicate reports by the same reader.
3. Keep report records available for future administrator moderation tools;
   they are not exposed to ordinary room members.

## Validation

- Run the required database test before backend restart.
- Run focused Go tests, the frontend test suite, and a frontend build.
- Confirm both containers start and protected routes reject anonymous access.

## Follow-up work

- Add message reports, moderation/audit actions, room ownership transfer,
  deletion, rate limits, and WebSocket delivery.

## Implementation status

- Added public and invite-only rooms, membership records, SHA-256 hashed invite
  tokens, and member-only messages through an additive database migration.
- Public rooms are listed to signed-in users but require an explicit join before
  messages are visible; invite-only rooms are listed only to their members.
- Only the owner can create a reusable invite link. Passwords are not stored or
  accepted anywhere in the room flow.
- Validation passed: database tests (including chat privacy/membership), focused
  controller tests, and a rebuilt backend where anonymous `/chat/rooms` access
  returns `401`.
- Added an owner-only visibility update route. Tests verify that non-owners are
  rejected, existing members keep access after a public-to-invite-only change,
  and discovery resumes after switching back to public.
- Added optional room-cover and message-image paths, owner-only room-cover
  uploads, and member-only image-message uploads. Uploads accept detected
  JPEG/PNG/WebP content up to 5 MB and use generated static-file names.
- Validation passed: database and controller tests; the rebuilt backend applied
  the additive chat image migration and started successfully.
- Room theme-background uploads now accept detected GIF content and keep the
  `.gif` extension. Room covers and message images remain limited to static
  JPEG, PNG, and WebP files; controller coverage verifies that split.
- Added an author-only message-deletion operation and protected route. It
  deletes a stored message only when the requesting user wrote it and is still
  a room member, then removes any attached server-owned message image.
- Database coverage verifies both successful self-deletion and rejection when
  another room member attempts the same action.

### Reply navigation extension

1. Persist an optional same-room parent message ID for new replies.
2. Return the parent ID with member-visible messages and reject reply targets
   that do not exist in the current chatroom.
3. Keep replies when a parent is removed, allowing the UI to report that the
   original message is unavailable.

## Implementation status

- Replies now persist a same-room source message ID rather than embedding a
  text quote in the new message body. Invalid and cross-room reply targets are
  rejected before a message is created.
- Validation passed: database tests (including reply-target validation) and
  chat controller tests.

### Chat author profile preview

1. Include only the public display name and profile-image path for each
   message author in the member-visible chat-message response.
2. Preserve author identity privacy by not exposing email, themes, settings,
   sessions, or other profile fields.
3. Cover the new response field with existing database/controller validation.

## Implementation status

- Chat-message responses now include each author's display name and
  profile-image path, sourced from the existing user record. No email or other
  private profile data is included.
- Validation passed: database and chat controller tests.
- Added private, one-per-reporter message reports with a required 1–500
  character reason. Only a room member can report another member's message;
  duplicate reports and self-reports are rejected.
- Validation passed: database tests (including reporting), focused chat
  controller tests, and the backend rebuild/restart check.
