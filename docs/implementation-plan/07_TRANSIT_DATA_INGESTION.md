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
4. Official/canonical Portland Streetcar or other agency feed.
5. Rose City Transit only under documented API/permission/license.

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

## Portland Streetcar gate

1. Contact Streetcar/PBOT and identify canonical feeds/API.
2. Determine what is already in TriMet GTFS/RT.
3. Clarify UmoIQ relationship.
4. Obtain docs, credentials, limits, license, attribution.
5. Capture fixtures.
6. Implement adapter and deduplication.
7. Stage/field validate.
8. Enable by feature flag.

No HTML/JavaScript scraping by default.

## Rose City Transit gate

Request API/export, rate limits, license/redistribution, attribution/social links, advanced fields/history rights, change contact, and preferred contribution model. Treat it as a collaborator/reference, not a hidden dependency.

## Quality dashboards

Last success/age, entity counts, unmatched references, invalid coordinates, timestamp regression, duplicate vehicles, cancelled/skipped trips, alert counts, static row deltas, service coverage, parse/import duration.

## Fixtures

Normal/empty/stale/malformed/deleted/unmatched realtime; vehicle without trip; skipped stop; multi-entity alert; after-midnight; duplicate IDs; malicious ZIP; static regression; approved Streetcar examples.

## Acceptance

Import/validate/activate/rollback works; invalid feeds preserve active data; every realtime entity has source/freshness; ID mismatch diagnostics visible; optional sources remain gated; jobs idempotent; artifacts traceable by digest; retention/licenses enforced.
