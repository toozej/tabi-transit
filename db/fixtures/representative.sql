-- Sanitized deterministic data only. This is not provider data.
INSERT INTO catalog.sources (id, display_name, provider_name, source_kind, enabled, adapter_version)
VALUES
  ('fixture-gtfs', 'Fixture GTFS', 'Fixture Transit', 'gtfs', false, 'fixture'),
  ('fixture-rt', 'Fixture realtime', 'Fixture Transit', 'gtfs_rt', false, 'fixture');

INSERT INTO catalog.feed_versions (source_id, version_label, archive_sha256, fetched_at, activated_at, status)
VALUES ('fixture-gtfs', 'fixture-20260722', repeat('a', 64), now(), now(), 'active');

INSERT INTO transit.stops (feed_version_id, public_id, source_stop_id, name, mode, point)
SELECT id, 'fixture:stop:bus-nearest', 'bus-nearest', 'Fixture Bus Nearest', 'bus'::transit.mode, ST_SetSRID(ST_MakePoint(-122.6700, 45.5200), 4326)::geography FROM catalog.feed_versions
UNION ALL
SELECT id, 'fixture:stop:bus-second', 'bus-second', 'Fixture Bus Second', 'bus'::transit.mode, ST_SetSRID(ST_MakePoint(-122.6680, 45.5200), 4326)::geography FROM catalog.feed_versions
UNION ALL
SELECT id, 'fixture:stop:rail-nearest', 'rail-nearest', 'Fixture Rail Nearest', 'light_rail'::transit.mode, ST_SetSRID(ST_MakePoint(-122.6700, 45.5210), 4326)::geography FROM catalog.feed_versions
UNION ALL
SELECT id, 'fixture:stop:rail-second', 'rail-second', 'Fixture Rail Second', 'light_rail'::transit.mode, ST_SetSRID(ST_MakePoint(-122.6660, 45.5200), 4326)::geography FROM catalog.feed_versions;
