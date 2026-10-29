# Notification Center

## Goal

Add a navbar bell with unread count and a compact reader-owned notification
list linking to the relevant profile or chatroom.

## Scope

- Fetch protected notifications for signed-in users and poll the unread count.
- Open notifications in an offcanvas, mark entries read on navigation, and
  support mark-all-read.
- Render event titles, messages, and timestamps without exposing sender email.
- Manga-release presentation will use the same center once the backend
  scheduler creates release notifications.
- Keep notification entries translucent at rest, matching their hover state,
  so themed backgrounds remain visible instead of showing a white card.

## Validation

- The bell is absent for signed-out visitors.
- The unread count updates after reading notifications.
- Links use only server-provided internal paths.

## Implementation Status

- Added a signed-in navbar bell with a polled unread badge and an offcanvas
  list. Opening an entry marks it read and navigates to its server-provided
  internal destination; readers can mark every entry read at once.
- The complete frontend suite passes (17 suites, 67 tests) before rebuild.
- The rebuilt frontend is running on port 3000.
- Notification entries now use the translucent hover layer at rest as well,
  avoiding opaque white cards on themed surfaces.
