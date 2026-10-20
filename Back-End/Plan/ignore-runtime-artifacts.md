# Ignore Backend Runtime Artifacts

## Goal

Keep request-generated images and locally built backend artifacts out of Git
pushes while retaining them on a developer's machine.

## Scope

- Ignore the backend `images/` runtime cache.
- Ignore local backend logs and executable build outputs.
- Remove the currently tracked versions of those runtime artifacts from the
  Git index only; do not delete local files.

## API or data changes

None. The backend continues serving and writing its local `images/` directory
as before.

## Implementation steps

1. Add backend-scoped ignore rules for the runtime artifacts.
2. Remove existing cached images, logs, and executables from Git tracking.
3. Verify the rules apply and the local files remain present.

## Validation

- `git check-ignore` confirms that images, logs, and executable outputs match
  the new rules.
- Confirmed that no files remain indexed under `Back-End/images/` and that the
  local images directory still exists.
- `git diff --cached --check` completed without whitespace errors.

## Follow-up work

None.
