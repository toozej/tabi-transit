BEGIN;

-- The static feed's service clock is authoritative for GTFS times, including
-- values beyond 24:00. It is nullable for legacy/imports without agency.txt;
-- readers must fail closed rather than substitute a device or UTC timezone.
ALTER TABLE catalog.feed_versions
  ADD COLUMN service_timezone text;

CREATE INDEX feed_versions_active_timezone_idx
  ON catalog.feed_versions (status, service_timezone)
  WHERE status = 'active' AND service_timezone IS NOT NULL;

CREATE INDEX trip_update_stop_times_lookup_idx
  ON realtime.trip_update_stop_times_current (source_id, stop_public_id, stop_sequence);

COMMIT;
