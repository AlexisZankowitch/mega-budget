# Plan: Prod image build and deploy update

## Approach
- Add a minimal multi-stage Dockerfile for the megabudget app.
- Add a .dockerignore to keep the build context lean.
- Build and push the image to the local registry with tag `prod-2026-01-14`.
- Update the prod kustomize overlay to use the registry image tag.

## Steps
1) Create Dockerfile and .dockerignore.
2) Build and push `registry.lan:5000/megabudget:prod-2026-01-14` from repo root.
3) Update `deploy/kustomize/app/overlays/prod/patch-deployment.yaml` with the image tag.
4) Apply the prod app overlay and verify rollout.

## Verification
- `docker build -t registry.lan:5000/megabudget:prod-2026-01-14 .`
- `docker push registry.lan:5000/megabudget:prod-2026-01-14`
- `kubectl apply -k deploy/kustomize/app/overlays/prod`
- `kubectl -n megabudget rollout status deployment/megabudget-prod`

## Rollback
- Revert the image tag in the prod overlay or redeploy a previous tag.
