-- name: ListNearbyStopsPerMode :many
WITH input AS (
  SELECT ST_SetSRID(ST_MakePoint(sqlc.arg(lon)::double precision, sqlc.arg(lat)::double precision), 4326)::geography AS point
), nearby AS (
  SELECT s.public_id, s.name, s.mode, s.point,
         ST_Distance(s.point, input.point)::double precision AS distance_meters,
         row_number() OVER (PARTITION BY s.mode ORDER BY s.point <-> input.point, s.public_id) AS mode_rank
  FROM transit.stops AS s
  JOIN catalog.feed_versions f ON f.id = s.feed_version_id AND f.status = 'active'
  CROSS JOIN input
  WHERE ST_DWithin(s.point, input.point, sqlc.arg(radius_meters)::double precision)
    AND (cardinality(sqlc.narg(modes)::text[]) IS NULL OR s.mode::text = ANY(sqlc.narg(modes)::text[]))
    AND (sqlc.narg(wheelchair_accessible)::boolean IS NULL
      OR (sqlc.narg(wheelchair_accessible)::boolean = true AND s.wheelchair_boarding = 1)
      OR (sqlc.narg(wheelchair_accessible)::boolean = false AND s.wheelchair_boarding <> 1))
)
SELECT public_id, name, mode, ST_X(point::geometry) AS longitude, ST_Y(point::geometry) AS latitude, distance_meters
FROM nearby
WHERE mode_rank <= sqlc.arg(limit_per_mode)::bigint
ORDER BY distance_meters, public_id
LIMIT sqlc.arg(total_limit)::integer;
