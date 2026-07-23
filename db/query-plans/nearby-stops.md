# Nearby stops query-plan evidence

Run on 2026-07-22 by `db/test-migrations.sh` against an isolated
`postgis/postgis:17-3.5` container and the sanitized four-stop fixture.

```text
Limit (actual time=4.525..4.527 rows=4 loops=1)
  -> Sort
       Sort Key: (s.point <-> input.point), s.public_id
       -> Index Scan using stops_feed_name_idx on transit.stops s
            Index Cond: (feed_version_id = active feed version)
            Filter: st_dwithin(point, input.point, 500)
```

The harness also asserts the windowed per-mode result:

```text
bus:fixture:stop:bus-nearest
light_rail:fixture:stop:rail-nearest
```

The `transit.stops.point` GiST index exists and is the required spatial access
path for production-sized data. This four-row fixture is deliberately too
small to prove that the planner chooses it; the observed plan instead starts
with the active-feed B-tree index and filters spatially. Before a release,
load a representative sanitized dataset, run the same command with
`EXPLAIN (ANALYZE, BUFFERS)`, record latency/buffer budget evidence, and only
then make a performance claim.
