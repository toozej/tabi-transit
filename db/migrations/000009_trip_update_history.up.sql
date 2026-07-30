BEGIN;

-- Normalized historical schedule-deviation evidence only; raw feeds are never retained.
CREATE TABLE history.trip_update_observations (
  source_id text NOT NULL REFERENCES catalog.sources(id),
  snapshot_id bigint NOT NULL REFERENCES realtime.snapshots(id),
  entity_id text NOT NULL,
  trip_public_id text NOT NULL,
  route_public_id text,
  start_date date,
  stop_sequence integer NOT NULL DEFAULT 0,
  stop_public_id text,
  trip_schedule_relationship text,
  stop_schedule_relationship text,
  arrival_delay_seconds integer,
  departure_delay_seconds integer,
  arrival_time timestamptz,
  departure_time timestamptz,
  source_updated_at timestamptz,
  fetched_at timestamptz NOT NULL,
  processed_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (source_id, snapshot_id, entity_id, stop_sequence),
  CHECK (processed_at >= fetched_at - interval '5 minutes')
);
CREATE INDEX trip_update_observations_trip_time_idx ON history.trip_update_observations (trip_public_id, processed_at DESC);
CREATE INDEX trip_update_observations_processed_idx ON history.trip_update_observations (processed_at);

COMMIT;
