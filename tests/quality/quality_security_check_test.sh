#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
checker="$repo_root/scripts/quality-security-check.sh"

bash -n "$checker"
"$checker" "$repo_root"

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT
mkdir -p "$scratch/deployment" "$scratch/apps/mobile/src"
git -C "$scratch" init -q
printf '.env\n*.pem\n*.p8\ndeployment/secrets/*\n' > "$scratch/.gitignore"
printf 'services:\n  postgres:\n    ports:\n      - "5432:5432"\n' > "$scratch/deployment/compose.yaml"
if "$checker" "$scratch" >/dev/null 2>&1; then
  printf 'expected public Postgres port fixture to fail\n' >&2
  exit 1
fi

printf 'package main\nimport "log/slog"\nfunc main() { slog.Info("bad", "' > "$scratch/apps/mobile/log.go"
printf 'token' >> "$scratch/apps/mobile/log.go"
printf '", "value") }\n' >> "$scratch/apps/mobile/log.go"
if "$checker" "$scratch" >/dev/null 2>&1; then
  printf 'expected sensitive logging fixture to fail\n' >&2
  exit 1
fi

printf 'quality/security checker tests passed\n'
