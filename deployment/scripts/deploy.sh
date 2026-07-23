#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

if [[ $# -ne 1 ]]; then
  echo "usage: $0 /path/to/release.env.new" >&2
  exit 2
fi

deployment_root=${TABI_DEPLOYMENT_ROOT:-/opt/tabi}
cd "$deployment_root"
exec 9>/run/lock/tabi-deploy.lock
flock -n 9 || { echo "another deployment is running" >&2; exit 1; }

candidate_release=$(realpath -- "$1")
test -s "$candidate_release"
candidate=(docker compose --env-file .env --env-file "$candidate_release" -f compose.yaml -f compose.production.yaml)
"${candidate[@]}" config --quiet

./deployment/scripts/backup-postgres.sh
"${candidate[@]}" pull
"${candidate[@]}" --profile jobs run --rm migrate

[[ -f release.env ]] && cp -f release.env release.previous.env
install -m 0600 "$candidate_release" release.env
active=(docker compose --env-file .env --env-file release.env -f compose.yaml -f compose.production.yaml)
"${active[@]}" up -d --remove-orphans

domain=$(sed -n 's/^TABI_API_DOMAIN=//p' .env | tail -n 1)
[[ -n "$domain" ]] || { echo "TABI_API_DOMAIN missing from .env" >&2; exit 1; }
for _ in $(seq 1 30); do
  if curl --fail --silent --show-error "https://$domain/health/ready" >/dev/null; then
    exit 0
  fi
  sleep 5
done

echo "health check failed; returning application containers to previous image set" >&2
if [[ -f release.previous.env ]]; then
  rollback=(docker compose --env-file .env --env-file release.previous.env -f compose.yaml -f compose.production.yaml)
  "${rollback[@]}" up -d --remove-orphans
fi
exit 1
