#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

deployment_root=${TABI_DEPLOYMENT_ROOT:-/opt/tabi}
cd "$deployment_root"
lock_dir=${TABI_LOCK_DIR:-/run/lock}
exec 9>"$lock_dir/tabi-postgres-backup.lock"
flock -n 9 || exit 0

backup_dir=${TABI_BACKUP_DIR:-/var/lib/tabi/backups}/postgres
install -d -m 0700 "$backup_dir"
timestamp=$(date -u +%Y%m%dT%H%M%SZ)
temporary_dump="$backup_dir/tabi-$timestamp.dump.tmp"
final_dump="$backup_dir/tabi-$timestamp.dump"
compose=(docker compose --env-file .env --env-file release.env -f compose.yaml -f compose.production.yaml)

# shellcheck disable=SC2016 # Variables expand inside the container shell, not this host shell.
"${compose[@]}" exec -T postgres sh -c \
  'pg_dump --format=custom --no-owner --no-acl -U "$POSTGRES_USER" "$POSTGRES_DB"' > "$temporary_dump"
test -s "$temporary_dump"
mv "$temporary_dump" "$final_dump"
sha256sum "$final_dump" > "$final_dump.sha256"

# Local retention is deliberately short. Off-site retention is controlled by restic policy.
find "$backup_dir" -type f -name 'tabi-*.dump*' -mtime +14 -delete

if [[ -f /etc/tabi/secrets/restic_environment ]]; then
  # shellcheck disable=SC1091
  source /etc/tabi/secrets/restic_environment
  export RESTIC_PASSWORD_FILE=/etc/tabi/secrets/restic_password
  restic backup "$backup_dir" /var/lib/tabi/feed-archive /var/lib/tabi/static-artifacts \
    "$deployment_root/Caddyfile" "$deployment_root/compose.yaml" "$deployment_root/compose.production.yaml"
fi
