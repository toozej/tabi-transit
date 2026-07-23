# Implementation Status

Last updated: 2026-07-22 (Phase 0 integrated)

## Repository baseline

- Repository: newly initialized `main` branch with remote `origin` configured.
- Plan source: `tabi-implementation-plan.zip`, preserved at repository root.
- Extracted plan: `docs/implementation-plan/`.
- Existing approved ADRs: none found at assessment time.
- External account, credential, legal, and host status: unknown; no status is inferred.

## Dependency board

```text
WP-00 architecture ─┬─> WP-01 repository bootstrap ─┬─> WP-08 mobile foundation
                    └─> WP-02 OpenAPI contract ────┬─> WP-03 database
                                                     ├─> WP-06 TriMet adapter (AppID-gated)
                                                     └─> WP-07 public API
WP-03 ─> WP-04 GTFS importer ─> WP-05 GTFS-Realtime poller ─> WP-07
WP-02 + WP-03..06 + WP-08 ─> WP-09 live-map vertical slice
WP-13 infrastructure is independently gated by the host ADR.
```

## Work packages

| State                    | Work packages                                                                                                                     | Notes                                                                 |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| Completed                | WP-00–03, WP-08, WP-13, WP-16, WP-17, transit-data/PostGIS spike, Compose spike | WP-03 migrations/persistence and Phase 1 deployment/quality foundations are integrated. |
| Completed                | WP-00–04, WP-06, WP-08, WP-13, WP-16, WP-17, transit-data/PostGIS spike, Compose spike | GTFS importer has a fixture-proven transactional activation flow and static ID mapping. |
| Next unblocked           | WP-05 GTFS-Realtime poller | Uses WP-04's source-qualified stop/route/trip mapping and preserves last valid snapshots. |
| Pending                  | WP-07 through WP-18 | Start only when their stated dependencies and Phase gates are met. |
| Blocked / evidence-gated | WP-06 real TriMet access, optional sources, production-host deployment, physical RNMapbox proof, mobile RNTL/Vitest harness proof | Implement interfaces and fixtures without credentials; do not scrape. |

## Decisions awaiting evidence

- D-001 TriMet AppID, terms, rate/cache/attribution requirements.
- D-002/D-003 Streetcar and Rose City source rights; both remain disabled.
- D-004 Mapbox Search/Geocoding terms, tokens, budgets, and storage rules.
- D-010 Expo/RNMapbox/native SDK compatibility on physical builds.
- D-011 Vitest and React Native Testing Library compatibility proof.
- D-012 vehicle JSON versus GeoJSON hot-path benchmark.
- D-016 Linux host/provider, DNS, backup target, and SSH/VPN policy.

## Phase 0 validation evidence

- Passed: `make format-check`, `make lint`, `make typecheck`, `make test`, `make test-race`, `make build`, `make doctor`, OpenAPI structural validation/generation/drift checks, Compose topology validation, Go GTFS/GTFS-RT/PostGIS spike tests, and shell syntax checks.
- Passed: `corepack pnpm --dir apps/mobile typecheck`. Expo printed the expected public config with a feature-disabled empty public Mapbox token, but its command exited when this environment denied creation of `$HOME/.expo`; rerun it in a writable developer/CI home.
- Passed: an isolated real PostGIS geography query returned nearest bus and light-rail rows with `limitPerMode=1`; GTFS service time `25:15:30` remained `90930` seconds.
- Not passed / open: mobile RNTL/Vitest execution fails during React Native dependency transformation; native map, SQLite, Maestro, iOS, and Android device gates have not run. Caddy binary validation and systemd verification also require a suitable host environment.
- OpenAPI lint completes with warnings about component-example reference shape; structural validation and generated-client drift pass. Resolve those warnings before using examples as formal response-conformance fixtures.
- Passed in the approved network environment: pinned sqlc v1.30.0 generation and `go test ./...`; ordinary sandbox `make generate` cannot reach `proxy.golang.org` and is an environment network gate, not a generated-code drift failure.

## Next integration milestone

Integrate and validate WP-04 GTFS importer. WP-05 realtime poller starts after WP-04 provides the required static identifier-mapping contract.
