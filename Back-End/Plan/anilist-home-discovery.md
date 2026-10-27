# AniList Home Discovery Sections

## Goal

Use AniList's public manga metadata lists to power multiple discovery sections
on TaleHaven's home page, while reserving MangaDex for readable chapters.

## Scope

- Add typed AniList list queries for trending, all-time popular, top-rated, and
  new-release manga.
- Add a public TaleHaven endpoint for one named discovery section at a time.
- Add an on-demand public AniList details endpoint for a clicked title.
- Return AniList pagination metadata so each section can request another page
  later without fetching individual manga records.
- Do not match titles to MangaDex, alter reading routes, or change library
  storage in this step.

## API and data changes

- Add `GET /anilist/manga/:section?page=&perPage=`.
- Add `GET /anilist/manga/details/:anilistId`.
- Supported sections: `trending`, `popular`, `top-rated`, and `recent`.
- Responses contain public AniList metadata plus `total`, `page`, `perPage`,
  and `hasNextPage`.

## Implementation steps

1. Extend the AniList client with a reusable paged-list query and a validated
   section-to-sort mapping.
2. Register a controller that exposes the four discovery sections.
3. Add focused unit tests for client request variables and section validation.
4. Run backend tests, rebuild, and restart the backend.

## Validation

- Formatted AniList client and controller sources with `gofmt`.
- Passed `go test ./modules/anilist-api ./controllers`, including fixed sort
  selection, local GraphQL request variables, GraphQL error handling, and
  invalid discovery-route inputs.
- Passed the required `go test ./modules/database` suite.
- Rebuilt the backend from `main.go` and restarted it successfully on port
  3001.
- Confirmed `GET /anilist/manga/trending?page=1&perPage=2` returns HTTP 200
  with two live AniList records and pagination metadata.

## Remaining follow-up work

- Add AniList metadata details routes and pages.
- Add conservative AniList-to-MangaDex matching before displaying reader or
  library controls on AniList titles.
