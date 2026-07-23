#!/usr/bin/env bash
# Report only verified local dump freshness. It intentionally does not claim an
# off-site backup exists unless its required restic configuration is present.
set -Eeuo pipefail

max_age_hours=${1:-6}
[[ "$max_age_hours" =~ ^[1-9][0-9]*$ ]] || { printf 'max age must be a positive integer hour count\n' >&2; exit 2; }

backup_root=${TABI_BACKUP_DIR:-/var/lib/tabi/backups}/postgres
[[ -d "$backup_root" ]] || { printf 'backup directory is unavailable: %s\n' "$backup_root" >&2; exit 3; }

latest_dump=$(find "$backup_root" -maxdepth 1 -type f -name 'tabi-*.dump' -printf '%T@ %p\n' 2>/dev/null | sort -nr | head -n1 || true)
[[ -n "$latest_dump" ]] || { printf 'no local PostgreSQL dump is available\n' >&2; exit 3; }
dump_file=${latest_dump#* }
[[ -s "$dump_file" && -s "$dump_file.sha256" ]] || { printf 'dump or checksum is incomplete: %s\n' "$dump_file" >&2; exit 3; }
(cd "$(dirname "$dump_file")" && sha256sum -c "$(basename "$dump_file").sha256" >/dev/null) || {
  printf 'backup checksum verification failed: %s\n' "$dump_file" >&2
  exit 1
}

now=$(date -u +%s)
modified=$(stat -c %Y "$dump_file")
age_seconds=$((now - modified))
(( age_seconds >= 0 )) || { printf 'backup timestamp is in the future\n' >&2; exit 1; }
(( age_seconds <= max_age_hours * 3600 )) || {
  printf 'backup is older than %s hour(s): %s seconds\n' "$max_age_hours" "$age_seconds" >&2
  exit 1
}

printf 'verified local backup age: %s seconds\n' "$age_seconds"
if [[ ! -r /etc/tabi/secrets/restic_environment || ! -r /etc/tabi/secrets/restic_password ]]; then
  printf 'off-site backup status is unverified: required restic secret files are unavailable\n' >&2
  exit 3
fi
printf 'off-site backup credentials are present; run restic-check.sh for repository verification\n'
