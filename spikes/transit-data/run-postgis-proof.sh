#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
project_name=tabi_transit_spike

cleanup() {
  docker compose --project-name "$project_name" -f "$script_dir/postgis-compose.yaml" down --volumes --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker compose --project-name "$project_name" -f "$script_dir/postgis-compose.yaml" up --detach --wait
for attempt in {1..30}; do
  if docker compose --project-name "$project_name" -f "$script_dir/postgis-compose.yaml" exec --no-TTY postgis \
    psql --username tabi_spike --dbname tabi_spike --quiet --command 'SELECT 1' >/dev/null 2>&1; then
    break
  fi
  if [[ "$attempt" == 30 ]]; then
    echo 'PostGIS did not become query-ready after initialization.' >&2
    exit 1
  fi
  sleep 1
done
actual=$(docker compose --project-name "$project_name" -f "$script_dir/postgis-compose.yaml" exec --no-TTY postgis \
  psql --username tabi_spike --dbname tabi_spike --tuples-only --no-align --field-separator '|' \
  --file /dev/stdin < "$script_dir/sql/nearby_limit_per_mode.sql")
expected=$'bus-nearest|bus|0|1\nrail-nearest|light_rail|111|1'
if [[ "$actual" != "$expected" ]]; then
  printf 'unexpected PostGIS limitPerMode result:\n%s\n' "$actual" >&2
  exit 1
fi
printf 'PostGIS nearby + limitPerMode proof passed:\n%s\n' "$actual"
