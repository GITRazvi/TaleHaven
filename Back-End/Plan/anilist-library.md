# AniList ID Library

## Goal

Let users save AniList titles by ID only, then retrieve current public title
and cover metadata in one batched query when viewing their library.

## Scope

- Add a user-to-AniList-ID relation with an added-at timestamp.
- Add protected add, remove, and paged list routes under `/library/anilist`.
- Batch-resolve the current metadata for a page of saved IDs through AniList.
- Preserve the existing MangaDex library relation and data.

## API and data changes

- `GET /library/anilist?limit=&offset=` returns saved AniList IDs with live
  public card metadata.
- `POST` and `DELETE /library/anilist/:anilistId` are idempotent.

## Validation

- Passed `go test ./modules/anilist-api ./controllers` and the required
  `go test ./modules/database` suite.
- Rebuilt the backend from `main.go` and restarted it on port 3001.
