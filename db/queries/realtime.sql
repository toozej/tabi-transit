-- name: ListCurrentVehicles :many
SELECT source_id, public_id, source_vehicle_id, route_public_id, trip_public_id, mode, direction_id,
       headsign, ST_X(point::geometry)::double precision AS longitude, ST_Y(point::geometry)::double precision AS latitude,
       bearing, speed_meters_per_second, in_service, source_updated_at, entity_updated_at,
       fetched_at, processed_at, freshness_status, snapshot_id
FROM realtime.vehicle_current
WHERE (cardinality(sqlc.narg(source_ids)::text[]) IS NULL OR source_id = ANY(sqlc.narg(source_ids)::text[]))
ORDER BY public_id;

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
    AND health.last_valid_snapshot_at IS NOT NULL
);
