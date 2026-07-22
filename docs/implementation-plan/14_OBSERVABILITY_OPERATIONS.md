---
doc_id: TAB-PLAN-014
title: "Observability and Operations"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["sre-agent", "backend-agent", "mobile-agent", "support-agent"]
depends_on: ["TAB-PLAN-006", "TAB-PLAN-010", "TAB-PLAN-013"]
---


# Observability and Operations

## Stack/correlation

OpenTelemetry, Prometheus-compatible metrics, JSON `slog`, selected mobile crash/error provider, optional self-hosted dashboards, Docker/host metrics, and external uptime alerts. Never collect precise coordinates/full queries/tokens.

API request IDs, mobile safe correlation IDs, job/run IDs, feed/snapshot versions, and notification delivery IDs link events.

## Metrics

### HTTP

Count/status/route, latency, concurrency, size, rate limit, ETag/cache hit, panic.

### Sources/data

Fetch status/latency/last success, feed/entity age, size/count, parse/validation, empty/unmatched/invalid coordinate, timestamp regression, circuit/quota, import/activation.

### Database/jobs

Pool/query/slow/storage/migration/backup; poll lag; import duration; notification evaluation/send/receipt; retention and locks.

### Mobile

Crash/ANR/start, API error class, screen performance, map update/render failure, SQLite migration/recovery, notification registration/open, privacy-safe success/failure only.

## Initial SLOs

- Public API availability 99.9% monthly excluding separately classified upstream-only degradation.
- Cacheable p95 under 400 ms; nearby p95 under 500 ms.
- 5xx below 0.5%.
- Successful realtime poll cycles above 99%.
- Snapshot age within source target for 99% of healthy-source time.
- Mobile crash-free sessions at least 99.5%; launch success 99.8%.

## Alerts

Page: API/DB down, missing active static feed, security/billing, runaway push, core credential failure.

Ticket: optional source stale, unmatched growth, latency/storage/cost trends, import regression.

Every alert links a tested runbook and is deduplicated.

## Dashboards/runbooks

Dashboards: rider/API, source freshness, GTFS/data quality, DB/cache, mobile stability/map, notifications, host CPU/RAM/disk/container restarts, backup age, release comparison.

Runbooks: TriMet WS outage; each GTFS-RT stale/malformed/empty; GTFS rollback; DB restore; full disk; host reboot/loss; Docker/Compose failure; Caddy/TLS; backup failure; latency; Mapbox quota/token/billing; EAS/store; push; bad OTA; optional source disable; credential rotation.

## Synthetics and support

External checks for health/config, known stop/route, vehicle freshness when service active, alerts, and low-rate planner. Never rely permanently on one specific vehicle.

Diagnostics screen may expose app/API/feed versions, source statuses, platform, anonymized installation suffix, permissions, and cache size—never tokens or precise location.

## Incident process

Detect/classify → incident commander → protect users/data → disable source/feature → communicate → restore → verify freshness → postmortem/regression.

## Acceptance

Dashboards before launch; every page alert has runbook; source freshness independent; release annotations; restore/feed rollback rehearsal; no forbidden telemetry; SLO distinguishes Tabi and upstream while preserving user-impact view.
