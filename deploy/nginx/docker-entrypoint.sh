#!/bin/sh
set -eu

UPSTREAM="${BACKEND_UPSTREAM:-127.0.0.1:8081}"
sed "s|__BACKEND_UPSTREAM__|${UPSTREAM}|g" /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf
exec nginx -g "daemon off;"
