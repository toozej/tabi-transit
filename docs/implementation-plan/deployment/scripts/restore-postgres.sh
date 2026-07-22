#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

if [[ ${RESTORE_CONFIRM:-} != "YES" || $# -ne 1 ]]; then
  echo "usage: RESTORE_CONFIRM=YES $0 /path/to/verified.dump" >&2
  exit 2
fi

dump=$(realpath "$1")
test -s "$dump"
cd /opt/tabi

echo "This template restores into the configured database."
echo "Production operators must first stop writers and validate a separate restore target."

docker compose --env-file .env --env-file release.env \
  -f compose.yaml -f compose.production.yaml \
  exec -T postgres sh -c \
  'dropdb --if-exists -U "$POSTGRES_USER" "$POSTGRES_DB" &&
   createdb -U "$POSTGRES_USER" "$POSTGRES_DB"'

docker compose --env-file .env --env-file release.env \
  -f compose.yaml -f compose.production.yaml \
  exec -T postgres sh -c \
  'pg_restore --exit-on-error --no-owner --no-acl -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
  < "$dump"
