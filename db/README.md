# Database foundation

`migrations/` is append-only. Production changes use the expand → migrate →
contract pattern so that the prior application image remains compatible during
an image rollback (ADR-0004). Do not edit an applied migration: add a forward
fix with the next sequence number instead.

## Local commands

`TABI_DATABASE_URL` must point at a disposable PostgreSQL database with
PostGIS installed (for example `postgres://tabi:password@127.0.0.1:5432/tabi?sslmode=disable`).

```sh
db/migrate.sh
db/test-migrations.sh
```

The test script starts an isolated PostGIS container when Docker is available,
applies every migration twice, loads only sanitized fixtures, and checks the
nearby stop query and its `EXPLAIN (ANALYZE, BUFFERS)` output. Set
`POSTGIS_TEST_IMAGE` to a reviewed test image. It never uses deployment
secrets or a production database.

## Lock and rollback policy

DDL can take `ACCESS EXCLUSIVE` locks, especially when adding constraints or
indexes. Keep migrations short, set lock and statement timeouts in production
job configuration, create large indexes concurrently in a separately planned
migration, and measure against production-like data before release. This
initial foundation contains only additive tables and indexes. A failed
migration is fixed forward; database downgrades are not automated. Restore is
an isolated, rehearsed operation under the deployment runbook, not an
application rollback mechanism.

`sqlc.yaml` owns generated Go output in `internal/persistence/sqlcgen/`.
Generated files are committed after a deterministic `sqlc generate`; never
hand edit them. The hand-authored repository interfaces in
`internal/persistence/` are intentionally stable boundaries for importers,
pollers, and HTTP handlers.
