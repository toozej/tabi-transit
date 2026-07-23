#!/usr/bin/env bash
# Verify an archive can be inspected before an isolated restore. This is a
# preflight only: it never starts Compose or changes any database.
set -Eeuo pipefail

if [[ $# -ne 1 ]]; then
  printf 'usage: %s /path/to/tabi-YYYYMMDDTHHMMSSZ.dump\n' "$0" >&2
  exit 2
fi

dump_file=$1
[[ -r "$dump_file" && -s "$dump_file" ]] || { printf 'readable non-empty dump required\n' >&2; exit 3; }
[[ -r "$dump_file.sha256" && -s "$dump_file.sha256" ]] || { printf 'matching checksum is required\n' >&2; exit 3; }
command -v pg_restore >/dev/null 2>&1 || { printf 'restore prerequisite is unavailable: pg_restore\n' >&2; exit 3; }
(cd "$(dirname "$dump_file")" && sha256sum -c "$(basename "$dump_file").sha256" >/dev/null) || {
  printf 'restore candidate checksum verification failed\n' >&2
  exit 1
}
pg_restore --list "$dump_file" >/dev/null
printf 'restore candidate verified; use restore-postgres-isolated.sh only with a distinct restore project\n'
