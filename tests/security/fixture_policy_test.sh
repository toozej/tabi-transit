#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
checker="$repo_root/scripts/security/verify-fixture-policy.sh"

bash -n "$checker"
bash "$checker" "$repo_root"

token_scratch="$(mktemp -d)"
search_scratch="$(mktemp -d)"
trap 'rm -rf "$token_scratch" "$search_scratch"' EXIT
mkdir -p "$token_scratch/tests/fixtures" "$search_scratch/tests/fixtures"
printf '%s\n' 'token=pk.example_public_token_1234567890' > "$token_scratch/tests/fixtures/token.txt"
if bash "$checker" "$token_scratch" >/dev/null 2>&1; then
  printf 'expected Mapbox token fixture to fail\n' >&2
  exit 1
fi

printf '%s\n' '{"searchText":"home address"}' > "$search_scratch/tests/fixtures/search.json"
if bash "$checker" "$search_scratch" >/dev/null 2>&1; then
  printf 'expected search-text fixture to fail\n' >&2
  exit 1
fi

printf 'fixture policy checker tests passed\n'
