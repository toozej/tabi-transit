#!/usr/bin/env bash
# Dependency-free repository policy checks.  This is intentionally conservative:
# it flags committed/executable material, not missing future implementation work.
set -euo pipefail

repo_root="${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
if [[ ! -d "$repo_root" ]]; then
  printf 'error: repository path does not exist: %s\n' "$repo_root" >&2
  exit 2
fi

failures=0
warnings=0

violation() {
  printf 'VIOLATION: %s\n' "$*" >&2
  failures=$((failures + 1))
}

gate() {
  printf 'GATE: %s\n' "$*" >&2
  warnings=$((warnings + 1))
}

# Source, checked-in fixture, and deployment areas only.  Documentation and
# example environment files intentionally describe secret handling and are not
# evidence of a leaked credential.
scan_paths=(api apps db deployment internal packages services spikes tests scripts .github)
scan_files=()
for path in "${scan_paths[@]}"; do
  [[ -d "$repo_root/$path" ]] || continue
  while IFS= read -r -d '' file; do
    case "$file" in
      */node_modules/*|*/.cache/*|*/dist/*|*/build/*|*/coverage/*|*.example|*.template)
        continue
        ;;
    esac
    scan_files+=("$file")
  done < <(find "$repo_root/$path" -type f -print0)
done

scan_for() {
  local label="$1"
  local expression="$2"
  local match
  for match in "${scan_files[@]}"; do
    if grep -nE -I -- "$expression" "$match" >/dev/null 2>&1; then
      grep -nE -I -- "$expression" "$match" | while IFS= read -r line; do
        violation "$label: ${match#"$repo_root/"}:$line"
      done
    fi
  done
}

# These patterns require a recognisable production credential prefix or private
# key header; placeholders such as `change-me` are deliberately not reported.
scan_for "private key material" '-----BEGIN ([A-Z0-9 ]+ )?PRIVATE KEY-----'
scan_for "GitHub token" 'gh[pousr]_[A-Za-z0-9_]{20,}|github_pat_[A-Za-z0-9_]{20,}'
scan_for "AWS access key" 'AKIA[0-9A-Z]{16}'
scan_for "Stripe live secret" 'sk_live_[A-Za-z0-9]{16,}'
scan_for "Mapbox secret token" 'sk\.[A-Za-z0-9_-]{16,}'

# Raw location must not be emitted in application logs.  The patterns are tied
# to logging calls, so coordinate types and GeoJSON transformations are valid.
scan_for "raw coordinate logging" '(console\.(log|info|warn|error)|slog\.(Debug|Info|Warn|Error)|log\.(Print|Printf|Println)).*(latitude|longitude|coordinate|\blat\b|\blon\b)'

compose_file="$repo_root/deployment/compose.yaml"
if [[ -f "$compose_file" ]]; then
  if grep -nE '/var/run/docker\.sock|/run/docker\.sock' "$compose_file" >/dev/null; then
    while IFS= read -r line; do violation "public Docker socket mount: deployment/compose.yaml:$line"; done < <(grep -nE '/var/run/docker\.sock|/run/docker\.sock' "$compose_file")
  fi
  # A service named postgres/db/database must never publish a host port.  This
  # avoids interpreting `expose` as public publication.
  while IFS= read -r line; do
    violation "public Postgres/admin port: deployment/compose.yaml:$line"
  done < <(awk '
    /^  (postgres|db|database|admin|metrics):[[:space:]]*$/ { service=$1; sub(":", "", service); watched=1; next }
    /^  [A-Za-z0-9_-]+:[[:space:]]*$/ { watched=0 }
    watched && /^[[:space:]]+ports:[[:space:]]*$/ { ports=1; next }
    watched && ports && /^[[:space:]]+-[[:space:]]*["'"'"']?[0-9.]*:?(5432|5433|9187|3000)[^"'"'"']*["'"'"']?[[:space:]]*$/ { print NR ":" $0 }
    watched && ports && /^[[:space:]]+[A-Za-z][A-Za-z0-9_-]*:/ { ports=0 }
  ' "$compose_file")
else
  gate "deployment Compose configuration is not present; public-port checks are deferred"
fi

if [[ -d "$repo_root/apps/mobile/src" ]] && grep -R -I -q '@rnmapbox/maps' "$repo_root/apps/mobile/src"; then
  if ! grep -R -I -E -q 'FlatList|SectionList|accessibilityRole|accessibilityLabel' "$repo_root/apps/mobile/src"; then
    violation "map-only core flow: RNMapbox is present without an accessible list or semantic control in mobile source"
  fi
fi

if command -v git >/dev/null 2>&1 && [[ -d "$repo_root/.git" ]]; then
  for sensitive_path in .env deployment/secrets/quality-security-probe.pem signing-key.p8; do
    if ! git -C "$repo_root" check-ignore -q -- "$sensitive_path"; then
      violation "sensitive path is not ignored: $sensitive_path"
    fi
  done
else
  gate "Git metadata unavailable; sensitive-path ignore verification is deferred"
fi

if [[ -d "$repo_root/apps/mobile" && ! -f "$repo_root/apps/mobile/.env.example" ]]; then
  gate "mobile safe environment example is absent; token configuration review is deferred"
fi

if (( failures > 0 )); then
  printf 'quality/security policy check failed: %d violation(s), %d gate(s)\n' "$failures" "$warnings" >&2
  exit 1
fi
printf 'quality/security policy check passed: %d gate(s)\n' "$warnings"
