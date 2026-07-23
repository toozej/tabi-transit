# Transit data and PostGIS spike

This isolated, synthetic Phase 0 proof informs WP-03 through WP-05. It is not
a production importer, migration, source adapter, or retention implementation.
It makes no network request to a transit provider and contains no credentials.

## What is proved

- `fixtures/gtfs/` is a small valid static GTFS-shaped CSV set. The Go test
  verifies required-file presence, IDs/references, coordinates, and a trip
  containing `25:15:30` / `25:30:00` times.
- GTFS times are represented as seconds since the source service-day midnight:
  `25:15:30` is `90,930`, not a device-local-clock time. Production code must
  combine this value with service date and the source time zone before deriving
  an instant.
- `fixtures/gtfsrt/vehicle_positions.pb.base64` is a deterministic synthetic
  `FeedMessage` protobuf encoded with the pinned official MobilityData binding.
  The test decodes and parses those bytes, then also rejects malformed bytes and
  impossible coordinates.
- `sql/nearby_limit_per_mode.sql` runs real PostGIS geography distance and uses
  `row_number() over (partition by mode ...)` to prove deterministic
  `limitPerMode=1`: `bus-nearest` and `rail-nearest` are returned.

## Reproduce

```sh
cd spikes/transit-data
GOWORK=off GOCACHE=/tmp/tabi-transit-spike-go-build go test ./...
./run-postgis-proof.sh
```

The PostGIS image (`postgis/postgis:17-3.5`) is an experimental tag for this
spike, not a production image selection or digest pin. The locally tested pull
resolved to `sha256:404171ea9058c801f405af25d63b3b8e5c9e50f2759e49390dbcc3c7ee533f4d`;
WP-03/WP-13 must independently choose and pin a production digest through the
version-matrix ADR gate.

## Observations and limits

- Fixtures are deliberately tiny and sanitized, so they establish correctness
  shape, not importer or query performance budgets. WP-03 must measure an
  indexed representative dataset with `EXPLAIN (ANALYZE, BUFFERS)`.
- `ST_DWithin` on `geography(Point,4326)` gives meter-radius semantics. A
  production table needs a GiST index and a stable secondary ordering key.
- A malformed, stale, empty, or regressing realtime feed must update source
  health only; it must not replace a prior valid snapshot (ADR-0003). This
  spike validates bytes/coordinates but does not implement snapshots.
- GTFS service-day times over 24:00 are valid and must remain numeric seconds
  through static storage; never reinterpret them in the device time zone.
