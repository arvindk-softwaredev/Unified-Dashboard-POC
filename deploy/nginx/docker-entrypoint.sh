#!/bin/bash
set -euo pipefail

UPSTREAM="${BACKEND_UPSTREAM:-127.0.0.1:8081}"
sed "s|__BACKEND_UPSTREAM__|${UPSTREAM}|g" \
  /opt/app-root/etc/nginx.d/upstream.conf.template \
  > /opt/app-root/etc/nginx.d/upstream.conf

exec nginx -g "daemon off;"
