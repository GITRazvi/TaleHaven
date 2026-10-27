# AniList search with MangaDex reading matches

## Goal

Use AniList as the public catalog search source and resolve MangaDex chapters only when a reader explicitly starts a title.

## Scope

- Add a paged AniList title-search route.
- Extend the public AniList search route so one request can apply optional
  title, genre, release-year, format, publication-status, and sort filters.
- Validate filter values before querying AniList and allow filter-only browsing
  when at least one filter is selected.
- Omit unselected GraphQL arguments instead of passing `null`, because AniList
  returns an empty result set when a null `search` argument accompanies a
  filter-only request.
- Support AniList tags separately from genres, so catalog labels such as
  Isekai, Magic, and Martial Arts use AniList's `tag` argument.
- Add a reading-source resolver that retrieves AniList title variants and searches MangaDex for one confident match.
- Normalize punctuation, case, and accents; accept exact normalized titles or safe token matches.
- Query MangaDex with every distinct AniList title and synonym, then compare every returned MangaDex primary and alternative title as an order-independent word multiset. This avoids extra permutation requests while recognizing reordered translated titles.
- When otherwise-safe title candidates are tied, use AniList's publication year and author/artist metadata as conservative tie-breakers; still reject ties without one clear winner.
- Verify a matched MangaDex work exposes at least one chapter before reporting it as a readable source.
- Limit character detail appearances to AniList manga so character panels do not list anime.
- Protect public AniList metadata requests with a short-lived response cache, in-flight request coalescing, paced upstream calls, and one rate-limit retry.

## API and data

- `GET /anilist/manga/search?title=&genre=&tag=&year=&format=&status=&sort=&page=&perPage=` returns AniList metadata filtered upstream by AniList.
- `GET /anilist/manga/reading-source/:anilistId` returns `{ available, mangaId? }` only when the confident MangaDex match has at least one chapter.
- AniList metadata calls are cached for two minutes per query and variables; duplicate concurrent calls share one upstream request.

## Validation

- Add focused title-matcher tests for exact, reordered-token, ambiguous, and metadata tie-break candidates.
- Run the AniList controller and MangaDex client tests after enforcing source availability and manga-only character appearances.
- Add AniList client tests for cached, coalesced, and retried rate-limited requests.
- Run Go formatter, database tests before restart, and relevant backend tests.
- Passed controller and MangaDex matcher tests. The host database suite could
  not reach port 5433 because Windows blocked that forwarded port, while the
  rebuilt backend connected to Docker Postgres, completed migrations, and
  started on port 3001.
- Passed the focused MangaDex matcher and AniList controller tests after
  adding order-independent alternate-title matching. The required database
  suite also passes. The rebuilt backend is running on port 3001 and its logs
  confirm successful Redis/Postgres connections and migrations.
- Passed AniList client and controller tests for filter variables and invalid
  requests, plus `go test ./modules/database`. Rebuilt and restarted the
  backend. A live filter-only request returned HTTP 200 from AniList.
- Added regression coverage that generated filter-only queries omit absent
  variables. Focused AniList/controller tests and `go test ./modules/database`
  pass. The rebuilt backend is live on port 3001; an Action-only request now
  returns AniList results.
- Added tag-variable coverage. Focused AniList/controller tests and
  `go test ./modules/database` pass. The rebuilt backend now returns live
  AniList results for an Isekai-only request.

## Follow-up

Matching deliberately favors false negatives over attaching the wrong chapter source. A future UI can offer a manual source picker for ambiguous results.
