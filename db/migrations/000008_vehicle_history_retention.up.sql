BEGIN;

CREATE SCHEMA IF NOT EXISTS history;

-- Normalized observations only: provider payloads are never retained. The
-- realtime writer prunes rows older than the approved 30-day policy within
-- the same transaction that records each valid replacement snapshot.
CREATE TABLE history.vehicle_observations (
  source_id text NOT NULL REFERENCES catalog.sources(id),
  public_id text NOT NULL,
  source_vehicle_id text NOT NULL,
  snapshot_id bigint NOT NULL REFERENCES realtime.snapshots(id),
  route_public_id text,
  trip_public_id text,
  mode transit.mode NOT NULL DEFAULT 'unknown',
  point geography(Point, 4326) NOT NULL,
  source_updated_at timestamptz,
  entity_updated_at timestamptz,
  fetched_at timestamptz NOT NULL,
  processed_at timestamptz NOT NULL,
  freshness_status realtime.freshness_status NOT NULL DEFAULT 'unknown',
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (source_id, snapshot_id, public_id),
  CHECK (public_id ~ '^[a-z0-9][a-z0-9-]*:vehicle:.+$'),
  CHECK (processed_at >= fetched_at - interval '5 minutes')
);

CREATE INDEX vehicle_observations_processed_idx
  ON history.vehicle_observations (processed_at);
CREATE INDEX vehicle_observations_vehicle_time_idx
  ON history.vehicle_observations (source_id, public_id, processed_at DESC);

COMMIT;
