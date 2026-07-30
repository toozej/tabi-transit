---
doc_id: TAB-PLAN-015
title: "Phased Delivery Roadmap"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["program-manager", "technical-lead", "all-agents"]
depends_on: ["TAB-PLAN-001", "TAB-PLAN-002", "TAB-PLAN-016"]
---


# Delivery Roadmap

Phases are dependency gates rather than calendar promises. Parallel work begins after prerequisites pass.

## Phase 0 — Decisions and spikes

Work:

- Register/review TriMet AppID and terms.
- Create Mapbox environment tokens, budgets, and style prototype.
- Verify TriMet Streetcar and Rose City-inspired presentation coverage through the official TriMet interfaces.
- Pin Expo/RN/RNMapbox/Mapbox native/Go/Postgres versions.
- Physical RNMapbox and 1,500-point performance spike.
- Expo CNG/EAS builds.
- Vitest + RNTL harness.
- PostGIS nearest/group-per-mode query.
- GTFS and GTFS-RT parsing/fixtures.
- TriMet planner/filter capability.
- Linux host provider/sizing, domain/DNS, backup repository, and observability profile decisions.
- ADRs, source registry, OpenAPI skeleton, repository bootstrap.

Exit: reproducible stack, blockers categorized, accounts/controls present, no MVP-invalidating uncertainty.

## Phase 1 — Foundation and live-map vertical slice

Parallel:

- Repo/CI/local Docker.
- DB/source registry/importer/vehicle poller/config/routes/stops/vehicles API.
- Expo shell/SQLite/API client/Mapbox/live layer/search/detail/credits skeleton.
- Dev/preview Docker Compose, GHCR, Linux host, Caddy, PostGIS volumes, backups, and dashboards/alerts.

Demo: physical iOS/Android app opens normalized system vehicle map, exact vehicle search, source/freshness.

Exit: end-to-end slice, atomic GTFS activation, invalid RT preservation, automated dev deployment, accessible list alternative.

## Phase 2 — Rider information MVP

Nearby with `limitPerMode`, stop/arrivals, route/shapes, static schedules/offline manifest, alerts, favorites/recents, degraded states, performance/accessibility.

Exit: critical Maestro flows, offline static data, consistent freshness, budgets and source-outage tests.

## Phase 3 — Trip planning

Mapbox search/geocoding adapter and picker; TriMet planner adapter; itinerary schema; mode/transfer/walking/accessibility policy; timeline/map; deep links; quality comparison.

Exit: A-to-B from current/address/POI/map/stop; constraints enforced/disclosed; alert/freshness; terms review; location-denied planning.

## Phase 4 — Notifications and enriched vehicle status

Installation/push, alert/departure subscriptions, worker/receipts, quiet/expiry/delete, detailed vehicle/trip/block, background cache maintenance.

Exit: no duplicate/late time-sensitive push; rotation/deletion/privacy; outage isolation.

## Phase 5 — Streetcar and specialist views

Implement TriMet-provided Streetcar normalization and Rose City-inspired
presentation through the existing official source boundaries. ADR-0013 clears
the 30-day normalized-history persistence gate; rider-visible
status/block/adherence/history screens still need their API/UI contract and
TriMet D-001 production enablement.

Exit: TriMet terms/enablement, health/runbooks, no duplicates, accessible
alternatives, and the accepted 30-day retention ADR.

## Phase 6 — Production launch

Staging/production Compose hosts/projects, off-site backup and host-loss restore, capacity, security/privacy/accessibility review, store assets, beta, phased rollout, incident rehearsal/support.

Exit: product gate in `01_PRODUCT_REQUIREMENTS.md`.

## Optional parallel path — Fly.io trial or low-cost deployment

After the Docker images and contracts are stable, validate `fly.toml` process groups using the same backend image. Do not make Fly.io the primary dependency and do not assume a permanent free tier. Use a separate database plan and off-site backups.

## Phase 7 — Evidence-driven post-launch

Redis, OTP, static deltas, regional expansion, history/analytics, tablet, account sync, alternative providers only when measured need exists.

## Critical path

```text
source/legal verification
-> Phase 0 native/data proof
-> OpenAPI/DB
-> GTFS/GTFS-RT normalization
-> live map vertical slice
-> stops/routes/schedules/alerts
-> search/planner
-> production hardening
```

Notifications and optional sources run off-path after foundations.

## Management rules

Every phase has a releaseable demo; do not carry unbounded debt; keep incomplete providers flagged off; review risk weekly; re-estimate after spikes; failed approval gates/removes a feature and never justifies scraping.
