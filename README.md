# MegaBudget

Go app deployed to Kubernetes, connects to shared Postgres.
- DB provisioning and migrations run as Kubernetes Jobs
- App exposes /healthz
- Secrets are applied by a local deploy script (not committed)

## Local development (workstation)
Prereqs:
- Postgres is reachable on the LAN (dev database + role already provisioned)

Provision dev DB/user (once, or when resetting):
- Ensure local secret env files exist (not committed):
  - `deploy/secrets/postgres-admin.env` (POSTGRES_ADMIN_* values)
  - `deploy/secrets/megabudget-dev-db-secret.env` (APP_DB_PASSWORD)
- Apply secrets and run provisioning job:
```bash
export KUBECONFIG="$HOME/.kube/k0s-zankowitch_1"
kubectl -n databases create secret generic postgres-admin \
  --from-env-file=deploy/secrets/postgres-admin.env \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n databases create secret generic megabudget-dev-db-secret \
  --from-env-file=deploy/secrets/megabudget-dev-db-secret.env \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n databases delete job db-provision-dev --ignore-not-found
kubectl apply -k deploy/kustomize/db-provision/overlays/dev
kubectl -n databases logs -l job-name=db-provision-dev --tail=200
```

Setup:
```bash
source deploy/secrets/megabudget-dev-db-secret.env
export DEV_DATABASE_URL="postgres://megabudget_app_dev:${APP_DB_PASSWORD}@192.168.1.23:5432/megabudget_dev?sslmode=disable"
```

Migrations:
```bash
goose -dir migrations postgres "$DEV_DATABASE_URL" status
goose -dir migrations postgres "$DEV_DATABASE_URL" up
```

Run the app:
```bash
go run ./cmd/megabudget
curl http://localhost:8080/healthz
```

## Prod deploy (k0s)
Prereqs:
- External Postgres Service/Endpoints applied (`deploy/kustomize/external-postgres/overlays/dev`)
- Local secret env files (not committed)

Secrets (examples, stored locally under `deploy/secrets/`):
- `megabudget-prod-db-secret.env` with:
  - `APP_DB_NAME=megabudget`
  - `APP_DB_USER=megabudget_app`
  - `APP_DB_PASSWORD=...`

Note: the migration job uses SQL files from `deploy/kustomize/db-migrations/base/migrations/`, which should mirror `migrations/`.

Apply:
- `export KUBECONFIG="$HOME/.kube/k0s-zankowitch_1"`
- `kubectl -n databases create secret generic megabudget-prod-db-secret --from-env-file=deploy/secrets/megabudget-prod-db-secret.env --dry-run=client -o yaml | kubectl apply -f -`
- `kubectl -n megabudget create secret generic megabudget-prod-db-secret --from-env-file=deploy/secrets/megabudget-prod-db-secret.env --dry-run=client -o yaml | kubectl apply -f -`
- `kubectl apply -k deploy/kustomize/db-provision/overlays/prod`
- `kubectl apply -k deploy/kustomize/db-migrations/overlays/prod`
- `kubectl apply -k deploy/kustomize/app/overlays/prod`

## Prod rollout update
Use this when you have a new app build to deploy.

1) Build and push a new image tag (multi-arch):
```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  -t registry.lan:5000/megabudget:<tag> --push .
```

2) Update the deployment image tag:
- Edit `deploy/kustomize/app/overlays/prod/patch-deployment.yaml`
- Set:
  - `image: registry.lan:5000/megabudget:<tag>`

3) Apply the deployment:
```bash
export KUBECONFIG="$HOME/.kube/k0s-zankowitch_1"
kubectl apply -k deploy/kustomize/app/overlays/prod
kubectl -n megabudget rollout status deployment/megabudget-prod
```

4) If migrations changed, run them before or right after the rollout:
```bash
kubectl -n megabudget delete job db-migrate-prod --ignore-not-found
kubectl apply -k deploy/kustomize/db-migrations/overlays/prod
kubectl -n megabudget logs -l job-name=db-migrate-prod --tail=200
```
