-- name: InsertNotificationDelivery :one
-- The unique dedupe_key is the concurrency boundary: concurrent workers may
-- attempt the same event, but only one delivery is materialized.
INSERT INTO app.notification_deliveries (
  id, subscription_id, push_token_id, notification_type, dedupe_key, payload,
  occurred_at, expires_at, status, next_attempt_at
) VALUES (
  sqlc.arg(id), sqlc.arg(subscription_id), sqlc.arg(push_token_id),
  sqlc.arg(notification_type), sqlc.arg(dedupe_key), sqlc.arg(payload),
  sqlc.arg(occurred_at), sqlc.arg(expires_at), 'pending', sqlc.arg(next_attempt_at)
)
ON CONFLICT (dedupe_key) DO NOTHING
RETURNING id;

-- name: ClaimNotificationDeliveries :many
-- Claiming is a single transaction statement. SKIP LOCKED permits independent
-- workers without duplicate sends; claim_until recovers a crashed worker while
-- the expiry predicate prevents a late time-sensitive retry.
WITH candidates AS (
  SELECT d.id
  FROM app.notification_deliveries d
  JOIN app.push_tokens token ON token.id = d.push_token_id
  JOIN app.subscriptions subscription ON subscription.id = d.subscription_id
  WHERE d.expires_at > sqlc.arg(now_at)
    AND subscription.expires_at > sqlc.arg(now_at)
    AND subscription.deleted_at IS NULL
    AND token.disabled_at IS NULL
    AND d.attempts < sqlc.arg(max_attempts)::smallint
    AND (
      (d.status IN ('pending', 'retry_pending') AND (d.next_attempt_at IS NULL OR d.next_attempt_at <= sqlc.arg(now_at)))
      OR (d.status = 'sending' AND d.claim_until <= sqlc.arg(now_at))
    )
  ORDER BY d.expires_at, d.created_at, d.id
  FOR UPDATE OF d SKIP LOCKED
  LIMIT sqlc.arg(row_limit)::integer
), claimed AS (
  UPDATE app.notification_deliveries d
  SET status = 'sending', attempts = d.attempts + 1,
      claim_until = sqlc.arg(claim_until), updated_at = sqlc.arg(now_at)
  FROM candidates c
  WHERE d.id = c.id
  RETURNING d.id, d.subscription_id, d.push_token_id, d.notification_type,
            d.payload, d.expires_at, d.attempts
)
SELECT claimed.id, claimed.subscription_id, claimed.push_token_id,
       claimed.notification_type, claimed.payload, claimed.expires_at,
       claimed.attempts, token.token_ciphertext, token.encryption_key_id,
       subscription.quiet_start, subscription.quiet_end, subscription.quiet_time_zone
FROM claimed
JOIN app.push_tokens token ON token.id = claimed.push_token_id
JOIN app.subscriptions subscription ON subscription.id = claimed.subscription_id
ORDER BY claimed.expires_at, claimed.id;

-- name: DisablePushToken :exec
UPDATE app.push_tokens
SET disabled_at = sqlc.arg(disabled_at), disabled_reason = sqlc.arg(disabled_reason), updated_at = sqlc.arg(disabled_at)
WHERE id = sqlc.arg(id) AND disabled_at IS NULL;

-- name: ExpirePendingDeliveries :execrows
-- Expired notifications are terminal. They must never be retried after a
-- worker outage or a quiet-hours deferral.
UPDATE app.notification_deliveries
SET status = 'expired', next_attempt_at = NULL, updated_at = sqlc.arg(now_at)
WHERE status IN ('pending', 'retry_pending', 'sending')
  AND expires_at <= sqlc.arg(now_at);
