#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

cd /opt/tabi
exec 9>/run/lock/tabi-postgres-backup.lock
flock -n 9 || exit 0

backup_dir=/var/lib/tabi/backups/postgres
mkdir -p "$backup_dir"
timestamp=$(date -u +%Y%m%dT%H%M%SZ)
tmp="$backup_dir/tabi-$timestamp.dump.tmp"
final="$backup_dir/tabi-$timestamp.dump"

docker compose --env-file .env --env-file release.env \
  -f compose.yaml -f compose.production.yaml \
  exec -T postgres sh -c \
  'pg_dump --format=custom --no-owner --no-acl -U "$POSTGRES_USER" "$POSTGRES_DB"' \
  > "$tmp"

test -s "$tmp"
mv "$tmp" "$final"
sha256sum "$final" > "$final.sha256"

find "$backup_dir" -type f -mtime +14 -delete

if [[ -f /etc/tabi/secrets/restic_environment ]]; then
  # shellcheck disable=SC1091
  source /etc/tabi/secrets/restic_environment
  export RESTIC_PASSWORD_FILE=/etc/tabi/secrets/restic_password
  restic backup \
    /var/lib/tabi/backups \
    /var/lib/tabi/feed-archive \
    /var/lib/tabi/static-artifacts \
    /opt/tabi/Caddyfile \
    /opt/tabi/compose.yaml \
    /opt/tabi/compose.production.yaml
  restic forget --keep-daily 14 --keep-weekly 8 --keep-monthly 12 --prune
fi
