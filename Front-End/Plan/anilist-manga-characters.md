# AniList Manga Characters

## Goal

Show AniList-provided manga characters in the metadata details page.

## Scope

- Render the characters returned with a manga-details request.
- Use a responsive two-column list with portrait, name, and role.
- Order the list by story prominence: main characters first, supporting
  characters second, then any other AniList roles. Preserve AniList's original
  order within each role.
- Do not add character favorites or load-more behavior in this change.

## API and data changes

- Consume the optional AniList `characters` connection in the existing manga
  details response.

## Implementation steps

1. Add typed character response fields.
2. Render a character section below the manga metadata.
3. Add focused UI coverage and validate the frontend server.
4. Sort the returned character edges in the details UI without mutating the
   cached AniList response.

## Validation

- Passed `npm test -- --watchAll=false --runInBand` (6 suites, 12 tests).
- Passed `npm run build`; it has an existing unused-type warning in
  `src/app/clients/backend/backend.ts` and standard Create React App dependency warnings.
- Restarted the development server; it compiled successfully with no type-check issues and
  is listening on port 3000.
- The role-ordering UI coverage passes as part of `npm test --
  --watchAll=false --runInBand` (6 suites, 16 tests). The production frontend
  rebuild also completed successfully; the existing lint and Create React App
  maintenance warnings remain outside this change's scope.

## Follow-up work

Add character pagination and character favorites separately.
