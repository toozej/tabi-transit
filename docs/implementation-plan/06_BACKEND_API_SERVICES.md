---
doc_id: TAB-PLAN-006
title: "Backend API and Service Implementation Plan"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["backend-agent", "api-agent", "reliability-agent"]
depends_on: ["TAB-PLAN-002", "TAB-PLAN-003", "TAB-PLAN-007", "TAB-PLAN-008", "TAB-PLAN-017"]
---


# Backend API and Services

## Objective and stack

Production Go services protect credentials, normalize upstreams, expose a stable mobile API, perform geospatial queries, and isolate provider outages while running as separate Docker Compose services on one Linux host.

Use current pinned Go, `net/http` + Chi, `pgx`, `sqlc`, PostgreSQL/PostGIS, `slog`, OpenTelemetry, Prometheus-compatible metrics, MobilityData GTFS-Realtime Go bindings, and OpenAPI generation.

## Package direction

```text
api/jobs -> application -> domain
api/jobs -> persistence/source interfaces
persistence/sources -> domain
domain -> no infrastructure imports
```

Packages: `domain`, `application`, `api`, `persistence`, `sources`, `jobs`, `config`, `observability`.

## HTTP middleware order

Trusted proxy/IP handling → request ID → panic recovery → structured access log → trace context → compression → security headers → rate limit → body size limit → OpenAPI validation → handler → normalized error/metrics.

Never log exact coordinates, full search queries, AppIDs/tokens, or push tokens.

## Endpoints

- `/health/live`, `/health/ready`; private `/metrics`.
- `/v1/config`.
- `/v1/stops`, `/routes`, `/vehicles`, `/alerts`, `/schedules`.
- `/v1/search`, `/v1/journeys`.
- `/v1/installations`, `/v1/subscriptions`.

See `17_API_CONTRACTS.md`.

## Source adapter contract

Each adapter has typed request/response mapping, auth injection, explicit base URL, timeouts/retry policy, error classification, fixture tests, source/attribution metadata, and call/latency/error/last-success metrics.

### TriMet

Adapters for arrivals, route config, stop location, vehicle locations/detail, trip status, block status, and trip planner. Beta endpoints are feature-flagged and contract-tested.

### GTFS-Realtime

Fetch bounded protobuf bytes, validate feed header/timestamp, normalize vehicles/trip updates/alerts, preserve unknown values/IDs, reject impossible coordinates with diagnostics, and support differential semantics only if correct.

### Mapbox

Normalize search, geocoding, POIs, proximity, attribution, and terms-aware cache; meter rate/cost.

### Streetcar/Rose City

Disabled until source/rights decision. Interface/fixtures can be prepared; no production scraper.

## Caching

Start with bounded in-process reference caches, PostgreSQL current snapshots, HTTP ETags, and short terms-aware upstream caches. No Redis until multiple replicas, distributed dedupe/locks, or measured hot-data need.

TTLs are configured by endpoint/source. Any stale response includes age/state.

## Poller

Independent jittered loops: fetch → parse → validate → transform → transactional snapshot write → update source status → metrics → sleep/backoff. Use one desired task plus database advisory lock or equivalent leader guard. Failed cycles preserve last valid snapshot.

## Response efficiency

- ETag vehicle snapshots and static resources.
- Support system-wide and filtered/viewport vehicle queries.
- Avoid N+1 enrichment.
- Benchmark GeoJSON versus domain JSON before freezing the hot path.
- Paginate large schedules/history.
- Compression over cryptic field names.

## Validation

Bound coordinates/radius/limits; validate IDs/enums/date/time zone/quiet hours; limit strings/control characters; never accept an arbitrary upstream URL.

## Errors

- 400 `validation_error`.
- 404 `not_found`.
- 409 `conflict`.
- 429 `rate_limited`.
- 503 `source_unavailable` or `temporarily_unavailable`.
- 500 `internal_error`.

Include request ID and safe details; never relay raw provider bodies.

## Installations/subscriptions

Issue anonymous installation credential, upsert token rotation, delete installation/tokens/subscriptions, validate referenced entities, cap subscriptions, and avoid precise location in payloads. Protect mutations with installation authentication.

## Lifecycle

Each image includes a dependency-free `/app/tabi-healthcheck` binary or equivalent for Compose health checks. Separate liveness/readiness; handle SIGTERM; stop intake, cancel loops, finish bounded transactions, flush telemetry, exit in grace period. Verify migrations/config on startup but tolerate optional source outages.

## Acceptance

OpenAPI conformance; fixtures/failure tests for adapters; race and fuzz pass; scoped outages do not cause global failure; rate limits and log redaction work; spatial queries meet budgets; invalid/empty feed cannot replace valid state; graceful shutdown and secret handling verified.
