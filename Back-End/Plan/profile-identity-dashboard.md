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

## API and Data Changes

- `GET /profile` returns the authenticated safe user profile.
- `PATCH /profile` accepts `{ "displayName": "..." }`.
- `POST /profile/photo` accepts multipart field `photo` and returns the
  updated safe user profile.
- `users` gains `display_name` and `profile_image_path` columns.

## Implementation Steps

1. Add idempotent user-profile migrations and database update operations.
2. Extend authentication user loading/responses with safe profile fields.
3. Add protected profile read, display-name update, and photo upload routes.
4. Validate image size/type, generate server-owned filenames, and avoid raw
   upload filenames or arbitrary file paths.
5. Add database coverage for profile updates and retain user isolation.

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

## Follow-up

- Cropping, avatar presets, and external-image URL import are intentionally
  out of scope for this first upload implementation.
