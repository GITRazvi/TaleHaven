# Profile and Chatroom Display Themes

## Goal

Let readers personalize their My Profile presentation and let chatroom owners
give each room a distinct, readable visual identity without accepting arbitrary
CSS or unsafe color values.

## Scope

- Provide five fixed themes: `midnight`, `forest`, `plum`, `ember`, and
  `parchment`.
- Persist one personal profile display theme per user, defaulting to
  `midnight`.
- Persist one chatroom display theme per room, defaulting to `midnight`.
- Persist a separate user-wide site theme and optional custom base color for
  general pages, defaulting to `midnight`.
- Let each user keep a bounded collection of validated site-theme background
  images and choose static, each-refresh, per-browser-session, or a
  user-entered minute interval rotation.
- Optionally persist one validated custom base color (`#RRGGBB`) per user and
  room, with no arbitrary styles or CSS accepted.
- Let a user upload one profile theme background and a room owner upload one
  chatroom theme background, using JPEG, PNG, WebP, or animated GIF files up
  to 5 MB.
- Permit only the authenticated user to update their profile theme and only a
  chatroom owner to update that room's theme.

## API and data changes

- Add idempotent preset-theme and custom-color columns to users and chatrooms.
- Include the user theme in safe auth/profile responses and the room theme in
  chatroom responses.
- Add `PATCH /profile/theme` and `PATCH /chat/rooms/:roomID/theme` routes.
- Add protected multipart upload routes for profile and owner-managed room
  theme backgrounds; store only generated local paths.
- Add a user-scoped site-theme update route and include its safe fields in
  authenticated user responses.
- Add authenticated upload/removal routes for site-theme images; only generated
  local paths may be stored, and user-owned images may be removed individually.

## Implementation steps

1. Add database defaults, models, query projections, update operations, and
   per-user/per-owner authorization tests.
2. Validate requested theme keys against the fixed server-side allowlist and
   custom colors as normalized six-digit hexadecimal values.
3. Update frontend client types and settings/chatroom controls.
4. Apply scoped CSS theme variables; do not alter the user’s global app theme.
5. Run required tests, rebuild, restart, and verify both services.
6. Validate and replace theme-background uploads without accepting arbitrary
   paths or content types.
7. Persist and return a user-wide site palette without changing the scoped
   profile or chatroom presentation controls.
8. Persist site-theme image paths and rotation settings, with strict image
   validation, per-user ownership checks, and a bounded image count.
9. Replace fixed-second rotation presets with a validated rotation mode and a
   user-entered minute interval for timer-based rotation.
10. Accept detected GIF files only for theme-background upload routes while
    leaving avatar, room-cover, and message-image rules unchanged.

## Follow-up work

- Offer a background-image setting with crop/overlay controls after the fixed
  palette system is established.
- Consider separate light/dark variants and reduced-motion choices.

## Implementation status

- Added idempotent user and chatroom theme columns, both defaulting to
  `midnight`, plus safe response fields and owner/user-scoped update routes.
- The backend allowlist accepts only the five documented theme keys; arbitrary
  CSS and color values are never stored or returned.
- Database, auth, and controller tests passed before the rebuilt backend was
  restarted successfully.
- Added server-side normalization and persistence for optional custom
  `#RRGGBB` base colors on profiles and chatrooms; malformed values are
  rejected before they reach the database.
- Re-ran database, auth, and controller tests, then rebuilt and restarted the
  backend successfully.
- Added owner-scoped/profile-scoped JPEG, PNG, and WebP theme-background
  uploads (5 MB maximum). The database stores generated local paths only;
  replacements remove the previous stored background image.
- The rebuilt backend applied the extended user and chatroom migration and
  started successfully after the image-theme work.
- Added owner/user-scoped removal endpoints so a saved theme image can be
  cleared without affecting the selected preset or custom color.
- Added a distinct persisted site palette and custom color for each user,
  returned through every safe authenticated-user response and updated through
  `PATCH /profile/site-theme`.
- Added a bounded per-user site-theme image collection (up to 10 images) with
  authenticated upload/removal endpoints and a persisted rotation setting.
- Replaced the fixed-second rotation setting with persisted `static`,
  `refresh`, `session`, or `interval` modes; interval values are validated as
  1–1440 minutes.
- Site-theme image uploads are now submitted together with the primary Save
  site theme action, preventing a selected image from being silently omitted.
- Profile and site-theme uploads now accept detected GIF content in addition to
  JPEG, PNG, and WebP; generated `.gif` paths are stored under the same
  server-owned directories and retain the 5 MB limit.
