#!/usr/bin/env sh
set -eu
root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
out=${1:-"$root/../packages/api-client/src/generated/openapi.ts"}
mkdir -p "$(dirname -- "$out")"
cd "$root/.."
corepack pnpm exec openapi-typescript "$root/openapi.yaml" -o "$out"
corepack pnpm exec prettier --write "$out"
