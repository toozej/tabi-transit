#!/usr/bin/env bash
set -Eeuo pipefail

# Runs only local, deterministic checks. It never contacts a transit provider
# and never needs credentials or a running production-like host.
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$root"

ruby tests/contract/verify_vertical_slice.rb
ruby api/scripts/validate-openapi.rb
bash api/scripts/generate-check.sh
GOCACHE="${GOCACHE:-/tmp/tabi-go-integration-cache}" \
  go test ./internal/api ./internal/config ./services/transit-api/cmd/transit-api
corepack pnpm --dir apps/mobile test
bash deployment/scripts/validate-compose.sh
