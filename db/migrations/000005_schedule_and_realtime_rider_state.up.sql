BEGIN;

-- GTFS clock values remain seconds since service-day midnight; values over
-- 86400 are intentional and must never be coerced to a calendar-day time.
CREATE TABLE transit.services (
  feed_version_id bigint NOT NULL REFERENCES catalog.feed_versions(id) ON DELETE CASCADE,
  service_id text NOT NULL,
  PRIMARY KEY (feed_version_id, service_id)
);
CREATE TABLE transit.service_calendars (
  feed_version_id bigint NOT NULL REFERENCES catalog.feed_versions(id) ON DELETE CASCADE,
  service_id text NOT NULL,
  monday boolean NOT NULL, tuesday boolean NOT NULL, wednesday boolean NOT NULL,
  thursday boolean NOT NULL, friday boolean NOT NULL, saturday boolean NOT NULL,
  sunday boolean NOT NULL,
  start_date date NOT NULL,
  end_date date NOT NULL,
  PRIMARY KEY (feed_version_id, service_id),
  FOREIGN KEY (feed_version_id, service_id) REFERENCES transit.services(feed_version_id, service_id) ON DELETE CASCADE,
  CHECK (start_date <= end_date)
);
CREATE TABLE transit.service_calendar_dates (
  feed_version_id bigint NOT NULL,
  service_id text NOT NULL,
  service_date date NOT NULL,
  exception_type smallint NOT NULL CHECK (exception_type IN (1, 2)),
  PRIMARY KEY (feed_version_id, service_id, service_date),
  FOREIGN KEY (feed_version_id, service_id) REFERENCES transit.services(feed_version_id, service_id) ON DELETE CASCADE
);
CREATE INDEX service_calendar_dates_lookup_idx ON transit.service_calendar_dates (feed_version_id, service_date, service_id);
CREATE INDEX stop_times_arrivals_idx ON transit.stop_times (feed_version_id, stop_public_id, arrival_seconds, trip_public_id);

-- These are bounded current-state projections, not historical event storage.
CREATE TABLE realtime.trip_updates_current (
  source_id text NOT NULL REFERENCES catalog.sources(id),
  entity_id text NOT NULL,
  snapshot_id bigint NOT NULL REFERENCES realtime.snapshots(id),
  feed_version_id bigint REFERENCES catalog.feed_versions(id),
  trip_public_id text,
  route_public_id text,
  start_date date,
  schedule_relationship text,
  source_updated_at timestamptz,
  fetched_at timestamptz NOT NULL,
  processed_at timestamptz NOT NULL,
  PRIMARY KEY (source_id, entity_id)
);
CREATE INDEX trip_updates_current_trip_idx ON realtime.trip_updates_current (trip_public_id, source_updated_at DESC);
CREATE TABLE realtime.trip_update_stop_times_current (
  source_id text NOT NULL,
  entity_id text NOT NULL,
  stop_sequence integer NOT NULL,
  stop_public_id text,
  arrival_delay_seconds integer,
  arrival_time timestamptz,
  departure_delay_seconds integer,
  departure_time timestamptz,
  schedule_relationship text,
  PRIMARY KEY (source_id, entity_id, stop_sequence),
  FOREIGN KEY (source_id, entity_id) REFERENCES realtime.trip_updates_current(source_id, entity_id) ON DELETE CASCADE
);
CREATE TABLE realtime.alerts_current (
  source_id text NOT NULL REFERENCES catalog.sources(id),
  entity_id text NOT NULL,
  snapshot_id bigint NOT NULL REFERENCES realtime.snapshots(id),
  cause text,
  effect text,
  header_text text,
  description_text text,
  url text,
  active_from timestamptz,
  active_until timestamptz,
  fetched_at timestamptz NOT NULL,
  processed_at timestamptz NOT NULL,
  PRIMARY KEY (source_id, entity_id),
  CHECK (active_until IS NULL OR active_from IS NULL OR active_until >= active_from)
);

COMMIT;
