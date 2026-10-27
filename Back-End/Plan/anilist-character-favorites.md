# AniList Character Favorites

## Goal

Allow each signed-in reader to favorite AniList characters and return their
current public display metadata for My Profile.

## Scope

- Persist a per-user AniList character ID, its source manga ID, and the
  character's role in that manga.
- Add protected list, add, and remove endpoints below `/favorites/anilist/characters`.
- Resolve current character names and portraits through AniList when listing
  profile favorites.
- Do not add character detail pages or character pagination in this change.

## Character detail extension

- Add a public, on-demand character-details endpoint with AniList's biography,
  personal fields, and first page of associated manga/anime.
- Keep character details out of the initial manga-details response so opening a
  manga page remains lightweight.

## API / data changes

- Add `user_favorite_anilist_character` with a user/character primary key.
- Add database operations for listing, adding, and removing character
  favorites.
- Add AniList character batch retrieval for profile display metadata.

## Implementation steps

1. Add the migration, operations, and database coverage.
2. Expose authenticated favorite routes and character enrichment.
3. Verify the frontend can toggle favorites from manga details and display
   favorite character cards in My Profile.
4. Add and validate the on-demand character-details query and endpoint.

## Validation

- Passed `go test ./modules/anilist-api ./modules/database ./controllers`.
- Passed the required `go test ./modules/database` immediately before restart.
- Rebuilt and restarted the backend; it is listening on port 3001.
- Confirmed `GET /favorites/anilist/characters` is registered (returns the
  expected unauthenticated `401`, not `404`).
- Passed the AniList client and controller tests after adding the public
  on-demand character-details endpoint.
- Rebuilt and restarted the backend after the character-details extension.

## Follow-up

Character detail pages and a dedicated full favorites page can be considered
later.
