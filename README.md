# go-db-app

Go app deployed to Kubernetes, connects to shared Postgres.
- DB provisioning and migrations run as Kubernetes Jobs
- App exposes /healthz and /metrics
- Secrets are applied by a local deploy script (not committed)

## Prod deploy (k0s)
Prereqs:
- External Postgres Service/Endpoints applied (`deploy/kustomize/external-postgres/overlays/dev`)
- Local secret env files (not committed)

Secrets (examples, stored locally under `deploy/secrets/`):
- `spendtrack-prod-db-secret.env` with:
  - `APP_DB_PASSWORD=...`
  - `DATABASE_URL=postgres://spendtrack_app:<password>@postgres.databases.svc.cluster.local:5432/spendtrack?sslmode=disable`

Apply:
- `export KUBECONFIG="$HOME/.kube/k0s-zankowitch_1"`
- `kubectl -n databases create secret generic spendtrack-prod-db-secret --from-env-file=deploy/secrets/spendtrack-prod-db-secret.env --dry-run=client -o yaml | kubectl apply -f -`
- `kubectl apply -k deploy/kustomize/db-provision/overlays/prod`
- `kubectl apply -k deploy/kustomize/app/overlays/prod`
