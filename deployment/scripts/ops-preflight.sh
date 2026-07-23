#!/usr/bin/env bash
# Validate only the local deployment shape. This script never starts Compose,
# contacts a provider, reads a secret value, or prints configuration contents.
set -Eeuo pipefail

usage() {
  printf 'usage: %s {runtime|backup|restore}\n' "$0" >&2
  exit 2
}

[[ $# -eq 1 ]] || usage
case "$1" in runtime|backup|restore) ;; *) usage ;; esac

deployment_root=${TABI_DEPLOYMENT_ROOT:-/opt/tabi}
[[ -d "$deployment_root" ]] || { printf 'deployment root is unavailable: %s\n' "$deployment_root" >&2; exit 3; }

for required in .env release.env compose.yaml compose.production.yaml Caddyfile; do
  [[ -r "$deployment_root/$required" ]] || {
    printf 'required deployment file is unavailable: %s\n' "$required" >&2
    exit 3
  }
done

case "$1" in
  backup)
    secrets_dir=${TABI_SECRETS_DIR:-/etc/tabi/secrets}
    for required in postgres_password database_url; do
      [[ -r "$secrets_dir/$required" ]] || {
        printf 'backup prerequisite secret is unavailable: %s\n' "$required" >&2
        exit 3
      }
    done
    ;;
  restore)
    command -v pg_restore >/dev/null 2>&1 || {
      printf 'restore prerequisite is unavailable: pg_restore\n' >&2
      exit 3
    }
    ;;
esac

printf 'operations preflight passed for %s\n' "$1"
