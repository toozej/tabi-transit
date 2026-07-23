#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
out="$root/../packages/api-client/src/generated/openapi.ts"
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
if [ ! -f "$out" ]; then
  printf 'Generated API client is missing: %s\n' "$out" >&2
  exit 1
fi
cd "$root/.."
corepack pnpm exec openapi-typescript "$root/openapi.yaml" -o "$tmp"
corepack pnpm exec prettier --parser typescript --write "$tmp"
diff -u "$out" "$tmp"
