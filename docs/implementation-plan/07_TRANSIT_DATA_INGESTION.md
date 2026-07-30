---
doc_id: TAB-PLAN-007
title: "Transit Data Ingestion and Normalization"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["data-agent", "backend-agent", "transit-domain-agent"]
depends_on: ["TAB-PLAN-001", "TAB-PLAN-002", "TAB-PLAN-008"]
---


# Transit Data Ingestion

## Source order

1. TriMet official GTFS Schedule.
2. TriMet official GTFS-Realtime.
3. TriMet official Web Services.
4. Streetcar rows included in TriMet official GTFS, GTFS-Realtime, and Arrivals V2.
5. Rose City-inspired presentation from normalized TriMet data only.

No human webpage becomes a production machine dependency without legal/technical ADR approval.

## Source registry

Store source ID/name, provider/agency, type, URL, auth, cadence, expected freshness, terms/license/review date, attribution, environments, contact/owner, last success/failure, and adapter version.

## GTFS importer

Pipeline:

1. Fetch safely and record URL, timestamp, HTTP metadata, size, SHA-256.
2. Save the immutable raw ZIP to the host feed-archive volume, keyed by source/date/digest, and include it in encrypted off-site backups when retention permits.
3. Defend against zip bombs, path traversal, duplicate files, and encoding issues.
4. Run MobilityData validator/equivalent.
5. Apply Tabi checks: required files, IDs/foreign keys, coordinates, service-calendar coverage, non-empty essentials, shape order, regression thresholds.
6. Bulk-load versioned staging tables.
7. Transform normalized tables and build indexes/derivatives.
8. Compare active version.
9. Atomically activate and expose version/checksum.
10. Emit report/metrics; retain prior version for rollback.

Support relevant GTFS files: agency, stops, routes, trips, stop_times, calendars/exceptions, shapes, feed_info, transfers, pathways, levels, translations, and later fares. Preserve TriMet extensions in raw/source tables and map only approved fields.

### Service day

Store service date and seconds since service-day midnight because times can exceed 24:00. Derive instants only with source time zone and date; do not use device time zone for schedule semantics.

## GTFS-Realtime

Feeds: Vehicle Positions, Trip Updates, Service Alerts.

Per fetch validate protobuf, feed version/timestamp/regression, entity/deletion counts, coordinate/bearing/speed sanity, static references, empty-feed anomalies, and age. Preserve unmatched entities with diagnostics rather than silently dropping.

Current snapshot update is transactional and invalid input only updates health. Store source/entity/fetch/process timestamps, snapshot ID/hash, and optional approved bounded history.

Freshness states `fresh`, `aging`, `stale`, `unknown` are source-configurable and based on feed/entity timestamp, fetch time, last success, and clock-skew tolerance.

## TriMet Web Services

Use for detailed arrivals, route/stop, vehicle/trip/block status, and planning not represented adequately in GTFS/RT. Server-side AppID, terms-aware caching/retention, fixtures, schema-drift monitoring, feature flags for beta, explicit provider↔GTFS ID mapping.

## Portland Streetcar coverage through TriMet

1. Ingest and classify `route_type=5` as `streetcar` from TriMet GTFS.
2. Consume Streetcar arrivals only through TriMet Arrivals V2, which documents
   Streetcar results as included by default.
3. Preserve TriMet source/freshness semantics and the documented Streetcar
   estimate marker where represented internally.
4. Do not add a Portland Streetcar, PBOT, or UmoIQ adapter, credential, or
   HTML/JavaScript scraper.
5. Production use remains subject to the TriMet D-001 terms and enablement
   decision.

## Rose City Transit-inspired presentation

Use normalized TriMet data only. Do not request, fetch, store, or scrape Rose
City data, and do not imply a partnership. Any long-lived history remains
subject to the separate retention/data-rights decision.

## Quality dashboards

Last success/age, entity counts, unmatched references, invalid coordinates, timestamp regression, duplicate vehicles, cancelled/skipped trips, alert counts, static row deltas, service coverage, parse/import duration.

## Fixtures

Normal/empty/stale/malformed/deleted/unmatched realtime; vehicle without trip; skipped stop; multi-entity alert; after-midnight; duplicate IDs; malicious ZIP; static regression; TriMet Streetcar examples.

## Acceptance

Import/validate/activate/rollback works; invalid feeds preserve active data; every realtime entity has source/freshness; ID mismatch diagnostics visible; optional sources remain gated; jobs idempotent; artifacts traceable by digest; retention/licenses enforced.
