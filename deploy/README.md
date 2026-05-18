# Deploy Unified Dashboard

## Architecture (2 pods — independent deploys)

```
Route/Ingress → frontend Service → nginx Pod
                                      └─ proxy /api → backend Service → Go API Pod
```

| Workload | Deploy independently | Image |
|----------|----------------------|-------|
| `unified-dashboard-backend` | API / tracking changes | `unified-dashboard-backend` |
| `unified-dashboard-frontend` | UI changes | `unified-dashboard-frontend` |

## GitHub token (fix rate limits)

1. Put the token in **`backend/.env`** (not the repo root):
   ```bash
   cp backend/.env.example backend/.env
   # edit: GITHUB_TOKEN=ghp_...
   ```
2. Restart the backend from `backend/` or any directory (loads `.env` and `backend/.env`).
3. On startup you should see: `GITHUB_TOKEN loaded — using authenticated GitHub API`
4. Check `GET /api/health` → `"github_authenticated": true`

**Security:** Never commit tokens. If a token was committed, **revoke it** on GitHub and create a new one.

**Why limits still happen with a token:** Search API is capped at **30 requests/minute** (separate from the 5000/hr core limit). The tracking dashboard uses search + per-repo CI calls. We cache `/api/tracking/summary` for 2 minutes and pace search calls.

## Local (Docker Compose)

```bash
cp backend/.env.example backend/.env   # add GITHUB_TOKEN
make -C deploy docker-up
open http://localhost:8080
```

## OpenShift

```bash
oc new-project unified-dashboard
oc create secret generic unified-dashboard-secrets \
  --from-literal=GITHUB_TOKEN=<token> -n unified-dashboard

export REGISTRY=image-registry.openshift-image-registry.svc:5000/unified-dashboard
make -C deploy docker-build REGISTRY=$REGISTRY
# push images, then:
oc apply -k deploy/k8s/
oc apply -f deploy/k8s/route-openshift.yaml
oc get route unified-dashboard-frontend -n unified-dashboard
```

## Roll out only one component

```bash
oc set image deployment/unified-dashboard-backend \
  backend=$REGISTRY/unified-dashboard-backend:new-tag -n unified-dashboard

oc set image deployment/unified-dashboard-frontend \
  frontend=$REGISTRY/unified-dashboard-frontend:new-tag -n unified-dashboard
```

## Tekton

See `deploy/tekton/pipeline.yaml` — deploy task patches **both** deployments independently.
