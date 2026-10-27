# Remove MangaHook integration

## Goal

Remove the deleted MangaHook service's frontend tester, navigation, API client
methods, and temporary search presentation.

## Scope

- Delete the `/mangahook` route, navigation entry, tester component, and test.
- Remove MangaHook-specific client types and methods.
- Restore normal MangaDex search cards and response handling.
- Remove obsolete local-server documentation and the prior MangaHook setup
  plan.

## API and data changes

- The frontend will no longer call `/mangahook/*` endpoints.
- Search responses use the normal MangaDex paged response type.

## Implementation steps

1. Remove tester navigation, route, component, test, and client methods.
2. Remove temporary provider badges, raw-response display, and non-clickable
   search-card behavior.
3. Remove stale MangaHook startup documentation.
4. Run frontend tests and confirm the development server recompiles.

## Validation

- Passed `npm test -- --watchAll=false --runInBand`: 4 suites and 8 tests.
- Passed the TypeScript no-output check with zero errors.
- Restarted the frontend and confirmed it compiled successfully on port 3000.

## Remaining follow-up work

- Keep future scraper evaluations behind a separate, explicitly approved
  integration plan.
