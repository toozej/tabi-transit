#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
caddyfile="$repo_root/deployment/Caddyfile"
compose_file="$repo_root/deployment/compose.yaml"

grep -Fq 'Content-Security-Policy-Report-Only' "$caddyfile"
grep -Fq "frame-ancestors 'none'" "$caddyfile"
grep -Fq 'root * /srv/tabi-web' "$caddyfile"
grep -Fq 'Cache-Control "public, max-age=31536000, immutable"' "$caddyfile"
grep -Fq 'TABI_WEB_ASSET_DIR' "$compose_file"

# Development CORS can be allowlisted, but must never grant every origin.
if rg -n "Access-Control-Allow-Origin.*[\"']\\*[\"']" "$repo_root/internal"; then
  printf 'wildcard CORS origin found\n' >&2
  exit 1
fi

printf 'web deployment policy tests passed\n'
