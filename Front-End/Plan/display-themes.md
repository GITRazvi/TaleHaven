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
- When a background image is active, render profile/chat content and form
  surfaces as translucent, blurred layers so the image remains the primary
  visual theme without sacrificing readability.
- Apply each chatroom's themed backdrop to the surrounding document area while
  its page is open, matching the My Profile behavior.
- Let a signed-in user choose a separate global site palette for general pages
  such as Home, manga search, and the chatroom browser.
- Let a signed-in user upload several site-wide background images and select a
  static image, each-refresh image, per-browser-session image, or a
  user-entered minute rotation for browsing pages.
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
7. Make image-backed presentation surfaces translucent and restore the
   document background when navigating away.
8. Keep image-backed surface tint light enough for the uploaded artwork to
   remain clearly visible through message and input areas.
9. Apply the saved global site palette to shared page surfaces while allowing
   My Profile and an open chatroom to retain their scoped overrides.
10. Upload, preview, remove, and rotate saved site-theme images; preserve
    readable translucent browsing surfaces while an image is active.
11. Offer refresh/session/timed rotation modes and validate a reader-entered
    minute interval for timed rotation.
12. Anchor site-theme images to the viewport so a long page does not enlarge
    or crop the background based on its scroll height.
13. Permit GIF files in theme-background upload controls and make the animated
    nature and existing file-size guidance clear to readers.
14. Apply an active profile or open-chatroom background image to the shared
    viewport behind the app shell instead of repeating it inside only the
    centered content container.

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
- Image-backed profile and chatroom content now uses translucent, blurred
  surfaces and inputs; a chatroom also applies its selected palette or image
  to the surrounding document backdrop while open, then restores it on leave.
- Validation passed: 47 frontend tests, production build, and rebuilt
  frontend serving a chatroom route successfully.
- Site, profile, and chatroom theme controls now offer GIF selection alongside
  JPEG, PNG, and WebP and explain that GIFs play as animated backgrounds.
- Validation passed: 50 frontend tests and the production build.
- Added an independent Site theme control in Settings. Its palette applies to
  shared browsing pages below the unchanged navbar, while My Profile and open
  chatroom pages retain their scoped owner/user themes.
- Added site image uploads and thumbnail removal controls. Browsing pages use
  the current image behind translucent theme-aware surfaces.
- Replaced the fixed-second selector with static, per-refresh, per-browser-
  session, and user-entered-minute modes. Session selection survives refreshes
  in the same browser tab; timed rotation uses the reader's minute value.
- Selected site background images now upload as part of Save site theme rather
  than requiring a separate upload action.
- Site-theme images are now viewport-fixed, so scrolling or a long page no
  longer causes `cover` to scale the artwork against the full document height.
- Calibrated image-theme surfaces to a 42% theme tint and inputs to 38%, with
  reduced blur so the background artwork is clearly visible rather than
  covered by an effectively opaque palette.
- Validation passed: 47 frontend tests, production build, and rebuilt
  frontend serving a chatroom route successfully.
- Active My Profile and open-chatroom images now sit on the fixed document
  backdrop. The app shell and centered page container become transparent so
  the same artwork fills the outer side areas as well as the content region.
- Validation passed: 51 frontend tests and the production build.
