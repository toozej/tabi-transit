#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

if [[ $# -ne 1 ]]; then
  echo "usage: $0 /path/to/tabi-YYYYMMDDTHHMMSSZ.dump" >&2
  exit 2
fi

dump_file=$(realpath -- "$1")
[[ -r "$dump_file" && -s "$dump_file" ]] || { echo "readable non-empty dump required" >&2; exit 2; }
pg_restore --list "$dump_file" >/dev/null

deployment_root=${TABI_DEPLOYMENT_ROOT:-/opt/tabi}
cd "$deployment_root"
restore_project=${TABI_RESTORE_PROJECT:-tabi-restore}
[[ $restore_project =~ ^[a-z0-9][a-z0-9_-]*$ ]] || { echo "invalid restore project name" >&2; exit 2; }
[[ $restore_project != "${COMPOSE_PROJECT_NAME:-tabi}" ]] || {
  echo "restore project must differ from the live Compose project" >&2
  exit 2
}

compose=(docker compose -p "$restore_project" --env-file .env --env-file release.env -f compose.yaml -f compose.production.yaml)
"${compose[@]}" up -d postgres
for _ in $(seq 1 30); do
  if "${compose[@]}" exec -T postgres pg_isready -U "${POSTGRES_USER:-tabi}" -d "${POSTGRES_DB:-tabi}" >/dev/null; then
    break
  fi
  sleep 2
done
"${compose[@]}" exec -T postgres pg_isready -U "${POSTGRES_USER:-tabi}" -d "${POSTGRES_DB:-tabi}" >/dev/null
"${compose[@]}" exec -T postgres pg_restore --exit-on-error --no-owner --no-acl -U "${POSTGRES_USER:-tabi}" -d "${POSTGRES_DB:-tabi}" < "$dump_file"
"${compose[@]}" exec -T postgres psql -U "${POSTGRES_USER:-tabi}" -d "${POSTGRES_DB:-tabi}" -c 'SELECT 1'

echo "restore completed in isolated Compose project: $restore_project" >&2
echo "Run database integrity and application smoke checks before any switch; this script never changes the live project." >&2
