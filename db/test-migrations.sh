#!/usr/bin/env bash
set -euo pipefail

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
image=${POSTGIS_TEST_IMAGE:-postgis/postgis:17-3.5}
container_name="tabi-postgis-wp03-$$"
port=${POSTGIS_TEST_PORT:-55432}

cleanup() { docker rm -f "$container_name" >/dev/null 2>&1 || true; }
trap cleanup EXIT
command -v docker >/dev/null || { printf '%s\n' 'Docker is required for PostGIS integration tests.' >&2; exit 127; }

# Expose only migrations and sanitized fixtures to the disposable database;
# ignored environment files and the rest of the repository are not mounted.
docker run --rm -d --name "$container_name" \
  -e POSTGRES_PASSWORD=tabi_test -e POSTGRES_USER=tabi_test -e POSTGRES_DB=tabi_test \
  -v "$root_dir/db:/workspace/db:ro" "$image" >/dev/null
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
routes=$(psql_in_container -At -v ON_ERROR_STOP=1 -c "SELECT public_id || ':' || short_name FROM transit.routes WHERE feed_version_id=(SELECT id FROM catalog.feed_versions WHERE source_id='fixture-gtfs' AND status='active') ORDER BY public_id")
[[ "$routes" == 'fixture:route:20:20' ]] || { printf 'unexpected catalog route result:\n%s\n' "$routes" >&2; exit 1; }
stop_routes=$(psql_in_container -At -v ON_ERROR_STOP=1 -c "SELECT st.stop_public_id || ':' || trip.route_public_id FROM transit.stop_times st JOIN transit.trips trip ON trip.feed_version_id=st.feed_version_id AND trip.public_id=st.trip_public_id ORDER BY st.stop_public_id")
[[ "$stop_routes" == 'fixture:stop:bus-nearest:fixture:route:20' ]] || { printf 'unexpected catalog stop route result:\n%s\n' "$stop_routes" >&2; exit 1; }
service_day=$(psql_in_container -At -v ON_ERROR_STOP=1 -c "SELECT departure_seconds FROM transit.stop_times WHERE trip_public_id='fixture:trip:20-1'")
[[ "$service_day" == '90930' ]] || { printf 'expected after-midnight GTFS service-day time, got: %s\n' "$service_day" >&2; exit 1; }
exception=$(psql_in_container -At -v ON_ERROR_STOP=1 -c "SELECT exception_type FROM transit.service_calendar_dates WHERE service_id='weekday' AND service_date='2026-07-04'")
[[ "$exception" == '2' ]] || { printf 'expected service-calendar exception, got: %s\n' "$exception" >&2; exit 1; }
ready=$(psql_in_container -At -v ON_ERROR_STOP=1 -c "SELECT EXISTS (SELECT 1 FROM realtime.vehicle_current vehicle JOIN realtime.snapshots snapshot ON snapshot.id=vehicle.snapshot_id JOIN ops.source_health health ON health.source_id=vehicle.source_id WHERE snapshot.is_valid AND health.last_valid_snapshot_at IS NOT NULL)")
[[ "$ready" == 't' ]] || { printf '%s\n' 'expected valid realtime snapshot readiness' >&2; exit 1; }
psql_in_container -v ON_ERROR_STOP=1 -c "EXPLAIN (ANALYZE, BUFFERS, COSTS OFF) WITH input AS (SELECT ST_SetSRID(ST_MakePoint(-122.6700,45.5200),4326)::geography AS point) SELECT s.public_id FROM transit.stops s CROSS JOIN input WHERE s.feed_version_id=(SELECT id FROM catalog.feed_versions WHERE status='active') AND ST_DWithin(s.point,input.point,500) ORDER BY s.point <-> input.point,s.public_id LIMIT 10"
printf '%s\n' 'WP-03 PostGIS migration and nearby/limitPerMode integration test passed.'
