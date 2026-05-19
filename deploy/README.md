# Deploy Unified Dashboard

## Architecture (2 pods — separate images, one URL)

```
Internet
   │
   ▼
OpenShift Route  (single public URL — only the frontend is exposed)
   │
   ▼
frontend Pod (nginx)  ──proxy /api/*──►  backend Pod (Go API)
   image: .../tektoncd-unified-dashboard-frontend   image: .../tektoncd-unified-dashboard-backend
```

| Workload | Image | Exposed? |
|----------|--------|----------|
| `unified-dashboard-frontend` | Quay `.../tektoncd-unified-dashboard-frontend:tag` | **Yes** — Route points here |
| `unified-dashboard-backend` | Quay `.../tektoncd-unified-dashboard-backend:tag` | **No** — cluster-internal Service only |

The browser only opens the **Route URL**. Nginx serves the UI and forwards `/api` to the backend Service.

After deploy, the app URL is in the GitHub Actions job summary, or:

```bash
oc get route unified-dashboard -n tektoncd-unified-dashboard -o jsonpath='https://{.spec.host}{"\n"}'
```

---

## Local (Docker Compose)

From the repo root:

```bash
cp backend/.env.example backend/.env   # GITHUB_TOKEN, GEMINI_API_KEY
DOCKER_PLATFORM= make -C deploy docker-up   # host arch for Apple Silicon; omit on amd64
open http://localhost:8080
```

Or:

```bash
cp backend/.env.example backend/.env
docker compose -f deploy/docker-compose.yml up --build
```

Stop: `make -C deploy docker-down`

Build images only (e.g. before pushing to Quay):

```bash
export QUAY_REGISTRY=quay.io/your-org
export TAG=latest
make -C deploy docker-build REGISTRY=$QUAY_REGISTRY TAG=$TAG
```

---

## Automated deploy (GitHub Actions)

Workflow: [`.github/workflows/deploy-openshift.yml`](../.github/workflows/deploy-openshift.yml)

The workflow builds and pushes both images to Quay, deploys to OpenShift, applies secrets, and prints the Route URL.

### One-time on [quay.io](https://quay.io)

Create repositories:

- `<your-org>/tektoncd-unified-dashboard-backend`
- `<your-org>/tektoncd-unified-dashboard-frontend`

### GitHub configuration

**Repository variable:** `QUAY_REGISTRY` (e.g. `quay.io/rh-ee-apalit`)

**Secrets:** `QUAY_USERNAME`, `QUAY_PASSWORD`, `OPENSHIFT_SERVER`, `OPENSHIFT_TOKEN`, `GEMINI_API_KEY`, and optionally `GH_PAT` (else uses `GITHUB_TOKEN`).

**Run:** Actions → **Deploy to OpenShift** → Run workflow.

Manifests live under `deploy/k8s/`; deploy scripts used by CI:

- `deploy/scripts/ensure-namespace.sh`
- `deploy/scripts/apply-app-secrets.sh`

Image names/tags are patched at deploy time in `deploy/k8s/overlays/openshift/`.

---

## Container base images (Red Hat UBI)

| Stage | Image |
|-------|--------|
| Go build | `registry.access.redhat.com/ubi9/go-toolset` |
| Go runtime | `registry.access.redhat.com/ubi9/ubi-minimal` |
| Frontend build | `registry.access.redhat.com/ubi9/nodejs-22` (+ Bun) |
| Frontend runtime | `registry.access.redhat.com/ubi9/nginx-124` |
