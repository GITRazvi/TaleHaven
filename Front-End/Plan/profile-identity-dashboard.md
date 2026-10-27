# Profile Identity and Dashboard

## Goal

Turn My Profile into a personal manga dashboard while keeping Settings focused
on editable account and reader preferences.

## Scope

- Add a My Profile header with the user's photo, display name, member date,
  saved-title count, and a link to Settings.
- Add a Currently Reading section sourced from existing `Reading` library
  status entries.
- Retain and visually organize the existing statistics, favorites, and
  filterable full library as dashboard sections.
- Add Settings controls for display name and profile-photo upload alongside
  existing reader and account controls.
- Introduce shared profile/dashboard CSS for a consistent visual hierarchy.

## API and Data Changes

- Consume safe user profile fields from login, registration, and `/auth/me`.
- Use `GET/PATCH /profile` and `POST /profile/photo` through the credentialed
  backend client.
- Reuse the existing protected AniList library, favorite, and statistics APIs.

## Implementation Steps

1. Extend frontend user/API types and add profile client methods.
2. Add identity editing and image preview/upload feedback to Settings.
3. Add the profile dashboard header and Currently Reading section without
   duplicating existing library/favorites requests.
4. Apply shared responsive profile styles and preserve existing controls.
5. Add focused UI coverage, then run the full suite/build and restart the
   frontend.

## Validation

- The profile header updates after a display-name or photo change.
- Currently Reading contains only titles with the Reading status.
- Existing statistics, favorites, library filtering, rating, and status
  controls remain functional.

## Implementation Status

- Settings now lets the user save a display name and upload a JPEG, PNG, or
  WebP profile photo. The auth bootstrap refreshes the safe profile afterward.
- My Profile now has an identity header, member date, saved-title count,
  Settings shortcut, and a Currently Reading section above the existing
  statistics, library, and favorites sections.
- Added shared responsive profile styles and dashboard coverage for the
  identity header/Currently Reading section. The complete frontend suite
  passes (8 suites, 32 tests) and the production build completes with only
  existing Create React App and unused-type warnings.
- The rebuilt frontend container started successfully and returned HTTP 200.

## Follow-up

- Theme selection belongs to the later display-preferences implementation;
  this change establishes reusable profile components/styles only.
