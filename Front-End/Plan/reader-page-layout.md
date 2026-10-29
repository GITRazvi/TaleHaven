# Reader Page Layout

## Goal

Display consecutive pages in continuous vertical reading without visual gaps between their images, on a distraction-free reader background.

## Scope

- Frontend reader layout only.
- The reader uses pure black in dark mode and pure white in light mode, rather than an uploaded or colored site theme.
- No reader API, progress, chapter navigation, or persistence changes.

## Implementation

1. Remove the vertical flex gap from the streaming page container.
2. Ensure dynamically inserted page containers have no bottom margin in vertical mode.
3. Keep paged left-to-right and right-to-left layouts unchanged.
4. Override the surrounding site-theme surface on the reader route so the reader background fills the available page area.

## Validation

- Run the frontend test suite and production build.
- Confirm the running frontend recompiles after the styling change.

## Follow-up

- None currently.
