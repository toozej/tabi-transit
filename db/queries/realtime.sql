-- name: ListCurrentVehicles :many
SELECT source_id, public_id, source_vehicle_id, route_public_id, trip_public_id, mode, direction_id,
       headsign, ST_X(point::geometry)::double precision AS longitude, ST_Y(point::geometry)::double precision AS latitude,
       bearing, speed_meters_per_second, in_service, source_updated_at, entity_updated_at,
       fetched_at, processed_at, freshness_status, snapshot_id
FROM realtime.vehicle_current
WHERE (cardinality(sqlc.narg(source_ids)::text[]) IS NULL OR source_id = ANY(sqlc.narg(source_ids)::text[]))
ORDER BY public_id;

-- name: ListVehicleHistory :many
-- Observations are normalized positions, not a claim about schedule adherence.
-- A caller-enforced 30 day window keeps this read aligned with retention.
SELECT source_id, public_id, source_vehicle_id, route_public_id, trip_public_id,
       mode, ST_X(point::geometry)::double precision AS longitude,
       ST_Y(point::geometry)::double precision AS latitude, fetched_at,
       processed_at, freshness_status
FROM history.vehicle_observations
WHERE public_id=sqlc.arg(public_id)
  AND processed_at >= sqlc.arg(from_at)
  AND processed_at <= sqlc.arg(to_at)
  AND (sqlc.narg(cursor_at)::timestamptz IS NULL OR processed_at < sqlc.narg(cursor_at)::timestamptz)
ORDER BY processed_at DESC
LIMIT sqlc.arg(row_limit);

-- name: HasReadyVehicleData :one
-- Readiness deliberately requires a valid snapshot and its corresponding
-- successful source-health record. A database connection alone is not enough
-- to describe stale or absent transit data as ready.
SELECT EXISTS (
  SELECT 1
  FROM realtime.vehicle_current vehicle
  JOIN realtime.snapshots snapshot ON snapshot.id = vehicle.snapshot_id
  JOIN ops.source_health health ON health.source_id = vehicle.source_id
  WHERE snapshot.is_valid
    AND snapshot.fetched_at >= now() - interval '90 seconds'
    AND health.last_valid_snapshot_at IS NOT NULL
    AND health.last_success_at >= now() - interval '90 seconds'
);

-- name: ListStopSchedule :many
-- Calendar exceptions take precedence over weekday service. Service-day seconds
-- are deliberately returned unchanged so after-midnight trips remain correct.
WITH active AS (
  SELECT id, version_label FROM catalog.feed_versions WHERE status = 'active'
), applicable AS (
  SELECT s.feed_version_id, s.service_id
  FROM transit.services s JOIN active f ON f.id = s.feed_version_id
  LEFT JOIN transit.service_calendar_dates d ON d.feed_version_id=s.feed_version_id AND d.service_id=s.service_id AND d.service_date=sqlc.arg(service_date)::date
  LEFT JOIN transit.service_calendars c ON c.feed_version_id=s.feed_version_id AND c.service_id=s.service_id
  WHERE COALESCE(d.exception_type=1, (c.start_date <= sqlc.arg(service_date)::date AND c.end_date >= sqlc.arg(service_date)::date AND CASE EXTRACT(ISODOW FROM sqlc.arg(service_date)::date) WHEN 1 THEN c.monday WHEN 2 THEN c.tuesday WHEN 3 THEN c.wednesday WHEN 4 THEN c.thursday WHEN 5 THEN c.friday WHEN 6 THEN c.saturday ELSE c.sunday END))
    AND COALESCE(d.exception_type, 0) <> 2
)
SELECT t.public_id AS trip_id, t.route_public_id AS route_id, st.stop_public_id AS stop_id, t.direction_id, COALESCE(t.headsign,'') AS headsign,
       COALESCE(st.departure_seconds, st.arrival_seconds) AS service_day_seconds, f.version_label
FROM transit.stop_times st
JOIN transit.trips t ON t.feed_version_id=st.feed_version_id AND t.public_id=st.trip_public_id
JOIN applicable a ON a.feed_version_id=t.feed_version_id AND a.service_id=t.service_id
JOIN active f ON f.id=t.feed_version_id
WHERE st.stop_public_id=sqlc.arg(stop_id)
  AND (sqlc.arg(route_id)::text='' OR t.route_public_id=sqlc.arg(route_id))
  AND (sqlc.narg(direction_id)::smallint IS NULL OR t.direction_id=sqlc.narg(direction_id)::smallint)
  AND (sqlc.arg(cursor)::text='' OR concat(t.public_id, ':', st.stop_sequence) > sqlc.arg(cursor))
ORDER BY COALESCE(st.departure_seconds, st.arrival_seconds), t.public_id, st.stop_sequence
LIMIT sqlc.arg(row_limit);

