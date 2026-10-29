# Profile Identity and Dashboard

## Goal

Give each authenticated TaleHaven user an editable identity and provide the
data needed for a polished My Profile dashboard.

## Scope

- Add `display_name` and `profile_image_path` to users.
- Default a new user's display name to the local part of their email address.
- Provide authenticated read/update profile endpoints and a validated image
  upload endpoint for JPEG, PNG, and WebP profile photos up to 5 MB.
- Store profile images below the existing local `images/` static directory and
  delete a replaced profile image only after the database update succeeds.
- Include the safe display name, profile image path, and role in login,
  registration, and `/auth/me` user responses.
- Reuse the existing protected library statistics, favorites, and reading
  status data; no duplicate dashboard tables are needed.
- Expose a signed-in public-profile route that returns only display identity
  fields needed when a chatroom member is selected: ID, display name, profile
  image path, and joined date.
- Allow a profile owner to turn profile comments on or off. When enabled,
  signed-in readers can leave short comments using their safe public identity.

## API and Data Changes

- `GET /profile` returns the authenticated safe user profile.
- `PATCH /profile` accepts `{ "displayName": "..." }`.
- `POST /profile/photo` accepts multipart field `photo` and returns the
  updated safe user profile.
- `users` gains `display_name` and `profile_image_path` columns.
- `GET /profiles/:userID` returns a safe public profile without email, role,
  library, favorites, or presentation/settings data.
- `GET/POST/DELETE /profiles/:userID/comments` lists, creates, and removes
  profile comments. Authors may delete their own entries; profile owners may
  remove any entry on their own page.

## Implementation Steps

1. Add idempotent user-profile migrations and database update operations.
2. Extend authentication user loading/responses with safe profile fields.
3. Add protected profile read, display-name update, and photo upload routes.
4. Validate image size/type, generate server-owned filenames, and avoid raw
   upload filenames or arbitrary file paths.
5. Add database coverage for profile updates and retain user isolation.
6. Validate public-profile IDs and keep the public response explicitly limited
   to non-sensitive identity fields.
7. Add profile-comment storage, the owner-only enable setting, and scoped
   deletion authorization.

## Validation

- A user can update only their own display name/profile photo.
- Invalid image types and files over the limit are rejected.
- Safe user responses never contain passwords or session credentials.
- Required database tests pass before backend restart.

## Implementation Status

- Added idempotent `display_name` and `profile_image_path` migrations. New
  accounts default their display name from the email local part; existing
  accounts receive the same fallback during migration.
- Added protected `GET/PATCH /profile` and `POST /profile/photo` routes.
  Profile photos are limited to 5 MB and validated as JPEG, PNG, or WebP
  before storage under `images/profiles/` with a server-generated filename.
- Login, registration, and `/auth/me` now return safe display-name and photo
  path fields. A `user`/`admin` role remains present for later admin work.
- Profile/auth, controller, session/middleware, and required database tests
  pass, including per-user profile update isolation.
- Account deletion also removes the user's server-owned profile photo after
  deletion succeeds, avoiding retained uploads for deleted accounts.
- After the final rebuild, a disposable live flow verified registration,
  display-name update, image upload/retrieval, account deletion, and image
  removal. The backend started successfully with the profile migration.
- Added the protected `GET /profiles/:userID` route for read-only member
  identity. It validates the ID and returns only ID, display name, profile
  image path, and joined date; controller coverage confirms account fields
  such as email, role, and password never appear in this response.
- Controller tests pass and the rebuilt backend is running on port 3001. The
  standalone database package test was invoked before restart but is blocked
  in this Windows environment because the forwarded PostgreSQL port 5433
  denies the host test process; the containerized backend connected and
  completed migrations successfully.
- Added `profile_comments_enabled` (defaulting to enabled) and a
  `profile_comments` table with indexed, newest-first profile discussion
  retrieval. Profile owners can toggle the setting; signed-in users can post
  up to 1,000 characters while enabled, and only the author or profile owner
  may delete a comment. Comment reads expose only safe author identity.
- Controller tests pass for the profile controller. The required standalone
  database test is recorded above as blocked by the host PostgreSQL port
  restriction. The rebuilt backend connected to PostgreSQL, applied the
  `profile_comments` migration, and is running on port 3001.

## Follow-up

- Cropping, avatar presets, and external-image URL import are intentionally
  out of scope for this first upload implementation.
