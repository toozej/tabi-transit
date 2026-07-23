BEGIN;

CREATE TABLE transit.stops (
  feed_version_id bigint NOT NULL REFERENCES catalog.feed_versions(id) ON DELETE CASCADE,
  public_id text NOT NULL,
  source_stop_id text NOT NULL,
  name text NOT NULL,
  code text,
  parent_public_id text,
  platform_code text,
  mode transit.mode NOT NULL DEFAULT 'unknown',
  wheelchair_boarding smallint NOT NULL DEFAULT 0 CHECK (wheelchair_boarding BETWEEN 0 AND 2),
  point geography(Point, 4326) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (feed_version_id, public_id),
  UNIQUE (feed_version_id, source_stop_id),
  CHECK (public_id ~ '^[a-z0-9][a-z0-9-]*:stop:.+$')
);
CREATE INDEX stops_point_gist_idx ON transit.stops USING GIST (point);
CREATE INDEX stops_feed_mode_idx ON transit.stops (feed_version_id, mode, public_id);
CREATE INDEX stops_feed_name_idx ON transit.stops (feed_version_id, lower(name), public_id);

CREATE TABLE transit.routes (
  feed_version_id bigint NOT NULL REFERENCES catalog.feed_versions(id) ON DELETE CASCADE,
  public_id text NOT NULL,
  source_route_id text NOT NULL,
  short_name text,
  long_name text,
  description text,
  mode transit.mode NOT NULL DEFAULT 'unknown',
  color char(6),
  text_color char(6),
  sort_order integer,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (feed_version_id, public_id),
  UNIQUE (feed_version_id, source_route_id),
  CHECK (public_id ~ '^[a-z0-9][a-z0-9-]*:route:.+$'),
  CHECK (color IS NULL OR color ~ '^[0-9A-Fa-f]{6}$'),
  CHECK (text_color IS NULL OR text_color ~ '^[0-9A-Fa-f]{6}$')
);
CREATE INDEX routes_feed_mode_sort_idx ON transit.routes (feed_version_id, mode, sort_order NULLS LAST, public_id);

CREATE TABLE transit.shapes (
  feed_version_id bigint NOT NULL REFERENCES catalog.feed_versions(id) ON DELETE CASCADE,
  public_id text NOT NULL,
  source_shape_id text NOT NULL,
  line geometry(LineString, 4326) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (feed_version_id, public_id),
  UNIQUE (feed_version_id, source_shape_id),
  CHECK (public_id ~ '^[a-z0-9][a-z0-9-]*:shape:.+$'),
  CHECK (ST_NPoints(line) >= 2),
  CHECK (ST_IsValid(line))
);
CREATE INDEX shapes_line_gist_idx ON transit.shapes USING GIST (line);

CREATE TABLE transit.trips (
  feed_version_id bigint NOT NULL REFERENCES catalog.feed_versions(id) ON DELETE CASCADE,
  public_id text NOT NULL,
  source_trip_id text NOT NULL,
  route_public_id text NOT NULL,
  service_id text NOT NULL,
  headsign text,
  direction_id smallint CHECK (direction_id IN (0, 1)),
  block_id text,
  shape_public_id text,
  wheelchair_accessible smallint NOT NULL DEFAULT 0 CHECK (wheelchair_accessible BETWEEN 0 AND 2),
  bikes_allowed smallint NOT NULL DEFAULT 0 CHECK (bikes_allowed BETWEEN 0 AND 2),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (feed_version_id, public_id),
  UNIQUE (feed_version_id, source_trip_id),
  FOREIGN KEY (feed_version_id, route_public_id) REFERENCES transit.routes(feed_version_id, public_id),
  FOREIGN KEY (feed_version_id, shape_public_id) REFERENCES transit.shapes(feed_version_id, public_id),
  CHECK (public_id ~ '^[a-z0-9][a-z0-9-]*:trip:.+$')
);
CREATE INDEX trips_feed_route_idx ON transit.trips (feed_version_id, route_public_id, direction_id, public_id);

CREATE TABLE transit.stop_times (
  feed_version_id bigint NOT NULL,
  trip_public_id text NOT NULL,
  stop_public_id text NOT NULL,
  stop_sequence integer NOT NULL CHECK (stop_sequence >= 0),
  arrival_seconds integer CHECK (arrival_seconds IS NULL OR arrival_seconds >= 0),
  departure_seconds integer CHECK (departure_seconds IS NULL OR departure_seconds >= 0),
  pickup_type smallint NOT NULL DEFAULT 0 CHECK (pickup_type BETWEEN 0 AND 3),
  drop_off_type smallint NOT NULL DEFAULT 0 CHECK (drop_off_type BETWEEN 0 AND 3),
  timepoint boolean,
  PRIMARY KEY (feed_version_id, trip_public_id, stop_sequence),
  FOREIGN KEY (feed_version_id, trip_public_id) REFERENCES transit.trips(feed_version_id, public_id) ON DELETE CASCADE,
  FOREIGN KEY (feed_version_id, stop_public_id) REFERENCES transit.stops(feed_version_id, public_id),
  CHECK (arrival_seconds IS NULL OR departure_seconds IS NULL OR departure_seconds >= arrival_seconds)
);
CREATE INDEX stop_times_feed_stop_idx ON transit.stop_times (feed_version_id, stop_public_id, departure_seconds, trip_public_id);

COMMIT;
