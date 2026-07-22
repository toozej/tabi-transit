---
doc_id: TAB-PLAN-012
title: "Testing and Quality Strategy"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["qa-agent", "mobile-agent", "backend-agent", "data-agent", "security-agent"]
depends_on: ["TAB-PLAN-001", "TAB-PLAN-003"]
---


# Testing and Quality Strategy

## Pyramid

Static/schema checks → TypeScript/Go unit/property/fuzz → React Native components → API/DB/provider contracts → Maestro E2E → manual accessibility/field validation.

## TypeScript

### Vitest

Required for domain/filter/ranking/freshness/time/service-day/GeoJSON/search normalization/repository mapping/notification policy/Zod/API error logic. No Jest suite for these packages.

### React Native components

React Native Testing Library with a Vitest-compatible harness proven in Phase 0. Test loading/empty/error/stale, permission denial, labels/roles, arrival rows, filters, planner forms, alerts/sheets, and navigation intent.

Do not JS-test native Mapbox rendering; Vitest tests layer/source builders and Maestro tests native behavior. If the harness is incompatible, write an ADR; do not silently add a duplicate Jest suite. Keep logic in Vitest and move native assertions to Maestro/platform tests.

## Go

- Table unit tests with injected source/clock/store/sender.
- `httptest` with OpenAPI validation, middleware, cancellation, ETag, rate limits.
- Real PostGIS integration for migrations, geospatial, activation, snapshots, dedupe, retention.
- Race detector for concurrent packages.
- Fuzz GTFS-RT, CSV/time, public IDs, filters, upstream mappers, alert normalization.

## Provider/data contracts

Sanitized deterministic fixtures are primary. Low-rate scheduled real-upstream smoke detects schema/content/auth/empty/timestamp changes.

GTFS cases: normal/change/after-midnight/calendar exception/bad reference/empty essential/malicious ZIP/invalid coordinate/rollback. Realtime: stale/empty/malformed/deleted/unmatched.

## Maestro critical flows, both platforms

1. launch/migrate;
2. deny location and use stop search;
3. grant location and nearby;
4. nearest two per selected mode;
5. filter live map;
6. exact vehicle ID and detail;
7. route/schedule;
8. alert;
9. A-to-B filters;
10. save and offline;
11. notification create/disable and deep link when enabled;
12. clear cache/history;
13. credits/privacy/Mapbox telemetry.

PR smoke subset; staging/release full suite.

## Map performance

Synthetic fleet at 1x, 1.5x, 3x. Measure source update, JS/native memory, frames, selection/filter latency, polling responsiveness on physical mid-range Android and supported iPhone.

## Load tests

Use k6/equivalent for vehicle ETag hits/misses, nearby, commute arrivals, routes/schedules, alerts, constrained planner, and autocomplete limits. Include slow/down upstream; derive capacity/autoscaling.

## Accessibility

Automated semantics/lint/contrast where possible. Manual VoiceOver, TalkBack, max text, reduce motion, color/contrast, and map-independent completion. Blocking accessibility defects are release-severity.

## Security/failure tests

Secret/dependency/container/SBOM/license/Dockerfile/Compose scans; API input abuse; DAST staging; installation ownership; push relay; log redaction; binary secret inspection.

Failure injection: each TriMet feed/API down/stale/malformed, Mapbox rate limit, database issue, full disk, container restart, host reboot, backup destination failure, Caddy/TLS failure, push failure, mobile offline/slow, old static feed, and clock skew.

## Coverage/defects

Targets: pure TS 85%; Go domain/application/mappers 80%; critical freshness/ranking/dedupe near-complete branches. Generated excluded. Production bugs get regression tests.

P0 safety/privacy/security/corruption; P1 core blocked or materially false realtime; P2 major workaround; P3 minor. No P0/P1 release.

## Deliverables

Device/test matrix, fixtures, automated reports, map/performance baseline, accessibility report, capacity/failure report, release checklist.
