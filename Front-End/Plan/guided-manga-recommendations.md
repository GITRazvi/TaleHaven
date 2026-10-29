# Guided manga recommendations

## Goal

Add a compact, preset-question recommendation chat to the home page so readers
can discover AniList manga without writing free-form prompts.

## Scope

- Offer fixed paths for genres, most popular, new releases, top rated, and a
  surprise recommendation.
- Convert each choice into the existing public AniList manga-search request.
- Show a small row of recommended covers and titles within the chat panel.
- When signed in, exclude AniList titles already saved in the reader's library.
- Support Show more and Start over without storing a conversation or sending
  user text to a third party.
- Accept a short typed request and recognize supported genre, format, and
  status keywords, plus “similar to <title>” and “same author as <title>”.
- Let the guided genre buttons select multiple genres before recommendations
  are requested, without requiring the reader to type a query.
- Make **Surprise me** personalize its top-rated AniList search with one of
  the reader's tied most-read genres from completed titles, while retaining a
  general fallback for a new reader.

## API and data changes

- Reuse the existing `GET /anilist/manga/search` contract and authenticated
  AniList-library list endpoint; no new backend route or stored data is needed.
- Extend the existing public AniList search query with an optional staff ID so
  a resolved title's author can be used as a precise recommendation filter.

## Implementation steps

1. Create an isolated home-page recommendation chat component with a finite
   preset-question state machine.
2. Map each choice to a valid AniList filter and request score-sorted results.
3. Filter saved IDs, render compact clickable result cards, and paginate on
   demand.
4. Add focused component tests, run the full frontend suite/build, and restart
   the frontend service.
5. Expose the recommender through the navigation bar in a right-side panel and
   remove the duplicated home-page placement.
6. Parse typed requests into catalog filters; resolve the title in similar and
   same-author requests before issuing the final recommendation search.
7. Make guided genre buttons toggleable and submit all selected genres as the
   existing repeated AniList genre filter.
8. Reuse the authenticated library response's `stats.topGenres` for the
   surprise action; do not create a separate recommendation-profile endpoint.
9. Replace the guided series-type and publication-status paths with direct
   popularity, newest-release, and top-rated discovery choices while leaving
   those filters available in typed requests.

## Validation

- Added focused coverage for preset-to-AniList query mapping and excluding a
  signed-in reader's saved titles.
- Passed the focused recommendation-chat test, all 55 frontend tests, and the
  production build. Existing React test and Create React App warnings remain
  outside this feature's scope.
- After typed requests were added, all 60 frontend tests passed and the
  production build completed. Existing React test warnings and the unrelated
  unused `SyncReadingProgressBody` lint warning remain outside this feature.

## Implementation status

- Added the guided recommendation chat near the top of the home page.
- The chat offers genre, series type, publication status, and Surprise me;
  each route uses a score-sorted existing AniList search and renders direct
  links to manga details.
- Signed-in readers' saved AniList IDs are loaded before choices are enabled,
  preventing those titles from appearing in the initial recommendation set.
- Moved the recommender from the home page to a **Recommendations** navbar
  control. It opens in a right-side panel, so it stays available while readers
  browse any part of the site.
- Rename the reader-facing navbar and panel label to **Chatbot** while keeping
  the guided-recommendation scope explicit within the panel.
- The navbar item and right-side panel are now labelled **Chatbot**. Validation
  passed: all 56 frontend tests and the production build.
- Apply the active site palette and custom-color variables to the Chatbot
  panel, including its controls and cards, despite the panel rendering outside
  the normal themed page subtree.
- Added navbar interaction coverage; all 56 frontend tests and the production
  build pass.
- The Chatbot panel now inherits the active site palette, custom color, and
  uploaded site artwork even though Bootstrap renders it outside the normal
  page-theme subtree. Its recommendation card and controls use the same
  readable translucent treatment when a background image is active.
- Chatbot now accepts a typed request alongside the preset prompts. It supports
  multiple recognized genres, `manga`/`one-shot`, and
  `completed`/`ongoing` combinations. It also resolves `similar to <title>`
  through the title's public genres and top non-spoiler tag, and `same author
  as <title>` through AniList staff IDs; the referenced manga is left out of
  the resulting cards.
- Added parser and component coverage for typed filter, similar-title, and
  same-author flows.
- The guided genre route supports selecting multiple visible genre buttons and
  submits their intersection only when the reader chooses to view results.
- **Surprise me** now selects a top-rated manga from one of the signed-in
  reader's tied most-read completed genres. Readers without completed history
  receive a varied general-genre recommendation instead.
- Guided buttons now provide **Most popular**, **New releases**, and **Top
  rated** alongside genre selection. Their selected AniList sort order is
  preserved rather than being overwritten by the top-rated default.

## Follow-up work

- Similar-title, mood, and personalized genre-history recommendations can be
  expanded later only after their ranking rules are defined.
