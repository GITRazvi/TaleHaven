# Public and invite-only chatrooms

## Naming note

Use **Private** as the reader-facing access label. The API continues to send
`invite_only` internally for compatibility with existing rooms.

## Goal

Provide a signed-in chat area for public and invite-only rooms without room
passwords.

## Scope

- Browse and join public rooms.
- Create public or invite-only rooms.
- Enter invite-only rooms through an owner-shared link.
- Read and send messages with periodic refresh.
- Keep invite-only rooms out of public discovery.
- Let the room owner switch existing rooms between public and invite-only.
- Keep the room-creation form collapsed by default behind a clear Create
  chatroom button, so browsing rooms remains the primary view.
- Keep all per-room management controls inside a collapsed creator-only Room
  settings panel.
- Let readers search the rooms they are permitted to see by name, including
  public rooms and their own private rooms, without disclosing other private
  room names.

## API and data changes

- Consume protected `/chat` room, membership, invite, and message routes.
- Consume the owner-only room visibility update route.
- Add routes for `/chat`, room conversation, and invite redemption.

## Implementation steps

1. Add API client types and calls.
2. Add protected routes, room browser, room creation, invite joining, and chat
   messages.
3. Show owner-only invite-link controls and clear authorization/errors.
4. Add focused tests and validate build/runtime behavior.

### Safe room-name search extension

1. Filter the already authorized room-list response by a case-insensitive room
   name query.
2. Apply the query to public rooms and private rooms the signed-in reader can
   access; do not add a discovery path for private rooms outside that list.
3. Keep the public and private sections visible with clear empty-search states
   and cover filtering in the room-browser test.
4. Keep the search field collapsed behind a Search button; closing it clears
   the query so no hidden filter remains active.

### Visibility update extension

1. Add the visibility update API call.
2. Show the owner an access-mode toggle inside the room.
3. Refresh the displayed room access state after saving and test the control.

### Room and message images extension

1. Add optional room-image and message-image API fields and multipart calls.
2. Let room owners upload a room image; show it beside the room title in room
   cards and the conversation header.
3. Let members attach one image to a message, optionally with text, and render
   it safely in the conversation.
4. Replace the visible message file picker with paste-to-attach behavior in
   the message box, while keeping an explicit remove action before sending.

### Keyboard message sending extension

1. Submit a chat message when Enter is pressed in the message box.
2. Preserve Shift+Enter for composing a new line and do not submit while an
   input-method composition is active.
3. Add focused coverage for both keyboard paths.

### Conversation follow extension

1. Open a room at its newest message and keep following the conversation as
   new messages arrive.
2. Detect when the reader scrolls away from the bottom and preserve that
   position during polling updates.
3. Resume following when the reader returns to the latest message and cover
   the behavior with a focused test.

### Message actions extension

1. Show a compact three-dot actions menu only for messages authored by the
   signed-in user.
2. Add a Delete action that removes the message immediately after the protected
   server request succeeds, leaving room for future menu actions.
3. Cover author-only visibility and deletion behavior.

### Replies and reporting extension

1. Put Reply in every message's three-dot menu, keep Delete for the author,
   and add Report for other members' messages.
2. Show the selected parent message above the composer and persist/show a
   compact reply reference when the new message is posted.
3. Collect a required written report reason in a modal before sending the
   protected report request, with clear success and failure feedback.

### Reply navigation extension

1. Give each rendered message a stable DOM anchor based on its message ID.
2. Store the source message ID with a reply reference and make that reference
   clickable so it scrolls the original message into view.
3. Preserve existing reply, deletion, reporting, and auto-scroll behaviour;
   add focused frontend coverage for reply navigation where practical.

## Validation

- Run the full frontend test suite and production build before restart.
- Verify the rebuilt frontend serves the chat route.

## Follow-up work

- Add live socket updates, typing state, pagination/history controls, message
  reports, and room moderation.

### Room moderation extension

1. Show the room creator a member-management panel inside Room settings to
   assign/revoke moderator status using safe names and avatars.
2. Mark the current reader's moderator status in room data so owners and
   moderators can use delete/ban actions for other messages.
3. Confirm permanent bans before sending them and remove the banned member
   from the active conversation and owner member list after success.
4. Cover owner-only moderator assignment plus the moderator action controls.

### Member directory extension

1. Add a Members button in an open chatroom for every current member.
2. Open a compact read-only member modal containing names as public-profile
   links; leave management actions inside creator-only Room settings.
