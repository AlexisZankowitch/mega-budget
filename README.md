# go-db-app

Go app deployed to Kubernetes, connects to shared Postgres.
- DB provisioning and migrations run as Kubernetes Jobs
- App exposes /healthz and /metrics
- Secrets are applied by a local deploy script (not committed)

## Local development (workstation)
Prereqs:
- Postgres is reachable on the LAN (dev database + role already provisioned)
- Local `.pgpass` contains the dev DB password for Goose

Setup:
- `export DEV_DATABASE_URL='postgres://spendtrack_app_dev@192.168.1.23:5432/spendtrack_dev?sslmode=disable'`
- `export PGPASSFILE="$HOME/.pgpass"`

Migrations:
- `goose -dir migrations postgres "$DEV_DATABASE_URL" status`
- `goose -dir migrations postgres "$DEV_DATABASE_URL" up`

Run the app:
- `go run ./cmd/spendtrack`
- `curl http://localhost:8080/healthz`

## Prod deploy (k0s)
Prereqs:
- External Postgres Service/Endpoints applied (`deploy/kustomize/external-postgres/overlays/dev`)
- Local secret env files (not committed)

Secrets (examples, stored locally under `deploy/secrets/`):
- `spendtrack-prod-db-secret.env` with:
  - `APP_DB_NAME=spendtrack`
  - `APP_DB_USER=spendtrack_app`
  - `APP_DB_PASSWORD=...`

Note: the migration job uses SQL files from `deploy/kustomize/db-migrations/base/migrations/`, which should mirror `migrations/`.

Apply:
- `export KUBECONFIG="$HOME/.kube/k0s-zankowitch_1"`
- `kubectl -n databases create secret generic spendtrack-prod-db-secret --from-env-file=deploy/secrets/spendtrack-prod-db-secret.env --dry-run=client -o yaml | kubectl apply -f -`
- `kubectl -n spendtrack create secret generic spendtrack-prod-db-secret --from-env-file=deploy/secrets/spendtrack-prod-db-secret.env --dry-run=client -o yaml | kubectl apply -f -`
- `kubectl apply -k deploy/kustomize/db-provision/overlays/prod`
- `kubectl apply -k deploy/kustomize/db-migrations/overlays/prod`
- `kubectl apply -k deploy/kustomize/app/overlays/prod`

## Prod rollout update
Use this when you have a new app build to deploy.

1) Build and push a new image tag (multi-arch):
```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  -t registry.lan:5000/spendtrack:<tag> --push .
```

2) Update the deployment image tag:
- Edit `deploy/kustomize/app/overlays/prod/patch-deployment.yaml`
- Set:
  - `image: registry.lan:5000/spendtrack:<tag>`

3) Apply the deployment:
```bash
export KUBECONFIG="$HOME/.kube/k0s-zankowitch_1"
kubectl apply -k deploy/kustomize/app/overlays/prod
kubectl -n spendtrack rollout status deployment/spendtrack-prod
```

4) If migrations changed, run them before or right after the rollout:
```bash
kubectl -n spendtrack delete job db-migrate-prod --ignore-not-found
kubectl apply -k deploy/kustomize/db-migrations/overlays/prod
kubectl -n spendtrack logs -l job-name=db-migrate-prod --tail=200
```
