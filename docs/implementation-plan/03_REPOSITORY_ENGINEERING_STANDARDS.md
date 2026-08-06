---
doc_id: TAB-PLAN-003
title: "Repository and Engineering Standards"
status: implementation-ready
last_updated: 2026-08-05
intended_agents: ["repo-agent", "all-implementation-agents"]
depends_on: ["TAB-PLAN-002"]
---


# Repository and Engineering Standards

## Monorepo

```text
tabi/
├── apps/mobile/
├── apps/web/
├── services/{transit-api,gtfs-importer,realtime-poller,notification-worker}/
├── internal/{api,application,config,domain,persistence,sources,observability}/
├── packages/{api-client,transit-domain,eslint-config,tsconfig}/
├── api/{openapi.yaml,examples/}
├── db/{migrations,queries,seeds,fixtures}/
├── deployment/
│   ├── compose.yaml
│   ├── compose.production.yaml
│   ├── Caddyfile
│   ├── scripts/
│   ├── systemd/
│   └── fly/
├── tests/{contract,e2e/maestro,load,fixtures/upstream}/
├── docs/{adr,runbooks,product}/
├── .github/workflows/
├── pnpm-workspace.yaml
├── go.work
├── Makefile
└── README.md
```

## Toolchain

Pin a current compatible Expo SDK, React Native, `@rnmapbox/maps`, Mapbox native SDK, Node LTS, pnpm, Go, PostgreSQL/PostGIS, Docker Engine, Docker Compose, Caddy, and restic in Phase 0. Record versions in tool files, lockfiles, Go directives, EAS profiles, image digests, and deployment documentation. No floating production tags.

Use pnpm workspaces, Go modules/`go.work`, `sqlc`, OpenAPI, a pinned Go OpenAPI generator, and `openapi-typescript` (or equivalent) for the generated TypeScript client.

## Stable root commands

```text
make prereqs format lint typecheck test test-unit test-integration
make test-e2e test-race test-load generate generate-check
make db-up db-migrate dev-api dev-mobile dev-poller build doctor
```

CI and agents use these interfaces rather than private package commands.

## Change policy

- Trunk-based, short branches, scoped PRs.
- Breaking API/schema changes require ADR and compatibility/migration plan.
- Feature flags guard incomplete providers/features.
- No manual production drift except emergency action documented afterward.
- Source code, tests, contracts, migrations, runbooks, and privacy/security changes land together.

## TypeScript

- `strict: true`; avoid `any`.
- Zod validates network, persistence, and deep-link input.
- Domain package has no React Native imports.
- API timestamps are ISO 8601; coordinates are `[longitude, latitude]`; distances meters; durations seconds.
- Exhaustive unions include `unknown`.
- Vitest for pure domain/data functions.
- Accessibility semantics before test IDs.

## React Native

- Thin Expo Router files.
- TanStack Query for remote state; no duplicate remote entities in Zustand.
- Zustand only for filters, camera preferences, selection, drafts, transient UI.
- Repository-wrapped SQLite.
- Platform APIs and Mapbox behind adapters.
- Development builds, not Expo Go.
- Custom native modules require ADR and platform tests.

## Web

- `apps/web` is a first-class responsive React application, not an embedded native build.
- Semantic HTML is the accessibility baseline; pointer-only and hover-only interactions are prohibited for required tasks.
- Browser APIs, IndexedDB, map SDK, service worker, notifications, and external links stay behind web platform adapters.
- Share pure domain, API, validation, formatting, query-key, and token code with mobile; do not share platform UI components by default.
- Public web configuration contains no secrets. Production API calls are same-origin; development CORS origins are explicit.

## Go

- Standard library first, Chi for routing.
- Context through all request/job boundaries.
- `slog` structured logs and correlation IDs.
- Inject HTTP clients, clocks, random, stores, senders.
- No package mutable global state.
- Classified wrapped errors.
- Explicit transactions.
- Table tests, `httptest`, race detector, parser fuzzing.
- Reviewed SQL with `pgx/sqlc`; no ORM hiding geospatial/GTFS queries.

## Database

- Migrations append-only after merge.
- Each migration documents purpose, lock risk, ordering, and rollback/forward-fix.
- `timestamptz` for instants.
- PostGIS geography for meter-distance point queries and geometry for shapes as appropriate.
- High-cardinality queries require representative `EXPLAIN`.
- Device/subscription data is separated from public transit data and retention.

## Configuration

Typed environment loader per binary. Categories include database, HTTP, TriMet, Mapbox, source intervals/stale thresholds, archive paths, off-site backup, push, observability, flags, and limits. Every secret supports a `_FILE` form so Docker Compose secrets can be mounted under `/run/secrets`. Startup fails on missing required values. Secrets are never logged.

## API compatibility

- `/v1`.
- Additive fields allowed.
- Removal/semantic change requires new version or compatibility window.
- Safe unknown enums.
- Request IDs and freshness.
- Opaque cursor pagination.
- ETag on snapshots/static resources.

## Documentation/ADRs

Each component README covers run/config/test/failure/attribution. Save sanitized external fixtures. Required ADRs include Linux host sizing/provider, version matrix, Compose/Caddy layout, Mapbox search/storage, TriMet versus OTP, push provider, observability profile, backup repository, history retention, Fly.io optional path, and Vitest/RNTL harness. Streetcar and Rose City-inspired presentation follow the TriMet source ADR boundary.
