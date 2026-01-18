# Plan: Rename app to MegaBudget

## Approach
- Enumerate all references to the current app name and update to MegaBudget in code, manifests, docs, and plans.
- Update Kubernetes object names, image names, secrets, and database/user identifiers where they are app-scoped.
- Ensure module paths and import paths stay stable unless explicitly requested.

## Steps
1) Search for `spendtrack` and other app-name identifiers; classify by code, manifests, docs, and plans.
2) Update Go entrypoint path/name references as needed (binary, Docker image tags, k8s names).
3) Update kustomize resources (deployment/service/job names, secret refs, DB names/users).
4) Update documentation and plans to match the new app name.
5) Sanity-check for lingering references.

## Verification
- `rg -n "spendtrack|megabudget|go-db-app"`
- `kubectl kustomize deploy/kustomize/app/overlays/prod > /tmp/megabudget-app.yaml`
- `kubectl kustomize deploy/kustomize/db-migrations/overlays/prod > /tmp/megabudget-migrations.yaml`

## Rollback
- Revert the rename commit(s) or restore previous names in kustomize manifests and docs.
