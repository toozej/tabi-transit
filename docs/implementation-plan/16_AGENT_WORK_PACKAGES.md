---
doc_id: TAB-PLAN-016
title: "Agent Work Packages and Handoff Contracts"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["orchestrator-agent", "all-agents"]
depends_on: ["TAB-PLAN-000", "TAB-PLAN-003", "TAB-PLAN-015"]
---


# Agent Work Packages

Agents work through contracts, migrations, fixtures, generated clients, and documented adapters. Do not modify another package’s internals without coordination; shared contract changes require API/architecture ownership.

| ID | Owner | Outputs | Dependencies |
|---|---|---|---|
| WP-00 | Architecture | ADRs, version matrix, source/license and risk registers, vocabulary review | Plan |
| WP-01 | Repo/tooling | Monorepo, pnpm/go.work, Makefile, lint/test/generate, Docker, contribution guide | WP-00 |
| WP-02 | API contract | OpenAPI, examples/errors/freshness, Go/TS generation, compatibility tests | WP-00 |
| WP-03 | Database | Migrations, sqlc, schemas/indexes/fixtures, backup notes | WP-02 |
| WP-04 | GTFS static | fetch/archive/validate/load/activate/rollback/reports | WP-03 |
| WP-05 | GTFS-RT | vehicle/trip/alert adapters, snapshots, freshness, fuzz/fixtures | WP-03/04 |
| WP-06 | TriMet WS | arrivals, route/stop, vehicle/trip/block, planner, caches/fixtures | AppID/WP-02 |
| WP-07 | Public API | middleware/endpoints/ETag/spatial/health/load | WP-02–06 |
| WP-08 | Mobile foundation | Expo/EAS/RNMapbox/SQLite/query/state/harness/shell | WP-00/01 |
| WP-09 | Maps/vehicles | adapter/layers/filters/search/detail/list/performance | WP-07/08 |
| WP-10 | Rider info | nearby/stops/routes/schedules/alerts/offline/favorites/accessibility | WP-07/08 |
| WP-11 | Search/planner | Mapbox adapter, picker, planner/ranking, itinerary/deep links | WP-06–08 |
| WP-12 | Notifications | install/subscription schema/API, mobile, worker, receipts/dedupe | WP-03/07/08 |
| WP-13 | Infrastructure | Linux bootstrap, Compose/Caddy, GHCR pull, PostGIS volumes, systemd jobs, secrets, firewall, backup/restore, optional Fly configs | Host ADR |
| WP-14 | Release | GitHub workflows, EAS, image signing/SBOM, deploy/migration/rollback | WP-01/08/13 |
| WP-15 | SRE | OTel/logs, dashboards/alerts/SLOs/synthetics/runbooks | WP-07/13 |
| WP-16 | QA/accessibility | matrix, Maestro, load/failure, manual reports, gates | Incremental |
| WP-17 | Security/privacy | threat model, terms, scans/redaction, store forms/notices | Inventory |
| WP-18 | Optional sources | written permission, source contract, adapters/dedup/credits | Approval |

## Detailed handoff expectations

Every package supplies:

- implementation and tests;
- README/local commands;
- configuration and secret requirements;
- OpenAPI/schema/migration changes;
- deterministic fixtures;
- metrics/dashboard/runbook changes;
- security/privacy/licensing impact;
- known limits;
- deployment and rollback notes;
- evidence for acceptance criteria.

## Integration cadence

- Review contract before implementation.
- Check generated drift daily/in CI.
- Produce one vertical slice per phase.
- Integrate provider/feature in staging after merge.
- Avoid long-lived cross-component private branches.
- Orchestrator tracks dependency graph, ADR blockers, and feature flags.

## Agent-specific hard boundaries

- Mobile agents do not call provider APIs directly.
- Source agents do not expose provider payloads as public contracts.
- Infra agents do not choose retention/privacy policy.
- Infrastructure work must keep the single-host Compose deployment canonical; Fly.io files are optional adapters.
- QA agents can block release for unmet product/security/accessibility gates.
- Optional-source agent cannot deploy a scraper without written ADR approval.
- Generated code is not hand-edited.
