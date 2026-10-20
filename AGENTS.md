# TaleHaven Working Agreement

Read this file before inspecting, planning, editing, or running the TaleHaven project.

## Change discipline

- Make no changes unless the user has asked for them.
- Keep changes scoped to the active request and preserve unrelated user work.
- Before starting a feature or a meaningful technical change, create or update a Markdown plan in the relevant `Plan/` directory. If the work spans both applications, update a plan in both directories.
- Plan files are part of the repository and should be included in Git commits alongside the implementation they describe.
- Record the goal, scope, API or data changes, implementation steps, validation, and remaining follow-up work in each plan.

## Running services

- Before every backend restart, run `go test ./modules/database` from
  `Back-End`. On this Windows workspace, use
  `& 'C:\Program Files\Go\bin\go.exe' test ./modules/database` when `go`
  is not on `PATH`. This includes the database-backed library add/list/remove
  and per-user isolation test.
- Before every frontend restart, run `npm test -- --watchAll=false --runInBand`
  from `Front-End`.
- After any change to backend code, rebuild and restart the backend server before handing the work back.
- After any change to frontend code, confirm the frontend development server has recompiled; restart it when its configuration or dependencies changed, or when a clean restart is useful for validation.
- When both applications change, restart both services.
- Confirm server startup from the local logs and report any validation limitation clearly.

## Validation

- Run formatting, focused tests, and a relevant build or type check when the required tools are available.
- Do not claim a server route or UI behavior is verified unless the running code contains the change and the relevant validation was performed.
