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

## Validation

- Run the full frontend test suite and production build before restart.
- Verify the rebuilt frontend serves the chat route.

## Follow-up work

- Add live socket updates, typing state, pagination/history controls, message
  reports, and room moderation.

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
