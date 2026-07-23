# Go services

Service entrypoints are added by their owning work packages. The directories are
reserved now so the root Go module and stable `make dev-*` commands have a
consistent layout:

- `transit-api` — WP-07 public API.
- `gtfs-importer` — WP-04 static GTFS import service.
- `realtime-poller` — WP-05 GTFS-Realtime poller.
- `notification-worker` — WP-12 notification delivery worker.

Do not add provider credentials to service directories; use the documented
environment and `_FILE` configuration conventions.
