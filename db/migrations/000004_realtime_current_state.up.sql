BEGIN;

CREATE TABLE realtime.snapshots (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  source_id text NOT NULL REFERENCES catalog.sources(id),
  feed_version_id bigint REFERENCES catalog.feed_versions(id),
  source_updated_at timestamptz,
  fetched_at timestamptz NOT NULL,
  processed_at timestamptz NOT NULL,
  entity_count integer NOT NULL CHECK (entity_count >= 0),
  content_sha256 char(64) NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
  is_valid boolean NOT NULL,
  validation_report jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (source_id, content_sha256),
  CHECK (processed_at >= fetched_at - interval '5 minutes')
);
CREATE INDEX snapshots_source_processed_idx ON realtime.snapshots (source_id, processed_at DESC);

CREATE TABLE realtime.vehicle_current (
  source_id text NOT NULL REFERENCES catalog.sources(id),
  public_id text NOT NULL,
  source_vehicle_id text NOT NULL,
  snapshot_id bigint NOT NULL REFERENCES realtime.snapshots(id),
  feed_version_id bigint REFERENCES catalog.feed_versions(id),
  route_public_id text,
  trip_public_id text,
  block_id text,
  mode transit.mode NOT NULL DEFAULT 'unknown',
  direction_id smallint CHECK (direction_id IN (0, 1)),
  headsign text,
  point geography(Point, 4326) NOT NULL,
  bearing numeric(5,2) CHECK (bearing IS NULL OR (bearing >= 0 AND bearing < 360)),
  speed_meters_per_second numeric(8,3) CHECK (speed_meters_per_second IS NULL OR speed_meters_per_second >= 0),
  current_stop_public_id text,
  next_stop_public_id text,
  schedule_deviation_seconds integer,
  in_service boolean NOT NULL DEFAULT true,
  source_updated_at timestamptz,
  entity_updated_at timestamptz,
  fetched_at timestamptz NOT NULL,
  processed_at timestamptz NOT NULL,
  expires_at timestamptz,
  freshness_status realtime.freshness_status NOT NULL DEFAULT 'unknown',
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (source_id, public_id),
  UNIQUE (source_id, source_vehicle_id),
  CHECK (public_id ~ '^[a-z0-9][a-z0-9-]*:vehicle:.+$'),
  CHECK (expires_at IS NULL OR expires_at >= fetched_at)
);
CREATE INDEX vehicle_current_point_gist_idx ON realtime.vehicle_current USING GIST (point);
CREATE INDEX vehicle_current_route_idx ON realtime.vehicle_current (route_public_id, mode, public_id);
CREATE INDEX vehicle_current_snapshot_idx ON realtime.vehicle_current (snapshot_id, freshness_status, public_id);
CREATE INDEX vehicle_current_exact_id_idx ON realtime.vehicle_current (public_id);

COMMIT;
