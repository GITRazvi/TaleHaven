# Remove MangaKakalot scraper trial

## Goal

Remove the unsuccessful MangaKakalot scraper trial so it does not leave local
source files, Docker resources, or an unused port configuration in TaleHaven.

## Scope

- Remove the Compose-managed scraper container, its local image, and its
  Compose network.
- Delete `D:\TaleHaven\MangaKakalot-Scraper`, including the package install,
  Dockerfiles, tests, and runtime logs.
- Do not modify TaleHaven's frontend, backend, database, Redis, or Postgres
  services.

## API and data changes

- Removes the local-only scraper health endpoint and port 3003 listener.
- No TaleHaven API or database changes.

## Implementation steps

1. Bring down the dedicated Compose project and remove its local image.
2. Verify the scraper container is absent.
3. Delete the verified scraper workspace directory.
4. Confirm ports 3000, 3001, and 3003 are not affected by the removal.

## Validation

- Confirmed no scraper container was running before removal; removed the
  Compose project's local image and resources.
- Deleted the verified `D:\TaleHaven\MangaKakalot-Scraper` directory,
  including dependencies, tests, Docker configuration, and logs.
- Confirmed ports 3000 and 3001 remain served by TaleHaven, while port 3003
  has no listener. No scraper container or matching local image remains.

## Remaining follow-up work

- Evaluate a future manga provider under a separate integration plan.
