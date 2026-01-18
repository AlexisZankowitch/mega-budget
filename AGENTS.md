# MegaBudget — Agent Guide (AGENTS.md)

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
- Namespace used for the prod app: `megabudget`

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
We provision the logical prod DB + role the same way:
- Prod overlay: `deploy/kustomize/db-provision/overlays/prod`
- Job name in prod: `db-provision-prod`

Config vs secrets:
- Non-secrets in ConfigMap (dev overlay):
  - `APP_DB_NAME=megabudget_dev`
  - `APP_DB_USER=megabudget_app_dev`
- Secrets are NOT committed:
  - `deploy/secrets/*.env` is ignored by git
  - `postgres-admin` secret contains admin creds
  - `megabudget-dev-db-secret` contains `APP_DB_PASSWORD` only
  - `megabudget-prod-db-secret` contains `APP_DB_NAME`, `APP_DB_USER`, and `APP_DB_PASSWORD`

## Migrations (pressly/goose)
- Migrations live in `migrations/` as SQL files.
- Schema migrations already exist:
  - init tables: `categories`, `transactions`
  - indexes: `(category_id, transaction_date)` and `transaction_date`
  - seed migration inserts default categories.
Kubernetes Job (prod):
- Base: `deploy/kustomize/db-migrations/base`
- Prod overlay: `deploy/kustomize/db-migrations/overlays/prod`
- Job name in prod: `db-migrate-prod`
- Keep `deploy/kustomize/db-migrations/base/migrations/` in sync with `migrations/` when adding new SQL files.

Typical dev commands (workstation):
- Set DB URL (password from local env file):
  - `source deploy/secrets/megabudget-dev-db-secret.env`
  - `export DEV_DATABASE_URL="postgres://megabudget_app_dev:${APP_DB_PASSWORD}@192.168.1.23:5432/megabudget_dev?sslmode=disable"`
- Then:
  - `goose -dir migrations postgres "$DEV_DATABASE_URL" status`
  - `goose -dir migrations postgres "$DEV_DATABASE_URL" up`

## Secrets workflow (keep it simple)
- Do not commit secrets.
- Keep passwords in local env files:
  - `deploy/secrets/postgres-admin.env`
  - `deploy/secrets/megabudget-dev-db-secret.env`
- Apply/update secrets (example pattern):
  - `kubectl -n databases create secret generic … --from-env-file=… --dry-run=client -o yaml | kubectl apply -f -`

## Kustomize / kubectl conventions
Use this kubeconfig for all cluster operations:
- `export KUBECONFIG="$HOME/.kube/k0s-zankowitch_1"`

Useful commands:
- Apply external postgres DNS wiring:
  - `kubectl apply -k deploy/kustomize/external-postgres/overlays/dev`
- Apply db provisioning (creates namespace/job/config):
  - `kubectl apply -k deploy/kustomize/db-provision/overlays/dev`
  - `kubectl apply -k deploy/kustomize/db-provision/overlays/prod`
- Re-run provisioning job:
  - `kubectl -n databases delete job db-provision-dev --ignore-not-found`
  - `kubectl apply -k deploy/kustomize/db-provision/overlays/dev`
  - `kubectl -n databases logs -l job-name=db-provision-dev --tail=200`
  - `kubectl -n databases delete job db-provision-prod --ignore-not-found`
  - `kubectl apply -k deploy/kustomize/db-provision/overlays/prod`
  - `kubectl -n databases logs -l job-name=db-provision-prod --tail=200`

App deployment (prod):
- Base: `deploy/kustomize/app/base`
- Prod overlay: `deploy/kustomize/app/overlays/prod`
- Apply:
  - `kubectl apply -k deploy/kustomize/app/overlays/prod`
- Update flow:
  - Build/push a new image tag (see registry section)
  - Update `deploy/kustomize/app/overlays/prod/patch-deployment.yaml` with the new image
  - `kubectl apply -k deploy/kustomize/app/overlays/prod`
  - If migrations changed, run the migration job before or right after the rollout

DB migrations (prod):
- Apply:
  - `kubectl apply -k deploy/kustomize/db-migrations/overlays/prod`
- Re-run:
  - `kubectl -n megabudget delete job db-migrate-prod --ignore-not-found`
  - `kubectl apply -k deploy/kustomize/db-migrations/overlays/prod`
  - `kubectl -n megabudget logs -l job-name=db-migrate-prod --tail=200`

## Coding conventions
- Keep `cmd/<app>/main.go` thin.
- Put logic in `internal/` packages (config/db/http/handlers).
- No ORM: SQL-first.
- Add a minimal health endpoint (`/healthz`).

## Local Docker Registry (LAN-only, HTTPS)

We use a private image registry hosted on the homelab server (same machine as the k0s node).

### Endpoint
- Registry: `registry.lan:5000`
- Do not use the IP (`192.168.1.23:5000`) when tagging images, the TLS cert is for `registry.lan`.

### Server setup
- Registry runs via Docker Compose in: `/srv/registry/compose.yaml`
- Persistent storage: `/srv/registry/data`
- TLS material: `/srv/registry/certs/`
  - `registry.crt`, `registry.key` (used by the registry)
  - `ca.crt` (CA to trust on clients and the k0s node)

Start/stop:
- `cd /srv/registry && sudo docker compose up -d`
- `cd /srv/registry && sudo docker compose down`

### Name resolution requirement
Both workstation and server must resolve `registry.lan` to the server LAN IP (e.g. via router DNS or `/etc/hosts`).

### Trust requirements (no “insecure registry”)
To push/pull without insecure settings, the CA must be trusted:
- On the server (k0s node):
  - `sudo cp /srv/registry/certs/ca.crt /usr/local/share/ca-certificates/registry-ca.crt`
  - `sudo update-ca-certificates`
- On the workstation:
  - Trust the same `ca.crt` (System trust store or Docker certs.d)

### Build & push workflow
Tag with the registry host, then push:
- `docker build -t registry.lan:5000/<app>:<tag> .`
- `docker push registry.lan:5000/<app>:<tag>`

### Kubernetes usage
Deployments must reference images like:
- `image: registry.lan:5000/<app>:<tag>`

(Optional) keep `imagePullPolicy: Always` for `:dev` tags.


## Planning for non-trivial changes
For any non-trivial change, create a new plan file under `plans/`:
- Name: `plans/YYYY-MM-DD_<short-topic>.md`
- Content: approach, ordered steps, verification commands, rollback notes.
- Keep the plan file updated if implementation details change later (treat it as the source of truth for that change).

## Keep this guide up to date
If you change workflows, repo layout, kustomize paths, namespaces, or the dev/prod database approach, update `AGENTS.md` in the same PR/commit.
