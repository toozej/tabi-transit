#!/usr/bin/env sh
set -eu
root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
ruby "$root/scripts/validate-openapi.rb" "$root/openapi.yaml"
cd "$root/.."
corepack pnpm exec redocly lint --config "$root/redocly.yaml" "$root/openapi.yaml"
