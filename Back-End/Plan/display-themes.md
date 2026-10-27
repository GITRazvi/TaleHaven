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
- Optionally persist one validated custom base color (`#RRGGBB`) per user and
  room, with no arbitrary styles or CSS accepted.
- Let a user upload one profile theme background and a room owner upload one
  chatroom theme background, using only JPEG, PNG, or WebP files up to 5 MB.
- Permit only the authenticated user to update their profile theme and only a
  chatroom owner to update that room's theme.

## API and data changes

- Add idempotent preset-theme and custom-color columns to users and chatrooms.
- Include the user theme in safe auth/profile responses and the room theme in
  chatroom responses.
- Add `PATCH /profile/theme` and `PATCH /chat/rooms/:roomID/theme` routes.
- Add protected multipart upload routes for profile and owner-managed room
  theme backgrounds; store only generated local paths.

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
