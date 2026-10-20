# Manga Hook Local Setup Plan

## Goal

Document how a developer can run Manga Hook locally while TaleHaven prepares
to support it as an optional secondary manga provider.

## Scope

- Add the upstream Manga Hook repository link to the frontend README.
- Provide concise Windows/PowerShell instructions to clone, install, configure,
  start, and smoke-test the server on port 3002.
- Clarify that TaleHaven does not yet send requests to Manga Hook.

## API or data changes

None. This is documentation only; no TaleHaven routes, environment variables,
or provider-selection behavior changes.

## Implementation steps

1. Document the non-conflicting local port alongside the existing TaleHaven
   frontend (3000) and backend (3001) ports.
2. Link to Manga Hook's README as the source of its setup and endpoint details.
3. State the boundary between the optional server process and the current
   MangaDex-only application integration.

## Validation

- Review the README commands and endpoint against the upstream Manga Hook
  README and server source.
- Confirm the documentation does not claim that Manga Hook is already wired
  into TaleHaven.

## Remaining follow-up work

- Add backend configuration and a provider adapter before TaleHaven uses
  Manga Hook requests.
- Add caching, timeout, rate-limiting, fallback, and provider-attribution
  behavior as part of that integration.
