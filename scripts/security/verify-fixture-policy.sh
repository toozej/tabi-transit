#!/usr/bin/env bash
# Reject credentials and production-looking user data from deterministic fixtures.
# This scanner is intentionally dependency-free and only inspects checked-in
# fixture directories; it never reads local environment files.
set -euo pipefail

repo_root="${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
if [[ ! -d "$repo_root" ]]; then
  printf 'error: repository path does not exist: %s\n' "$repo_root" >&2
  exit 2
fi

fixture_roots=(
  "$repo_root/tests/fixtures"
  "$repo_root/db/fixtures"
  "$repo_root/spikes/transit-data/fixtures"
)
fixture_files=()
for root in "${fixture_roots[@]}"; do
  [[ -d "$root" ]] || continue
  while IFS= read -r -d '' file; do
    fixture_files+=("$file")
  done < <(find "$root" -type f -print0)
done

failures=0
violation() {
  printf 'VIOLATION: %s\n' "$*" >&2
  failures=$((failures + 1))
}

scan_for() {
  local label="$1"
  local expression="$2"
  local file
  for file in "${fixture_files[@]}"; do
    if grep -nE -I -- "$expression" "$file" >/dev/null 2>&1; then
      while IFS= read -r line; do
        violation "$label: ${file#"$repo_root/"}:$line"
      done < <(grep -nE -I -- "$expression" "$file")
    fi
  done
}

# Public Mapbox tokens also do not belong in fixtures: their account and usage
# are real even though they are designed to be embedded in a native client.
scan_for "private key material" '-----BEGIN ([A-Z0-9 ]+ )?PRIVATE KEY-----'
scan_for "GitHub token" 'gh[pousr]_[A-Za-z0-9_]{20,}|github_pat_[A-Za-z0-9_]{20,}'
scan_for "AWS access key" 'AKIA[0-9A-Z]{16}'
scan_for "Mapbox token" '(pk|sk)\.[A-Za-z0-9_-]{16,}'
scan_for "TriMet AppID assignment" '([Tt][Rr][Ii][Mm][Ee][Tt][_ -]?([Aa][Pp][Pp])?[Ii][Dd]|[Aa][Pp][Pp][Ii][Dd])[[:space:]]*[:=][[:space:]]*[A-Za-z0-9_-]{12,}'
scan_for "push token" '(ExponentPushToken|ExpoPushToken)\[[^]]+\]'

# Full free-text searches and precise device locations are user data. Fixture
# schemas may contain transit coordinates, so this targets explicit user/query
# labels rather than valid GTFS or GeoJSON geometry.
scan_for "user search text" '"(search(Text)?|query)"[[:space:]]*:[[:space:]]*"[^"]{9,}"'
scan_for "precise user coordinate fixture" '"(user|device|current)(Location|Coordinate)?"[[:space:]]*:[[:space:]]*\[[[:space:]]*-?[0-9]{1,3}\.[0-9]{4,}'

if (( failures > 0 )); then
  printf 'fixture policy check failed: %d violation(s)\n' "$failures" >&2
  exit 1
fi
printf 'fixture policy check passed: %d file(s) inspected\n' "${#fixture_files[@]}"
