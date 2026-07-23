-- name: GetActiveFeedVersion :one
SELECT id, source_id, version_label, archive_sha256, activated_at
FROM catalog.feed_versions
WHERE source_id = sqlc.arg(source_id) AND status = 'active';

-- name: LatestActiveFeedVersionLabel :one
-- The public static endpoints currently combine enabled active feeds. The
-- newest activation is the response version marker until the API exposes a
-- multi-source static manifest.
SELECT version_label
FROM catalog.feed_versions
WHERE status = 'active'
ORDER BY activated_at DESC, id DESC
LIMIT 1;

-- name: ListRoutes :many
WITH active_feed_versions AS (
  SELECT id FROM catalog.feed_versions WHERE status = 'active'
)
SELECT r.public_id, r.mode, r.short_name, r.long_name, COALESCE(r.color::text, '') AS color,
       COALESCE(r.text_color::text, '') AS text_color
FROM transit.routes r
JOIN active_feed_versions a ON a.id = r.feed_version_id
WHERE (cardinality(sqlc.narg(modes)::text[]) IS NULL OR r.mode::text = ANY(sqlc.narg(modes)::text[]))
  AND (sqlc.arg(query)::text = '' OR lower(concat_ws(' ', r.public_id, r.short_name, r.long_name)) LIKE '%' || lower(sqlc.arg(query)::text) || '%')
  AND (sqlc.arg(cursor)::text = '' OR r.public_id > sqlc.arg(cursor)::text)
ORDER BY r.public_id
LIMIT sqlc.arg(row_limit);

-- name: ListStops :many
WITH active_feed_versions AS (
  SELECT id FROM catalog.feed_versions WHERE status = 'active'
), matching_stops AS (
  SELECT s.feed_version_id, s.public_id, s.name, s.mode, s.parent_public_id,
         s.wheelchair_boarding,
         ST_X(s.point::geometry)::double precision AS longitude,
         ST_Y(s.point::geometry)::double precision AS latitude
  FROM transit.stops s
  JOIN active_feed_versions a ON a.id = s.feed_version_id
  WHERE (cardinality(sqlc.narg(modes)::text[]) IS NULL OR s.mode::text = ANY(sqlc.narg(modes)::text[]))
    AND lower(concat_ws(' ', s.public_id, s.name)) LIKE '%' || lower(sqlc.arg(query)::text) || '%'
    AND (sqlc.arg(cursor)::text = '' OR s.public_id > sqlc.arg(cursor)::text)
  ORDER BY s.public_id
  LIMIT sqlc.arg(row_limit)
)
SELECT s.public_id, s.name, s.mode, s.parent_public_id, s.wheelchair_boarding,
       s.longitude, s.latitude,
       COALESCE(jsonb_agg(DISTINCT trip.route_public_id) FILTER (WHERE trip.route_public_id IS NOT NULL), '[]'::jsonb) AS route_public_ids
FROM matching_stops s
LEFT JOIN transit.stop_times st
  ON st.feed_version_id = s.feed_version_id AND st.stop_public_id = s.public_id
LEFT JOIN transit.trips trip
  ON trip.feed_version_id = st.feed_version_id AND trip.public_id = st.trip_public_id
GROUP BY s.feed_version_id, s.public_id, s.name, s.mode, s.parent_public_id,
         s.wheelchair_boarding, s.longitude, s.latitude
ORDER BY s.public_id;

-- name: GetSourceHealth :one
SELECT source_id, last_attempt_at, last_success_at, last_failure_at, last_source_updated_at,
       last_valid_snapshot_at, consecutive_failures, last_error_code, entity_count, updated_at
FROM ops.source_health
WHERE source_id = sqlc.arg(source_id);

-- name: UpsertSourceHealthSuccess :exec
INSERT INTO ops.source_health (
  source_id, last_attempt_at, last_success_at, last_source_updated_at,
  last_valid_snapshot_at, consecutive_failures, last_error_code, last_error_safe_detail,
  entity_count, updated_at
) VALUES (
  sqlc.arg(source_id), sqlc.arg(attempted_at), sqlc.arg(succeeded_at),
  sqlc.narg(source_updated_at), sqlc.narg(snapshot_at), 0, NULL, NULL,
  sqlc.arg(entity_count), now()
)
ON CONFLICT (source_id) DO UPDATE SET
  last_attempt_at = EXCLUDED.last_attempt_at,
  last_success_at = EXCLUDED.last_success_at,
  last_source_updated_at = EXCLUDED.last_source_updated_at,
  last_valid_snapshot_at = EXCLUDED.last_valid_snapshot_at,
  consecutive_failures = 0,
  last_error_code = NULL,
  last_error_safe_detail = NULL,
  entity_count = EXCLUDED.entity_count,
  updated_at = now();
