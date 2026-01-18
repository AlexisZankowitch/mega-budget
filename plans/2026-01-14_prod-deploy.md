# Plan: Prod megabudget deployment

## Approach
- Create prod DB provisioning overlay that runs in `databases` namespace and provisions `megabudget` + `megabudget_app`.
- Add a new `megabudget` namespace and app manifests (Deployment + Service) with DB connection from a Secret.
- Keep secrets local in `deploy/secrets/*.env` and apply via kubectl, same as dev.
- Validate each kustomize overlay via `kubectl kustomize` before applying.

## Steps
1) Add `deploy/kustomize/db-provision/overlays/prod` overlay.
2) Add `deploy/kustomize/app/base` and `deploy/kustomize/app/overlays/prod` for the prod deployment.
3) Document the prod secret env file names and apply commands in `README.md`.
4) Validate with `kubectl kustomize` for both overlays.

## Verification
- `kubectl kustomize deploy/kustomize/db-provision/overlays/prod`
- `kubectl kustomize deploy/kustomize/app/overlays/prod`

## Rollback
- Delete the prod Job and app resources, or remove the new overlays.
