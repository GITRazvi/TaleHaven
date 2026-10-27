# AniList Related Manga

## Goal

Return AniList relationship metadata with manga details so readers can discover
connected manga without treating chapter or volume counts as eligibility rules.

## Scope

- Request AniList media relations with the existing public manga-details query.
- Return relation type, related media ID, media type, title, cover, format,
  status, and AniList URL through the existing details endpoint.
- Keep upstream anime entries in the typed response so the frontend can make a
  presentation decision, but do not require any chapter or volume data.

## Validation

- Added GraphQL request/decoding coverage for relation metadata.
- Passed `go test ./modules/anilist-api ./controllers` and the required
  `go test ./modules/database` before restart.
- Rebuilt and restarted the backend successfully. A local request for AniList
  manga `105398` returned five relations, including one manga relation.

