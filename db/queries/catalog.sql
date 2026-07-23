-- name: GetActiveFeedVersion :one
SELECT id, source_id, version_label, archive_sha256, activated_at
FROM catalog.feed_versions
WHERE source_id = sqlc.arg(source_id) AND status = 'active';

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
