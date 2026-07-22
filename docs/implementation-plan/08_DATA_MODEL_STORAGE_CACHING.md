---
doc_id: TAB-PLAN-008
title: "Data Model, Storage and Caching"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["database-agent", "backend-agent", "mobile-data-agent"]
depends_on: ["TAB-PLAN-002", "TAB-PLAN-007"]
---


# Data Model, Storage and Caching

## PostgreSQL schemas

- `catalog`: sources, agencies, feed versions/imports.
- `transit`: normalized static data.
- `realtime`: current vehicle/trip/alert state.
- `history`: optional bounded observations/revisions.
- `app`: installations, push tokens, subscriptions, flags/config.
- `ops`: source status, locks, audit/job events.

## Entities

Catalog: `sources`, `agencies`, `feed_versions`, `import_runs`, `source_status`.

Static: `stops`, `routes`, `trips`, `stop_times`, calendars/exceptions, `shapes`, transfers/pathways, directions, source-key mappings.

Realtime: `vehicle_current`, `trip_update_current`, `stop_time_update_current`, `alert_current`, active periods, informed entities, snapshots.

App: `installations`, `push_tokens`, `subscriptions`, `notification_deliveries`.

## Spatial design

- Stops/vehicles: `geography(Point,4326)` for meter distance.
- Shapes: `geometry(LineString|MultiLineString,4326)`.
- Service bounds: polygon geometry.
- GiST indexes for points/shapes; B-tree for IDs/routes/trips/service dates/timestamps; partial current/active indexes.

## Public IDs

Source-qualified opaque IDs such as `trimet:stop:12345`, `trimet:route:20`, `trimet:vehicle:2901`. Never expose internal DB IDs. Stable through feed versions where upstream IDs are stable; aliases support known changes.

## Freshness columns

`source_updated_at`, `entity_updated_at`, `fetched_at`, `processed_at`, optional `expires_at`, `freshness_status`, `snapshot_id`, `source_id`.

## History

MVP keeps current state plus a short debugging window. Longer vehicle/history/heatmap storage requires ADR covering purpose, rights, cost, retention, partitions, aggregates, and privacy. If enabled, partition by date, deduplicate unchanged points, aggregate old data, and automate deletion.

## Cache layers

- HTTP ETag/Cache-Control and TanStack Query.
- Bounded server in-process references/config.
- PostgreSQL current snapshots as shared normalized truth.
- Redis only for measured distributed cache/dedupe/locks; never authoritative.
- Terms-aware search/plan caching.

## Mobile SQLite

Suggested:

```text
app_metadata
static_feed_manifest
stops
routes
route_shapes
stop_routes
schedule_departures
favorites
recent_searches
saved_places
saved_itineraries
cached_alerts
cached_vehicle_details
notification_subscriptions
preferences
```

Public data is replaceable; favorites/preferences durable. Do not retain exact origin coordinates in recents unless explicitly saved. Realtime cache expires and carries source age. Clear-cache must preserve user choices.

## Static sync

Compare:

A. versioned compressed SQLite/static artifact with checksum and atomic swap;
B. paginated JSON import.

Phase 1 profiles size, bandwidth, migration, licensing, and file-swap safety. JSON is acceptable for first vertical slice; public MVP should use the simpler option that meets budgets.

## Retention baseline

Configuration/legal approval sets exact values. Initial direction: source metadata 90 days; active/prior GTFS archives plus policy history; raw GTFS-RT disabled or short; current snapshots until replaced; short vehicle debugging; alert revisions sufficient for dedupe; short redacted access logs; notification deliveries 30–90 days; prompt purge on installation deletion.

## Backup/restore

- Run `pg_dump --format=custom` at least every six hours to a root-owned host directory.
- Back up database dumps, feed archives, deployment configuration, Caddy state, and non-regenerable application data with restic to an encrypted off-site repository.
- Provider/VPS snapshots are supplemental and never the only backup.
- Retain multiple daily, weekly, and monthly restore points according to available storage and privacy limits.
- Perform a monthly automated integrity check and a quarterly isolated restore rehearsal.
- Rebuildable public realtime snapshots do not need long retention; installation/subscription data does.
- Verify PostGIS extensions, migrations, active feed version, source registry, and subscription integrity after restore.
- Device SQLite is not a Tabi cloud backup; document platform backup behavior.

## Acceptance

Schema serves MVP without provider leakage; indexed nearby meets budget; feed activation atomic; API IDs stable across sample update; mobile cache rebuild preserves favorites; retention tested; restore rehearsal succeeds; migration lock risk measured.
