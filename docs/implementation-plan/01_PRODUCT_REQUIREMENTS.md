---
doc_id: TAB-PLAN-001
title: "Product Requirements and Acceptance Criteria"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["product-agent", "mobile-agent", "backend-agent", "qa-agent"]
depends_on: ["TAB-PLAN-000"]
---


# Product Requirements

## Product statement

Tabi helps riders find, understand, and navigate Portland-area transit using scheduled and realtime data. It supports map-first exploration without making the map the only accessible interface. It is not an official TriMet or Portland Streetcar product unless a future written partnership says otherwise.

## Platforms and roles

- iOS and Android phones; responsive tablet support without a separate first-release design.
- No account required for core use.
- Roles: rider, power user/transit enthusiast, anonymous notification subscriber, and operator/maintainer through internal observability only.

## Functional requirements

### Planning and directions

**FR-001 — A-to-B planning.** Accept current location, address/intersection, POI, map pin, saved/recent place, stop ID, or stop name. Results include departure/arrival, duration, walking, transit legs, route/headsign, boarding/alighting, transfers, realtime versus scheduled status, and relevant alerts.

**FR-002 — Filters.** Support allowed modes, maximum connections/transfers, less walking, depart-at/arrive-by, wheelchair accessibility where sourced, and optimization for fastest/fewer transfers/least walking. Unsupported upstream constraints are deterministically post-filtered/ranked and disclosed.

**FR-003 — External maps.** Open walking or coordinate directions through platform deep links while keeping Tabi’s transit itinerary authoritative.

### Nearby transit

**FR-010 — Nearby stops.** Accessible list and map sorted by defined distance.

**FR-011 — Mode filtering.** Find nearby stops/stations for selected modes.

**FR-012 — Per-mode limit.** General `limitPerMode`, including “nearest two of each selected mode.”

**FR-013 — Permission denial.** Typed location, stop ID, and map-center workflows remain available after location denial without repeated prompts.

### Stops, routes, schedules

**FR-020 — Stop details.** Name/ID, textual location, routes/modes, arrivals, source freshness, alerts, map position, saved state.

**FR-021 — Arrivals.** Distinguish scheduled/estimated, cancellation/skipped status, route/headsign, vehicle/trip when known, and freshness.

**FR-022 — Route details.** Directions, shapes, stops, alerts, current vehicles, schedules.

**FR-023 — Schedules.** Browse by route, direction, stop, and service date; installed static schedules remain usable offline.

### Vehicles and live map

**FR-030 — System map.** Display all normalized current vehicles with mode, route, direction, in-service, schedule/stale, and source filters where data exists.

**FR-031 — Rendering.** Use Mapbox `ShapeSource` and native style layers, not one React view per vehicle.

**FR-032 — Vehicle ID search.** Exact and partial search; exact first; center and open details.

**FR-033 — Vehicle details.** ID, mode, route, trip/block, direction/headsign, current/next stop, timestamp, bearing/location, schedule deviation, source, stale state when present. Omit unsupported fields.

**FR-034 — Staleness.** Clearly show age; never animate stale data as current.

**FR-035 — Streetcar.** Support Streetcar data supplied by TriMet's official GTFS, GTFS-Realtime, and Arrivals V2 interfaces; do not add a direct public-map scraper or imply a Portland Streetcar partnership.

**FR-036 — Specialist views.** Optional post-MVP SystemMapper-inspired service status, block/trip exploration, adherence, and history/heatmap screens, independently gated by rights and data.

### Alerts

**FR-040 — Alert list.** Filter by mode, route, stop, effect/severity, active state.

**FR-041 — Alert detail.** Header, description, cause/effect if provided, affected entities, active period, source URL, last update/freshness.

**FR-042 — Detours/blockages/outages.** Normalize only when source semantics allow; preserve provider wording and do not invent severity.

### Local/offline

**FR-050 — Favorites.** Device-local stops, routes, vehicles, places, and trips.

**FR-051 — Recents.** Bounded, clearable, local.

**FR-052 — Degraded mode.** Saved entities, latest downloaded static schedules, cached route shapes/stops, recent data marked with age, and legal/credits remain available. Realtime failures do not silently replace stale data.

### Notifications

**FR-060 — Contextual opt-in.** Route/stop alert changes, departure reminders, and later arrival/vehicle thresholds. Ask permission only after a user enables one.

**FR-061 — Anonymous installation.** Random installation ID and push token; no advertising ID/account.

**FR-062 — Controls.** Quiet hours, time zone, expiration, one-shot behavior, deletion.

### Credits/legal

**FR-070 — Credits.** Verified TriMet attribution (including its Streetcar-provided coverage), Mapbox, Expo, React Native, `@rnmapbox/maps`, Go, GTFS, and material open-source attribution; privacy, terms, and licenses. Do not imply a Rose City partnership.

**FR-071 — Map compliance.** Preserve Mapbox logo/attribution unless an approved equivalent is compliant; expose required telemetry/anonymous usage control.

## Non-functional requirements

### Accessibility

- Text/list alternative for every map flow.
- VoiceOver/TalkBack labels include route, direction, state, and freshness.
- Dynamic type/font scale, reduced motion, adequate targets and contrast.
- Color never the only signal.
- Both platforms manually tested before release.

### Initial performance budgets

Measured on a representative mid-range supported device:

- interactive shell target under 3 seconds on a warm network;
- vehicle map target 55+ FPS at expected fleet size;
- first vehicle layer render under 1 second after response;
- nearby API p95 under 500 ms excluding uncached upstream time;
- cacheable public API p95 under 400 ms;
- crash-free sessions at least 99.5% before broad rollout.

### Reliability/freshness

- Every realtime response exposes timestamps/source.
- Source failures are isolated.
- Static data survives realtime failure.
- Stale-while-revalidate is explicit.
- No fabricated realtime estimates.

### Privacy/security

- Precise location stays on device except user-requested nearby/search/plan operations.
- Production logs omit raw coordinates/search text by default.
- No background location for MVP.
- Users can clear local history and delete notification registration.
- TriMet AppID and privileged Mapbox tokens stay server/CI-side.
- Public Mapbox mobile token has minimum scopes/restrictions.
- Production deployment uses a narrowly scoped SSH deploy identity, immutable container images, protected GitHub environments, and no interactive root login.

## MVP

1. Nearby stops.
2. Stop arrivals and static schedules.
3. Route details/shapes.
4. Alerts.
5. System vehicle map and filters.
6. Vehicle ID search/detail.
7. TriMet-backed A-to-B planner.
8. Favorites and bounded offline cache.
9. Credits/privacy/attribution.
10. Production CI/CD, monitoring, backups, and store release.

Specialist views, push, history, and OpenTripPlanner are post-MVP unless approvals and capacity make them low-risk. TriMet-provided Streetcar coverage follows the normal TriMet source gate.

## Release gate

Blocked until MVP acceptance tests pass; no P0/P1 defects; source terms/attribution are documented; Mapbox token/billing controls are active; accessibility/privacy/load/failover/restore/store rehearsals pass; and each upstream source has a verified degraded-state test.
