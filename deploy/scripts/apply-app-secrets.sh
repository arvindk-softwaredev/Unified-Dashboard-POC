#!/usr/bin/env bash
# Apply GITHUB_TOKEN and/or GEMINI_API_KEY to unified-dashboard-secrets.
# Preserves existing keys when only one env var is set.
set -euo pipefail

NAMESPACE="${NAMESPACE:-tektoncd-unified-dashboard}"
SECRET_NAME="${SECRET_NAME:-unified-dashboard-secrets}"

b64decode() {
  if base64 --help 2>&1 | grep -q GNU; then
    base64 -d
  else
    base64 -D
  fi
}

read_existing() {
  local key=$1
  local val
  val="$(oc get secret "${SECRET_NAME}" -n "${NAMESPACE}" \
    -o "jsonpath={.data.${key}}" 2>/dev/null || true)"
  if [[ -n "$val" ]]; then
    printf '%s' "$val" | b64decode
  fi
}

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  GITHUB_TOKEN="$(read_existing GITHUB_TOKEN || true)"
fi
if [[ -z "${GEMINI_API_KEY:-}" ]]; then
  GEMINI_API_KEY="$(read_existing GEMINI_API_KEY || true)"
fi

args=()
[[ -n "${GITHUB_TOKEN:-}" ]] && args+=(--from-literal=GITHUB_TOKEN="${GITHUB_TOKEN}")
[[ -n "${GEMINI_API_KEY:-}" ]] && args+=(--from-literal=GEMINI_API_KEY="${GEMINI_API_KEY}")

if [[ ${#args[@]} -eq 0 ]]; then
  echo "Set GITHUB_TOKEN and/or GEMINI_API_KEY (or create the secret manually)." >&2
  exit 1
fi

oc create secret generic "${SECRET_NAME}" "${args[@]}" \
  -n "${NAMESPACE}" --dry-run=client -o yaml | oc apply -f -

echo "Secret ${SECRET_NAME} applied in namespace ${NAMESPACE}"
oc rollout restart deployment/unified-dashboard-backend -n "${NAMESPACE}" 2>/dev/null || true
