#!/usr/bin/env bash
# Exercise deployment-script release promotion with local command doubles. This
# never parses a real environment file, starts Compose, contacts a registry, or
# opens a network connection.
set -Eeuo pipefail

repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT

if bash "$repo_root/scripts/ci/validate-release-input.sh" 'ghcr.io/example/tabi-backend:latest' >/dev/null 2>&1; then
  printf 'release input validation unexpectedly accepted a mutable tag\n' >&2
  exit 1
fi
bash "$repo_root/scripts/ci/validate-release-input.sh" \
  'ghcr.io/example/tabi-backend@sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee' \
  >/dev/null

deployment_root="$scratch/deploy"
bin_dir="$scratch/bin"
secrets_dir="$scratch/secrets"
lock_dir="$scratch/lock"
mkdir -p "$deployment_root/deployment/scripts" "$bin_dir" "$secrets_dir" "$lock_dir"
cp "$repo_root/deployment/compose.yaml" "$repo_root/deployment/compose.production.yaml" \
  "$repo_root/deployment/Caddyfile" "$deployment_root/"
cp "$repo_root/deployment/scripts/"{backup-postgres.sh,deploy.sh,ops-preflight.sh} \
  "$deployment_root/deployment/scripts/"

printf '%s\n' 'TABI_API_DOMAIN=api.example.invalid' > "$deployment_root/.env"
old_image='ghcr.io/example/tabi-backend@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
candidate_image='ghcr.io/example/tabi-backend@sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd'
printf 'TABI_BACKEND_IMAGE=%s\nPOSTGIS_IMAGE=postgis/postgis@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\nCADDY_IMAGE=caddy@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc\n' "$old_image" > "$deployment_root/release.env"
printf 'TABI_BACKEND_IMAGE=%s\nPOSTGIS_IMAGE=postgis/postgis@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\nCADDY_IMAGE=caddy@sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc\n' "$candidate_image" > "$deployment_root/release.env.new"
for secret in postgres_password database_url; do
  printf 'fixture-only\n' > "$secrets_dir/$secret"
done

cat > "$bin_dir/docker" <<'EOF'
#!/usr/bin/env bash
set -Eeuo pipefail
if [[ "$*" == *' exec -T postgres sh -c '* ]]; then
  printf 'fixture custom dump\n'
fi
EOF
cat > "$bin_dir/curl" <<'EOF'
#!/usr/bin/env bash
exit "${TABI_TEST_CURL_STATUS:-0}"
EOF
cat > "$bin_dir/sleep" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
chmod +x "$bin_dir/docker" "$bin_dir/curl" "$bin_dir/sleep"

run_deploy() {
  PATH="$bin_dir:$PATH" \
    TABI_DEPLOYMENT_ROOT="$deployment_root" \
    TABI_SECRETS_DIR="$secrets_dir" \
    TABI_LOCK_DIR="$lock_dir" \
    TABI_BACKUP_DIR="$scratch/backups" \
    "$deployment_root/deployment/scripts/deploy.sh" "$deployment_root/release.env.new"
}

if TABI_TEST_CURL_STATUS=1 run_deploy >/dev/null 2>&1; then
  printf 'deployment unexpectedly succeeded after a failed readiness check\n' >&2
  exit 1
fi
grep -Fqx "TABI_BACKEND_IMAGE=$old_image" "$deployment_root/release.env"
[[ ! -e "$deployment_root/release.previous.env" ]]

run_deploy >/dev/null
grep -Fqx "TABI_BACKEND_IMAGE=$candidate_image" "$deployment_root/release.env"
grep -Fqx "TABI_BACKEND_IMAGE=$old_image" "$deployment_root/release.previous.env"

printf 'deployment release-promotion safety checks passed\n'
