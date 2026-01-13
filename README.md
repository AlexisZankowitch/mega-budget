# go-db-app

Go app deployed to Kubernetes, connects to shared Postgres.
- DB provisioning and migrations run as Kubernetes Jobs
- App exposes /healthz and /metrics
- Secrets are applied by a local deploy script (not committed)