3. Add focused coverage for opening and rendering the directory.

### Per-user room notification muting extension

1. Surface the signed-in member's current mute state in room metadata.
2. Add a header control to mute or unmute the current room without requiring
   creator privileges.
3. Keep the setting scoped to new-message notifications from that room only.

## Implementation status

- Added protected room browser, room creation, invite redemption, conversation,
  posting, and owner invite-link controls.
- The conversation refreshes messages every five seconds in this initial polling
  implementation.
- Added focused room browser, room view, and invite redemption tests.
- Validation passed: 46 frontend tests, production build, and rebuilt frontend
  serving `/chat` successfully.
- Room owners can now use the in-room **Make public** / **Make invite-only**
  toggle, with the displayed access badge updating after the server saves it.
- Chatroom cards and room headers now show an optional cover beside the name.
  Owners can replace that cover, while members can attach one image with an
  optional text caption; attached images open at their full size.
- Validation passed: all 46 frontend tests, production build, and a rebuilt
  frontend serving `/chat` successfully.
- Room creation is now opened on demand from a Create chatroom button and can
  be cancelled without affecting the room browser.
- Room access, invitations, cover image, palette, custom color, and theme
  background image are now hidden behind a Room settings button visible only
  to the room creator.
- Validation passed: 47 frontend tests, production build, and rebuilt
  frontend serving a chatroom route successfully.
- Validation passed: all 46 frontend tests, production build, and rebuilt
  frontend serving `/chat` successfully.
- Message images can now be pasted directly into the message box; the file
  chooser is removed from the normal chat interface.
- Validation passed: 48 frontend tests, production build, and rebuilt
  frontend serving a chatroom route successfully.
- Chat messages now send with Enter; Shift+Enter creates a new line, including
  safe handling for active input-method composition.
- Validation passed: 49 frontend tests and the production build.
- Rooms now open at the newest message and remain there through message/image
  updates until the reader scrolls upward; following resumes when they return
  to the bottom.
- Validation passed: 50 frontend tests and the production build.
- Messages authored by the signed-in user now have a three-dot actions menu
  with Delete. Successful deletion immediately removes that message from the
  conversation; other members' messages expose no action menu.
- Validation passed: 51 frontend tests and the production build.
- Every message now exposes Reply from its three-dot menu. The selected message
  is shown above the composer and sent as a compact quoted reference.
- Other members' message menus also expose Report, which requires a written
  reason before submission; a user’s own messages retain Delete instead.
- Validation passed: all 51 frontend tests and the production build.
- A reply reference now stores the source message ID and is rendered as a
  clickable control. Selecting it smoothly scrolls the original message into
  view; an unavailable source is clearly labelled.
- Validation passed: focused chatroom coverage, all 52 frontend tests, and the
  production build.
- Every room member can now open a compact Members dialog from the conversation
  header. It contains only clickable member names; moderation controls remain
  hidden in creator-only Room settings. Focused chatroom coverage and the full
  frontend suite pass (17 suites, 67 tests); the rebuilt frontend is running
  on port 3000.

### Chat author profile preview

1. Make the visible message author name a hover- and keyboard-focus trigger.
2. Present a compact preview whose profile image fills the card and whose name
   overlays the image; use an initial-based fallback when no image exists.
3. Keep the preview non-interactive and avoid adding any profile details beyond
   the public name and photo.

## Implementation status

- Hovering or keyboard-focusing a message author's name now opens a compact
  profile preview. The portrait fills the preview and the display name overlays
  its top edge; a letter avatar is used when no photo exists.
- Validation passed: focused chatroom coverage, all 53 frontend tests, and the
  production build.
- The room-name search is now collapsed behind a Search button beside Create
  chatroom. Closing the panel clears its query while retaining the same safe
  public/private filtering behavior.

### Room moderation implementation status

- Room settings now includes an owner-only member list with safe profile links
  and Make/Remove moderator controls.
- Owners and moderators receive delete and permanent-ban actions in another
  member's message menu. Bans require explicit confirmation and remove that
  member's messages from the active view after success.
- Full frontend validation passes (17 suites, 67 tests), including moderator
  action coverage. The rebuilt frontend container is running on port 3000.
- Every current chatroom member can now mute or unmute that room from its
  header. The preference is reader-specific and only stops future new-message
  notifications from that room. Focused coverage was added; the full frontend
  suite passes (17 suites, 68 tests) and the rebuilt frontend is running on
  port 3000.
