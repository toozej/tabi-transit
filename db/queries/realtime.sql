-- name: ListCurrentVehicles :many
SELECT public_id, source_vehicle_id, route_public_id, trip_public_id, mode, direction_id,
       headsign, ST_X(point::geometry) AS longitude, ST_Y(point::geometry) AS latitude,
       bearing, speed_meters_per_second, in_service, source_updated_at, entity_updated_at,
       fetched_at, processed_at, freshness_status, snapshot_id
FROM realtime.vehicle_current
WHERE (cardinality(sqlc.narg(source_ids)::text[]) IS NULL OR source_id = ANY(sqlc.narg(source_ids)::text[]))
ORDER BY public_id;
