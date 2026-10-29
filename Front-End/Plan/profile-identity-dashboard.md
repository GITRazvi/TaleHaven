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
- Let a chatroom member's name open a read-only public profile page.
- Let profile owners control whether their public profile accepts comments,
  and render a signed-in discussion section when comments are enabled.

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
6. Add a protected public-profile route that fetches and renders only a
   member's safe identity fields.
7. Add profile-comment settings, list/post/delete client calls, and safe
   author links in the public-profile discussion section.
8. Surface the same profile discussion section on the signed-in My Profile
   dashboard, including a clear disabled state and the owner deletion action.
9. Keep displayed profile comments translucent when an uploaded profile theme
   image is active, matching the comment input surface.
10. Match chat composer keyboard behavior in profile comments: Enter posts;
    Shift+Enter and Ctrl+Enter insert a newline.

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
- Chatroom message-author names and the room creator name now link to a
  protected, read-only `/profile/:userID` page. The existing author hover
  preview remains available.
- The public profile view displays only the member's display name, photo, and
  joined date, with a safe fallback avatar. Focused UI coverage verifies the
  route and client request; the full frontend suite passes (17 suites, 65
  tests), and the rebuilt frontend container is running on port 3000.
- Settings now includes an owner-controlled Profile comments switch. Enabled
  public profiles show a newest-first discussion section, signed-in posting,
  and deletion for the author or profile owner; disabled profiles show an
  explicit unavailable state. Every comment author uses the same safe,
  clickable public-profile identity.
- The complete frontend suite passes (17 suites, 67 tests), including profile
  comment rendering and author-profile links. The rebuilt frontend container
  is running on port 3000.
- My Profile now also renders the profile discussion section directly below
  its statistics. It shows the comment setting's disabled state, gives the
  owner a shortcut to Settings, and supports posting or deleting comments;
  dashboard coverage verifies visible comments and linked author identities.
- With an uploaded profile theme image, displayed comment cards now share the
  comment input's translucent, blurred surface so the image stays visible.
- Profile comment composers on both the public page and My Profile now post
  with Enter; Shift+Enter and Ctrl+Enter retain multiline input. Focused tests
  cover both forms.

## Follow-up

- Theme selection belongs to the later display-preferences implementation;
  this change establishes reusable profile components/styles only.
