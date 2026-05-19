#!/usr/bin/env bash
# Ensure an OpenShift namespace/project exists, then select it.
set -euo pipefail

NS="${1:-${NAMESPACE:-tektoncd-unified-dashboard}}"

if oc get namespace "$NS" >/dev/null 2>&1; then
  echo "Namespace exists: $NS"
else
  echo "Creating namespace: $NS"
  oc new-project "$NS" \
    --display-name="Tekton CD Unified Dashboard" \
    --description="Unified Dashboard (frontend + backend)"
fi

oc project "$NS" >/dev/null
echo "Active namespace: $(oc project -q)"
