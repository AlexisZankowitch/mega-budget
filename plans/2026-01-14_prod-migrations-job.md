# Plan: Prod database migrations job

## Approach
- Add a kustomize base/overlay for a Kubernetes Job that runs goose migrations against the prod database.
- Package migration SQL files into a ConfigMap mounted into the Job container.
- Use the existing prod DB secret to provide `DATABASE_URL`.
- Validate kustomize output locally before any apply.

## Steps
1) Create `deploy/kustomize/db-migrations/base` with Job + ConfigMap generator for `migrations/*.sql`.
2) Add `deploy/kustomize/db-migrations/overlays/prod` to set namespace and secret name.
3) Update `AGENTS.md` and `README.md` with prod migration commands.
4) Validate with `kubectl kustomize deploy/kustomize/db-migrations/overlays/prod`.

## Verification
- `kubectl kustomize deploy/kustomize/db-migrations/overlays/prod`

## Rollback
- Delete the migration Job or remove the new kustomize files.
