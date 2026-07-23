#!/usr/bin/env bash
set -Eeuo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
deployment_dir=$(cd -- "$script_dir/.." && pwd)
temporary_dir=$(mktemp -d)
trap 'rm -rf "$temporary_dir"' EXIT

for secret in postgres_password trimet_app_id mapbox_server_token installation_auth_key; do
  printf 'validation-placeholder\n' > "$temporary_dir/$secret"
done

cat > "$temporary_dir/env" <<EOF
COMPOSE_PROJECT_NAME=tabi-validation
TABI_API_DOMAIN=api.example.invalid
POSTGRES_DB=tabi
POSTGRES_USER=tabi
TABI_SECRETS_DIR=$temporary_dir
TABI_FEED_ARCHIVE_DIR=/tmp/tabi-validation/feed-archive
TABI_STATIC_ARTIFACT_DIR=/tmp/tabi-validation/static-artifacts
TABI_BACKEND_IMAGE=ghcr.io/example/tabi-backend@sha256:REPLACE_WITH_A_VERIFIED_DIGEST
POSTGIS_IMAGE=postgis/postgis@sha256:REPLACE_WITH_A_VERIFIED_DIGEST
CADDY_IMAGE=caddy@sha256:REPLACE_WITH_A_VERIFIED_DIGEST
EOF

rendered_config="$temporary_dir/compose.json"
docker compose --env-file "$temporary_dir/env" \
  -f "$deployment_dir/compose.yaml" \
  -f "$deployment_dir/compose.production.yaml" config --quiet
docker compose --env-file "$temporary_dir/env" \
  -f "$deployment_dir/compose.yaml" \
  -f "$deployment_dir/compose.production.yaml" config --format json > "$rendered_config"

# Caddy must be the only service publishing a host port. PostgreSQL stays private.
jq -e '
  (.services.postgres.ports // [] | length == 0) and
  ([.services | to_entries[] | select(.value.ports? and .key != "caddy")] | length == 0) and
  (.networks.backend.internal == true)
' "$rendered_config" >/dev/null
