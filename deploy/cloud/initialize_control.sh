#!/bin/sh
set -eu
for service in auth repo usage customer narrow; do
  docker compose -f /srv/slang/compose.data.yml -f /srv/slang/compose.control.yml run --rm --no-deps -e INIT_SCHEMA=true "$service" python -c 'import index; print("Schema ready")'
done
docker compose -f /srv/slang/compose.data.yml -f /srv/slang/compose.control.yml up -d
