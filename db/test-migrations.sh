#!/usr/bin/env bash
set -euo pipefail

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
image=${POSTGIS_TEST_IMAGE:-postgis/postgis:17-3.5}
container_name="tabi-postgis-wp03-$$"
port=${POSTGIS_TEST_PORT:-55432}

cleanup() { docker rm -f "$container_name" >/dev/null 2>&1 || true; }
trap cleanup EXIT
command -v docker >/dev/null || { printf '%s\n' 'Docker is required for PostGIS integration tests.' >&2; exit 127; }

docker run --rm -d --name "$container_name" \
  -e POSTGRES_PASSWORD=tabi_test -e POSTGRES_USER=tabi_test -e POSTGRES_DB=tabi_test \
  -v "$root_dir:/workspace:ro" "$image" >/dev/null
database_url="postgres://tabi_test:tabi_test@127.0.0.1:5432/tabi_test?sslmode=disable"
psql_in_container() { docker exec "$container_name" psql "$database_url" "$@"; }
for _ in $(seq 1 30); do
  if psql_in_container -Atqc 'SELECT 1' >/dev/null 2>&1; then break; fi
  sleep 1
done
psql_in_container -Atqc 'SELECT PostGIS_Version()' >/dev/null
docker exec -e "TABI_DATABASE_URL=$database_url" "$container_name" bash /workspace/db/migrate.sh
docker exec -e "TABI_DATABASE_URL=$database_url" "$container_name" bash /workspace/db/migrate.sh
psql_in_container -v ON_ERROR_STOP=1 -f /workspace/db/fixtures/representative.sql
actual=$(psql_in_container -At -v ON_ERROR_STOP=1 -c "WITH input AS (SELECT ST_SetSRID(ST_MakePoint(-122.6700,45.5200),4326)::geography AS point), nearby AS (SELECT s.public_id, s.mode, row_number() OVER (PARTITION BY s.mode ORDER BY s.point <-> input.point,s.public_id) AS mode_rank FROM transit.stops s CROSS JOIN input WHERE ST_DWithin(s.point,input.point,500)) SELECT mode || ':' || public_id FROM nearby WHERE mode_rank <= 1 ORDER BY mode, public_id")
expected=$'bus:fixture:stop:bus-nearest\nlight_rail:fixture:stop:rail-nearest'
[[ "$actual" == "$expected" ]] || { printf 'unexpected nearby result:\n%s\n' "$actual" >&2; exit 1; }
psql_in_container -v ON_ERROR_STOP=1 -c "EXPLAIN (ANALYZE, BUFFERS, COSTS OFF) WITH input AS (SELECT ST_SetSRID(ST_MakePoint(-122.6700,45.5200),4326)::geography AS point) SELECT s.public_id FROM transit.stops s CROSS JOIN input WHERE s.feed_version_id=(SELECT id FROM catalog.feed_versions WHERE status='active') AND ST_DWithin(s.point,input.point,500) ORDER BY s.point <-> input.point,s.public_id LIMIT 10"
printf '%s\n' 'WP-03 PostGIS migration and nearby/limitPerMode integration test passed.'
