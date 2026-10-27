# AniList Character Pagination

## Goal

Keep manga details lightweight by returning the first 10 character portraits,
then expose a paged API for loading further characters on demand.

## Scope

- Change the details query's initial character connection to 10 entries.
- Add a public, validated endpoint for a specific manga character page.
- Return AniList page metadata with every character-page response.
- Keep chapter-list behavior unchanged: chapters remain fully loaded in their
  existing scrollable panel.

## API changes

- Add `GET /anilist/manga/characters/:anilistId?page=<n>&perPage=<n>`.
- Limit the endpoint to pages of 1–25 character entries.

## Validation

- Added client coverage for the character-page GraphQL query and pagination
  variables. Focused AniList/controller tests and the required database suite
  passed.
- Rebuilt and restarted the backend. A local request confirmed details return
  10 characters and the page-two endpoint returns the next 10.

