---
doc_id: TAB-PLAN-000
title: "Tabi Implementation Plan — Agent Entry Point"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["technical-lead", "program-manager", "all-agents"]
depends_on: []
---


# Tabi Implementation Plan

This directory is the implementation source of truth for **Tabi**, a modern Portland-area transit application for iOS and Android.

## Selected stack

- React Native with Expo development builds and TypeScript.
- Expo Router, TanStack Query, Zustand, Zod, React Hook Form, Expo SQLite, SecureStore, Location, Notifications, and Background Task.
- `@rnmapbox/maps`, the community React Native wrapper over Mapbox native Maps SDKs.
- Go services using `net/http`, Chi, `pgx`, `sqlc`, PostgreSQL, and PostGIS.
- TriMet official Web Services, GTFS Schedule, and GTFS-Realtime as the primary transit sources.
- GitHub Actions, EAS Build/Submit, Docker Engine, Docker Compose, Caddy, and a single Linux server as the reference deployment.
- Vitest for framework-independent TypeScript; React Native Testing Library with a Vitest-compatible harness proven in Phase 0; Maestro for device E2E.
- OpenAPI-first client/server contracts.

Rose City Transit and Portland Streetcar integrations are gated by source, licensing, and permission verification. No scraping is included by default.

The canonical production deployment is a single Linux Docker host. Files under `deployment/` are implementation templates for Compose, Caddy, systemd timers, backup, restore, and deployment. Fly.io is optional and is not the canonical production control plane.

## How agents use the package

1. Read this file, `01_PRODUCT_REQUIREMENTS.md`, `02_SYSTEM_ARCHITECTURE.md`, and `03_REPOSITORY_ENGINEERING_STANDARDS.md`.
2. Read `17_API_CONTRACTS.md`.
3. Read the assigned component plan and `16_AGENT_WORK_PACKAGES.md`.
4. Check `18_OPEN_QUESTIONS_DECISIONS.md` before depending on an external source or unresolved technology.
5. Deliver code, tests, contracts, migrations, observability, runbooks, privacy/security updates, and rollback notes together.
6. Record changes to an architectural assumption in an ADR under `docs/adr/`.

## Document map

| File | Purpose |
|---|---|
| `01_PRODUCT_REQUIREMENTS.md` | Functional and non-functional requirements, MVP, acceptance gates |
| `02_SYSTEM_ARCHITECTURE.md` | Runtime boundaries, deployables, flows, resilience |
| `03_REPOSITORY_ENGINEERING_STANDARDS.md` | Monorepo, toolchain, code and contract rules |
| `04_MOBILE_APPLICATION_PLAN.md` | Expo/React Native implementation |
| `05_MAPS_SEARCH_TRIP_PLANNING.md` | Mapbox, geocoding, nearby and A-to-B |
| `06_BACKEND_API_SERVICES.md` | Go services and source adapters |
| `07_TRANSIT_DATA_INGESTION.md` | GTFS/GTFS-RT import and normalization |
| `08_DATA_MODEL_STORAGE_CACHING.md` | PostgreSQL/PostGIS and SQLite |
| `09_NOTIFICATIONS_BACKGROUND_WORK.md` | Push and background constraints |
| `10_INFRASTRUCTURE_ENVIRONMENTS.md` | Linux host, Docker Compose, storage, networking and backups |
| `11_CI_CD_RELEASE_MANAGEMENT.md` | CI, EAS, backend/store releases |
| `12_TESTING_QUALITY_STRATEGY.md` | Full test strategy |
| `13_SECURITY_PRIVACY_COMPLIANCE.md` | Threats, terms, privacy, licensing |
| `14_OBSERVABILITY_OPERATIONS.md` | SLOs, dashboards, alerts, runbooks |
| `15_DELIVERY_ROADMAP.md` | Ordered phases and exit criteria |
| `16_AGENT_WORK_PACKAGES.md` | Parallel assignments and handoffs |
| `17_API_CONTRACTS.md` | Initial REST contract |
| `18_OPEN_QUESTIONS_DECISIONS.md` | Decision gates and risks |
| `19_REFERENCE_SOURCES.md` | Authoritative documentation |
| `20_LINUX_DOCKER_COMPOSE_RUNBOOK.md` | Concrete installation, deployment, backup and recovery runbook |
| `21_FLY_IO_OPTIONAL_DEPLOYMENT.md` | Optional Fly.io deployment using the same images |

## Precedence

1. Approved ADR.
2. `01_PRODUCT_REQUIREMENTS.md`.
3. Executable OpenAPI and `17_API_CONTRACTS.md`.
4. Component plan.
5. Delivery roadmap.
6. Code comments.

Security, privacy, legal, and accessibility requirements cannot be overridden by schedule pressure.

## Global definition of ready

A work package is ready when dependencies are complete, decisions are resolved or explicitly assumed, contracts/fixtures exist, acceptance criteria are testable, required accounts/secrets are available, and the assigned subsystem runs locally.

## Global definition of done

A work package is done when production code and tests are merged, contracts/generated code are synchronized, migrations are safe, accessibility is validated, observability/runbooks exist, security/privacy impacts are addressed, credits/attribution are current, and rollback or mitigation is documented.

## Shared terminology

- **Realtime**: source data that can be delayed; never synonymous with exact.
- **Scheduled**: static GTFS or timetable data.
- **Freshness**: source/entity/fetch timestamps and a derived state.
- **Mode**: bus, MAX/light rail, WES/commuter rail, streetcar, aerial tram, or configured other.
- **Route**: rider-facing line/service.
- **Trip**: scheduled stop sequence for a service date.
- **Block**: operational assignment of one or more trips.
- **Source adapter**: provider-specific input converted to Tabi domain models.
