BEGIN;

CREATE TABLE catalog.sources (
  id text PRIMARY KEY,
  display_name text NOT NULL,
  provider_name text NOT NULL,
  source_kind text NOT NULL CHECK (source_kind IN ('gtfs', 'gtfs_rt', 'web_service', 'other')),
  agency_id text,
  base_url text,
  cadence_seconds integer CHECK (cadence_seconds IS NULL OR cadence_seconds > 0),
  fresh_after_seconds integer CHECK (fresh_after_seconds IS NULL OR fresh_after_seconds > 0),
  stale_after_seconds integer CHECK (stale_after_seconds IS NULL OR stale_after_seconds > 0),
  attribution text,
  terms_url text,
  enabled boolean NOT NULL DEFAULT false,
  adapter_version text NOT NULL DEFAULT 'unconfigured',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (id ~ '^[a-z0-9][a-z0-9-]{1,126}$'),
  CHECK (stale_after_seconds IS NULL OR fresh_after_seconds IS NULL OR stale_after_seconds >= fresh_after_seconds)
);

CREATE TABLE catalog.feed_versions (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  source_id text NOT NULL REFERENCES catalog.sources(id),
  version_label text NOT NULL,
  archive_sha256 char(64) NOT NULL CHECK (archive_sha256 ~ '^[0-9a-f]{64}$'),
  source_published_at timestamptz,
  fetched_at timestamptz NOT NULL,
  activated_at timestamptz,
  status catalog.feed_version_status NOT NULL DEFAULT 'staged',
  import_report jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (source_id, archive_sha256),
  CHECK ((status = 'active') = (activated_at IS NOT NULL))
);
CREATE UNIQUE INDEX feed_versions_one_active_per_source
  ON catalog.feed_versions (source_id) WHERE status = 'active';
CREATE INDEX feed_versions_source_created_idx ON catalog.feed_versions (source_id, created_at DESC);

CREATE TABLE catalog.import_runs (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  source_id text NOT NULL REFERENCES catalog.sources(id),
  feed_version_id bigint REFERENCES catalog.feed_versions(id),
  started_at timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz,
  outcome text NOT NULL CHECK (outcome IN ('running', 'succeeded', 'failed', 'rolled_back')),
  report jsonb NOT NULL DEFAULT '{}'::jsonb,
  failure_code text,
  CHECK ((outcome = 'running') = (finished_at IS NULL))
);
CREATE INDEX import_runs_source_started_idx ON catalog.import_runs (source_id, started_at DESC);

CREATE TABLE ops.source_health (
  source_id text PRIMARY KEY REFERENCES catalog.sources(id),
  last_attempt_at timestamptz,
  last_success_at timestamptz,
  last_failure_at timestamptz,
  last_source_updated_at timestamptz,
  last_valid_snapshot_at timestamptz,
  consecutive_failures integer NOT NULL DEFAULT 0 CHECK (consecutive_failures >= 0),
  last_error_code text,
  last_error_safe_detail text,
  entity_count integer CHECK (entity_count IS NULL OR entity_count >= 0),
  updated_at timestamptz NOT NULL DEFAULT now()
);

COMMIT;
