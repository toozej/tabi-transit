#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

if [[ $# -ne 1 ]]; then
  echo "usage: $0 /path/to/release.env.new" >&2
  exit 2
fi

deployment_root=${TABI_DEPLOYMENT_ROOT:-/opt/tabi}
cd "$deployment_root"
lock_dir=${TABI_LOCK_DIR:-/run/lock}
exec 9>"$lock_dir/tabi-deploy.lock"
flock -n 9 || { echo "another deployment is running" >&2; exit 1; }

candidate_release=$(realpath -- "$1")
test -s "$candidate_release"
[[ "$candidate_release" != "$deployment_root/release.env" ]] || {
  echo "candidate release file must not replace the active release.env in place" >&2
  exit 2
}

# A release file is data, not a shell program. Require every runtime image to be
# an immutable SHA-256 reference before Compose is allowed to interpolate it.
require_immutable_image() {
  local key=$1 value count
  count=$(grep -Ec "^${key}=" "$candidate_release" || true)
  [[ $count -eq 1 ]] || { echo "release file must contain exactly one $key" >&2; exit 2; }
  value=$(sed -n "s/^${key}=//p" "$candidate_release")
  [[ $value =~ ^[A-Za-z0-9][A-Za-z0-9./:_-]*@sha256:[a-f0-9]{64}$ ]] || {
    echo "$key must be an immutable lowercase SHA-256 image reference" >&2
    exit 2
  }
}
require_immutable_image TABI_BACKEND_IMAGE
require_immutable_image POSTGIS_IMAGE
require_immutable_image CADDY_IMAGE

candidate=(docker compose --env-file .env --env-file "$candidate_release" -f compose.yaml -f compose.production.yaml)
"${candidate[@]}" config --quiet

./deployment/scripts/backup-postgres.sh
"${candidate[@]}" pull
"${candidate[@]}" --profile jobs run --rm migrate

# Keep the active release file untouched until the candidate is both migrated
# and healthy.  Compose can use the candidate directly, which lets a failed
# health check restore the prior image set without leaving release.env pointing
# at an image that was rejected.
"${candidate[@]}" up -d --remove-orphans

domain=$(sed -n 's/^TABI_API_DOMAIN=//p' .env | tail -n 1)
[[ -n "$domain" ]] || { echo "TABI_API_DOMAIN missing from .env" >&2; exit 1; }
for _ in $(seq 1 30); do
  if curl --fail --silent --show-error "https://$domain/health/ready" >/dev/null; then
    [[ -f release.env ]] && cp -f release.env release.previous.env
    install -m 0600 "$candidate_release" release.env
    exit 0
  fi
  sleep 5
done

echo "health check failed; returning application containers to previous image set" >&2
if [[ -f release.env ]]; then
  rollback=(docker compose --env-file .env --env-file release.env -f compose.yaml -f compose.production.yaml)
  "${rollback[@]}" up -d --remove-orphans
fi
exit 1
