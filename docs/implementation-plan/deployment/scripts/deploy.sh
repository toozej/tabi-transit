#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

if [[ $# -ne 1 ]]; then
  echo "usage: $0 /path/to/release.env.new" >&2
  exit 2
fi

cd /opt/tabi
exec 9>/run/lock/tabi-deploy.lock
flock -n 9 || { echo "another deployment is running" >&2; exit 1; }

new_release=$(realpath "$1")
test -s "$new_release"

compose=(docker compose --env-file .env --env-file "$new_release" -f compose.yaml -f compose.production.yaml)
"${compose[@]}" config --quiet

./deployment/scripts/backup-postgres.sh

"${compose[@]}" pull
"${compose[@]}" --profile jobs run --rm migrate

if [[ -f release.env ]]; then
  cp -f release.env release.previous.env
fi
cp -f "$new_release" release.env

active=(docker compose --env-file .env --env-file release.env -f compose.yaml -f compose.production.yaml)
"${active[@]}" up -d --remove-orphans

for _ in $(seq 1 30); do
  if curl --fail --silent --show-error "https://${TABI_API_DOMAIN:-$(grep '^TABI_API_DOMAIN=' .env | cut -d= -f2-)}/health/ready" >/dev/null; then
    exit 0
  fi
  sleep 5
done

echo "health check failed; attempting image rollback" >&2
if [[ -f release.previous.env ]]; then
  rollback=(docker compose --env-file .env --env-file release.previous.env -f compose.yaml -f compose.production.yaml)
  "${rollback[@]}" up -d --remove-orphans
fi
exit 1
