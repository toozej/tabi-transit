#!/usr/bin/env bash
set -Eeuo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
scripts="$repo_root/deployment/scripts"

for script in "$scripts"/{ops-preflight.sh,verify-backup-age.sh,verify-restore-candidate.sh,backup-postgres.sh,restore-postgres-isolated.sh}; do
  bash -n "$script"
done

scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
mkdir -p "$scratch/deploy" "$scratch/backups/postgres"

if TABI_DEPLOYMENT_ROOT="$scratch/missing" "$scripts/ops-preflight.sh" runtime >/dev/null 2>&1; then
  printf 'preflight unexpectedly accepted a missing deployment root\n' >&2
  exit 1
fi
if TABI_DEPLOYMENT_ROOT="$scratch/deploy" "$scripts/backup-postgres.sh" >/dev/null 2>&1; then
  printf 'backup unexpectedly ran without deployment prerequisites\n' >&2
  exit 1
fi
if TABI_BACKUP_DIR="$scratch/backups" "$scripts/verify-backup-age.sh" 6 >/dev/null 2>&1; then
  printf 'backup-age check unexpectedly accepted an empty backup directory\n' >&2
  exit 1
fi
if "$scripts/verify-restore-candidate.sh" "$scratch/missing.dump" >/dev/null 2>&1; then
  printf 'restore candidate check unexpectedly accepted a missing dump\n' >&2
  exit 1
fi

printf 'operations script fail-safe tests passed\n'
