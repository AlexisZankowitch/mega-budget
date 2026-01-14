# Spendtrack — Agent Guide (AGENTS.md)

## Goal
Build a small “spending tracker” Go service backed by Postgres.
- Dev: run the Go app on the workstation against a dev database.
- Later: deploy the app to the k0s cluster, keep the same Postgres server but use a different logical database.

## Current architecture (today)
### Postgres
- Postgres runs on the server via Docker (outside Kubernetes).
- Postgres is exposed on the LAN (not only 127.0.0.1) so pods can reach it.

### Kubernetes cluster
- Cluster: k0s (kubeconfig: `~/.kube/k0s-zankowitch_1`)
- Namespace used for DB-related objects: `databases`

### “External Postgres” inside cluster (Service + Endpoints)
We provide a stable in-cluster DNS name for the external Postgres:
- Service/Endpoints: `postgres` in namespace `databases`
- DNS: `postgres.databases.svc.cluster.local`
- Manifests: `deploy/kustomize/external-postgres/overlays/dev`

### DB provisioning (logical DB + role per app)
We provision the logical dev DB + role by running a Kubernetes Job that executes SQL via `psql`:
- Base: `deploy/kustomize/db-provision/base`
- Dev overlay: `deploy/kustomize/db-provision/overlays/dev`
- Job name in dev: `db-provision-dev`

Config vs secrets:
- Non-secrets in ConfigMap (dev overlay):
  - `APP_DB_NAME=spendtrack_dev`
  - `APP_DB_USER=spendtrack_app_dev`
- Secrets are NOT committed:
  - `deploy/secrets/*.env` is ignored by git
  - `postgres-admin` secret contains admin creds
  - `spendtrack-dev-db-secret` contains `APP_DB_PASSWORD` only

## Migrations (pressly/goose)
- Migrations live in `migrations/` as SQL files.
- Schema migrations already exist:
  - init tables: `categories`, `transactions`
  - indexes: `(category_id, transaction_date)` and `transaction_date`
  - seed migration inserts default categories.

Typical dev commands (workstation):
- Set DB URL (no password in URL):
  - `export DEV_DATABASE_URL='postgres://spendtrack_app_dev@192.168.1.23:5432/spendtrack_dev?sslmode=disable'`
- Ensure Goose can read `.pgpass`:
  - `export PGPASSFILE="$HOME/.pgpass"`
- Then:
  - `goose -dir migrations postgres "$DEV_DATABASE_URL" status`
  - `goose -dir migrations postgres "$DEV_DATABASE_URL" up`

## Secrets workflow (keep it simple)
- Do not commit secrets.
- Keep passwords in local env files:
  - `deploy/secrets/postgres-admin.env`
  - `deploy/secrets/spendtrack-dev-db-secret.env`
- Apply/update secrets (example pattern):
  - `kubectl --kubeconfig "$KCFG" -n databases create secret generic … --from-env-file=… --dry-run=client -o yaml | kubectl apply -f -`

## Kustomize / kubectl conventions
Use this kubeconfig for all cluster operations:
- `export KCFG="$HOME/.kube/k0s-zankowitch_1"`

Useful commands:
- Apply external postgres DNS wiring:
  - `kubectl --kubeconfig "$KCFG" apply -k deploy/kustomize/external-postgres/overlays/dev`
- Apply db provisioning (creates namespace/job/config):
  - `kubectl --kubeconfig "$KCFG" apply -k deploy/kustomize/db-provision/overlays/dev`
- Re-run provisioning job:
  - `kubectl --kubeconfig "$KCFG" -n databases delete job db-provision-dev --ignore-not-found`
  - `kubectl --kubeconfig "$KCFG" apply -k deploy/kustomize/db-provision/overlays/dev`
  - `kubectl --kubeconfig "$KCFG" -n databases logs -l job-name=db-provision-dev --tail=200`

## Coding conventions (when we implement the Go service)
- Keep `cmd/<app>/main.go` thin.
- Put logic in `internal/` packages (config/db/http/handlers).
- No ORM: SQL-first.
- Add a minimal health endpoint early (`/healthz`) and later metrics (`/metrics`).

## Planning for non-trivial changes
For any non-trivial change, create a new plan file under `plans/`:
- Name: `plans/YYYY-MM-DD_<short-topic>.md`
- Content: approach, ordered steps, verification commands, rollback notes.
- Keep the plan file updated if implementation details change later (treat it as the source of truth for that change).

## Keep this guide up to date
If you change workflows, repo layout, kustomize paths, namespaces, or the dev/prod database approach, update `AGENTS.md` in the same PR/commit.