-- name: StopFeedTimezone :one
SELECT f.service_timezone
FROM transit.stops s
JOIN catalog.feed_versions f ON f.id=s.feed_version_id
WHERE s.public_id=sqlc.arg(stop_id) AND f.status='active' AND f.service_timezone IS NOT NULL
LIMIT 1;

-- name: ListStopArrivalCandidates :many
-- Current updates are optional enrichment. Static times are still returned
-- when no matching valid realtime entity exists.
WITH active AS (
  SELECT id, version_label FROM catalog.feed_versions WHERE status='active'
), applicable AS (
  SELECT s.feed_version_id, s.service_id
  FROM transit.services s JOIN active f ON f.id=s.feed_version_id
  LEFT JOIN transit.service_calendar_dates d ON d.feed_version_id=s.feed_version_id AND d.service_id=s.service_id AND d.service_date=sqlc.arg(service_date)::date
  LEFT JOIN transit.service_calendars c ON c.feed_version_id=s.feed_version_id AND c.service_id=s.service_id
  WHERE COALESCE(d.exception_type=1, (c.start_date <= sqlc.arg(service_date)::date AND c.end_date >= sqlc.arg(service_date)::date AND CASE EXTRACT(ISODOW FROM sqlc.arg(service_date)::date) WHEN 1 THEN c.monday WHEN 2 THEN c.tuesday WHEN 3 THEN c.wednesday WHEN 4 THEN c.thursday WHEN 5 THEN c.friday WHEN 6 THEN c.saturday ELSE c.sunday END))
    AND COALESCE(d.exception_type, 0) <> 2
), updates AS (
 SELECT DISTINCT ON (tu.trip_public_id, su.stop_sequence) tu.trip_public_id, su.stop_sequence, tu.schedule_relationship AS trip_relationship,
   su.arrival_delay_seconds, su.arrival_time, su.departure_delay_seconds, su.departure_time, su.schedule_relationship AS stop_relationship,
   tu.source_id, tu.source_updated_at, tu.fetched_at, tu.processed_at
 FROM realtime.trip_updates_current tu
 JOIN realtime.trip_update_stop_times_current su ON su.source_id=tu.source_id AND su.entity_id=tu.entity_id
 ORDER BY tu.trip_public_id, su.stop_sequence, tu.processed_at DESC
)
SELECT t.public_id AS trip_id, t.route_public_id AS route_id, st.stop_public_id AS stop_id, st.stop_sequence, t.direction_id, COALESCE(t.headsign,'') AS headsign,
 COALESCE(st.departure_seconds, st.arrival_seconds) AS service_day_seconds, f.version_label,
 u.trip_relationship, u.stop_relationship, u.arrival_delay_seconds, u.arrival_time, u.departure_delay_seconds, u.departure_time,
 u.source_id AS realtime_source, u.source_updated_at, u.fetched_at AS realtime_fetched_at, u.processed_at AS realtime_processed_at
FROM transit.stop_times st
JOIN transit.trips t ON t.feed_version_id=st.feed_version_id AND t.public_id=st.trip_public_id
JOIN applicable a ON a.feed_version_id=t.feed_version_id AND a.service_id=t.service_id
JOIN active f ON f.id=t.feed_version_id
LEFT JOIN updates u ON u.trip_public_id=t.public_id AND u.stop_sequence=st.stop_sequence
WHERE st.stop_public_id=sqlc.arg(stop_id)
 AND (sqlc.arg(route_id)::text='' OR t.route_public_id=sqlc.arg(route_id))
 AND (sqlc.narg(direction_id)::smallint IS NULL OR t.direction_id=sqlc.narg(direction_id)::smallint)
ORDER BY COALESCE(st.departure_seconds, st.arrival_seconds), t.public_id, st.stop_sequence;

-- name: ListCurrentAlerts :many
SELECT source_id, entity_id, cause, effect, COALESCE(header_text,'') AS header_text,
       COALESCE(description_text,'') AS description_text, COALESCE(url,'') AS url,
       active_from, active_until, fetched_at, processed_at
FROM realtime.alerts_current
WHERE (sqlc.arg(active_only)::boolean = false OR (active_from IS NULL OR active_from <= sqlc.arg(now_at)) AND (active_until IS NULL OR active_until >= sqlc.arg(now_at)))
  AND (sqlc.arg(effect)::text='' OR effect=sqlc.arg(effect))
  AND (sqlc.narg(updated_since)::timestamptz IS NULL OR processed_at >= sqlc.narg(updated_since))
  AND (concat(source_id, ':alert:', entity_id) > sqlc.arg(cursor)::text)
ORDER BY processed_at DESC, source_id, entity_id
LIMIT sqlc.arg(row_limit);

-- name: GetCurrentAlert :one
SELECT source_id, entity_id, cause, effect, COALESCE(header_text,'') AS header_text,
       COALESCE(description_text,'') AS description_text, COALESCE(url,'') AS url,
       active_from, active_until, fetched_at, processed_at
FROM realtime.alerts_current
WHERE source_id=sqlc.arg(source_id) AND entity_id=sqlc.arg(entity_id);
