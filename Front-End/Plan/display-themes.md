# Profile and Chatroom Display Themes

## Goal

Provide a personal, non-black-or-white visual treatment for My Profile and
owner-selected presentation themes for chatrooms.

## Scope

- Render five fixed palette choices with preview swatches in Settings.
- Apply the saved personal palette only to My Profile.
- Extend the selected profile palette to the full page backdrop around the
  centered profile layout.
- Let chatroom owners choose a room palette; apply it to the room list card
  and conversation surfaces for all room members.
- Add native color pickers for a precise custom profile or chatroom base color;
  derive the remaining surface and text colors automatically.
- Add an optional uploaded background image for My Profile and each owner-run
  chatroom, rendered beneath a readable contrast overlay.
- Keep the existing site dark/light toggle independent of these scoped themes.

## API changes

- Consume a `displayTheme` field for the signed-in user.
- Consume a `theme` field for chatrooms.
- Consume optional custom base-color fields for profiles and chatrooms.
- Call protected profile and room-theme update endpoints.

## Implementation steps

1. Add client theme types and update calls.
2. Add a Settings palette picker and save feedback.
3. Add an owner-only palette picker in a chatroom.
4. Define accessible scoped CSS variables for the five themes.
5. Add focused tests, run the full test suite/build, and verify the restarted
   frontend.
6. Offer replaceable image uploads with format/size guidance and apply the
   saved background only to its scoped profile or chatroom presentation.

## Follow-up work

- Add personal background-image uploads with positioning and contrast overlay
  controls.

## Implementation status

- Added a saved palette selector with preview in Settings and scoped the saved
  palette to My Profile.
- Added an owner-only chatroom palette selector and themed room list cards and
  conversation surfaces for every room member.
- Validation passed: all 46 frontend tests, production build, and rebuilt
  frontend startup.
- The active profile palette now also sets the full document backdrop while My
  Profile is open, so the surrounding side areas match the selected theme; the
  previous site background is restored when leaving the page.
- Added native custom-color pickers to the profile and owner chatroom theme
  controls. A selected base color derives readable scoped surface, border,
  muted-text, and accent values automatically while presets remain available.
- Re-ran all 46 frontend tests and the production build before restarting the
  frontend with the custom-color controls.
- Added image upload controls to Settings and the owner-only chatroom theme
  card. Saved images appear behind a fixed contrast overlay on My Profile,
  chatroom cards, and the chatroom page.
- Re-ran all 46 frontend tests and the production build, then restarted the
  frontend with the image-theme controls.
- Both theme image controls include a remove action, returning the scoped
  presentation to its selected color palette.
