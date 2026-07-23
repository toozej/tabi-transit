BEGIN;

-- WP-12 notification foundation. This is additive and deliberately separates
-- installation/push data from public transit schemas. No plaintext credential
-- or push token column exists: callers store a verifier or authenticated
-- ciphertext only. Forward fixes are required after this migration is applied.
CREATE SCHEMA IF NOT EXISTS app;

CREATE TYPE app.subscription_type AS ENUM ('service_alert', 'departure_reminder');
CREATE TYPE app.notification_delivery_status AS ENUM ('pending', 'sending', 'sent', 'retry_pending', 'expired', 'failed', 'disabled');

CREATE TABLE app.installations (
  id uuid PRIMARY KEY,
  credential_verifier char(64) NOT NULL UNIQUE CHECK (credential_verifier ~ '^[0-9a-f]{64}$'),
  platform text NOT NULL CHECK (platform IN ('ios', 'android')),
  locale text NOT NULL CHECK (length(locale) <= 35),
  time_zone text NOT NULL CHECK (length(time_zone) <= 100),
  app_version text NOT NULL CHECK (length(app_version) <= 100),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE TABLE app.push_tokens (
  id uuid PRIMARY KEY,
  installation_id uuid NOT NULL REFERENCES app.installations(id) ON DELETE CASCADE,
  token_hash char(64) NOT NULL UNIQUE CHECK (token_hash ~ '^[0-9a-f]{64}$'),
  token_ciphertext bytea NOT NULL CHECK (octet_length(token_ciphertext) > 28),
  encryption_key_id text NOT NULL CHECK (length(encryption_key_id) BETWEEN 1 AND 128),
  platform text NOT NULL CHECK (platform IN ('ios', 'android')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  last_used_at timestamptz,
  disabled_at timestamptz,
  disabled_reason text CHECK (disabled_reason IN ('invalid_token', 'installation_deleted', 'rotated', 'operator_disabled')),
  CHECK ((disabled_at IS NULL) = (disabled_reason IS NULL))
);
CREATE UNIQUE INDEX push_tokens_one_active_per_installation_platform
  ON app.push_tokens (installation_id, platform) WHERE disabled_at IS NULL;
CREATE INDEX push_tokens_active_lookup_idx ON app.push_tokens (installation_id, id) WHERE disabled_at IS NULL;

CREATE TABLE app.subscriptions (
  id uuid PRIMARY KEY,
  installation_id uuid NOT NULL REFERENCES app.installations(id) ON DELETE CASCADE,
  subscription_type app.subscription_type NOT NULL,
  route_ids text[] NOT NULL DEFAULT '{}',
  stop_ids text[] NOT NULL DEFAULT '{}',
  mode_ids text[] NOT NULL DEFAULT '{}',
  source_ids text[] NOT NULL DEFAULT '{}',
  trip_id text,
  remind_at timestamptz,
  expires_at timestamptz NOT NULL,
  quiet_start time,
  quiet_end time,
  quiet_time_zone text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CHECK (expires_at > created_at),
  CHECK ((quiet_start IS NULL AND quiet_end IS NULL AND quiet_time_zone IS NULL)
     OR (quiet_start IS NOT NULL AND quiet_end IS NOT NULL AND quiet_time_zone IS NOT NULL)),
  CHECK (
    (subscription_type = 'service_alert'
      AND cardinality(route_ids) + cardinality(stop_ids) + cardinality(mode_ids) + cardinality(source_ids) > 0
      AND trip_id IS NULL AND remind_at IS NULL)
    OR
    (subscription_type = 'departure_reminder'
      AND trip_id IS NOT NULL AND remind_at IS NOT NULL AND remind_at < expires_at)
  )
);
CREATE INDEX subscriptions_active_expiry_idx ON app.subscriptions (expires_at, id) WHERE deleted_at IS NULL;
CREATE INDEX subscriptions_route_ids_gin_idx ON app.subscriptions USING GIN (route_ids) WHERE deleted_at IS NULL;
CREATE INDEX subscriptions_stop_ids_gin_idx ON app.subscriptions USING GIN (stop_ids) WHERE deleted_at IS NULL;
CREATE INDEX subscriptions_source_ids_gin_idx ON app.subscriptions USING GIN (source_ids) WHERE deleted_at IS NULL;

CREATE TABLE app.notification_deliveries (
  id uuid PRIMARY KEY,
  subscription_id uuid NOT NULL REFERENCES app.subscriptions(id) ON DELETE CASCADE,
  push_token_id uuid NOT NULL REFERENCES app.push_tokens(id) ON DELETE CASCADE,
  notification_type text NOT NULL CHECK (notification_type IN ('service_alert', 'departure_reminder')),
  dedupe_key char(64) NOT NULL UNIQUE CHECK (dedupe_key ~ '^[0-9a-f]{64}$'),
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  status app.notification_delivery_status NOT NULL DEFAULT 'pending',
  attempts smallint NOT NULL DEFAULT 0 CHECK (attempts >= 0 AND attempts <= 3),
  next_attempt_at timestamptz,
  claim_until timestamptz,
  provider_ticket_id text UNIQUE,
  last_error_code text,
  sent_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (expires_at > occurred_at),
  CHECK (payload ? 'deepLink' AND payload ? 'entityId' AND payload ? 'subscriptionId'),
  CHECK (NOT (payload ? 'token') AND NOT (payload ? 'credential') AND NOT (payload ? 'coordinate'))
);
CREATE INDEX notification_deliveries_claim_idx
  ON app.notification_deliveries (status, next_attempt_at, claim_until, expires_at, created_at)
  WHERE status IN ('pending', 'retry_pending', 'sending');
CREATE INDEX notification_deliveries_subscription_idx ON app.notification_deliveries (subscription_id, created_at DESC);

CREATE TABLE app.notification_receipts (
  provider_ticket_id text PRIMARY KEY REFERENCES app.notification_deliveries(provider_ticket_id) ON DELETE CASCADE,
  received_at timestamptz NOT NULL,
  status text NOT NULL CHECK (status IN ('ok', 'error')),
  error_code text,
  safe_detail text,
  processed_at timestamptz NOT NULL DEFAULT now()
);

COMMIT;
