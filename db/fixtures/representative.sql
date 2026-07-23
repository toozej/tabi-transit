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

INSERT INTO transit.routes (feed_version_id, public_id, source_route_id, short_name, long_name, mode, color, text_color)
SELECT id, 'fixture:route:20', '20', '20', 'Fixture Route', 'bus'::transit.mode, '0055A4', 'FFFFFF'
FROM catalog.feed_versions WHERE source_id = 'fixture-gtfs' AND status = 'active';

INSERT INTO transit.trips (feed_version_id, public_id, source_trip_id, route_public_id, service_id)
SELECT id, 'fixture:trip:20-1', '20-1', 'fixture:route:20', 'weekday'
FROM catalog.feed_versions WHERE source_id = 'fixture-gtfs' AND status = 'active';

INSERT INTO transit.stop_times (feed_version_id, trip_public_id, stop_public_id, stop_sequence, arrival_seconds, departure_seconds)
SELECT id, 'fixture:trip:20-1', 'fixture:stop:bus-nearest', 1, 28800, 28800
FROM catalog.feed_versions WHERE source_id = 'fixture-gtfs' AND status = 'active';

INSERT INTO realtime.snapshots (source_id, feed_version_id, fetched_at, processed_at, entity_count, content_sha256, is_valid)
SELECT 'fixture-rt', id, now(), now(), 1, repeat('b', 64), true
FROM catalog.feed_versions WHERE source_id = 'fixture-gtfs' AND status = 'active';

INSERT INTO realtime.vehicle_current (source_id, public_id, source_vehicle_id, snapshot_id, feed_version_id, route_public_id, mode, point, fetched_at, processed_at, freshness_status)
SELECT 'fixture-rt', 'fixture:vehicle:2901', '2901', snapshot.id, feed.id, 'fixture:route:20', 'bus'::transit.mode,
       ST_SetSRID(ST_MakePoint(-122.6700, 45.5200), 4326)::geography, now(), now(), 'fresh'::realtime.freshness_status
FROM catalog.feed_versions feed
JOIN realtime.snapshots snapshot ON snapshot.source_id = 'fixture-rt'
WHERE feed.source_id = 'fixture-gtfs' AND feed.status = 'active';

INSERT INTO ops.source_health (source_id, last_attempt_at, last_success_at, last_valid_snapshot_at, consecutive_failures, entity_count)
VALUES ('fixture-rt', now(), now(), now(), 0, 1);
