# Implementation Status

Last updated: 2026-07-23 (Phase 1 vertical-slice implementation integrated)

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

| State                    | Work packages                                                                                                                     | Notes                                                                                                                                                                            |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Completed                | WP-00–03, WP-08, WP-13, WP-16, WP-17, transit-data/PostGIS spike, Compose spike                                                   | WP-03 migrations/persistence and Phase 1 deployment/quality foundations are integrated.                                                                                          |
| Completed                | WP-00–06, WP-08, WP-13, WP-16, WP-17, transit-data/PostGIS spike, Compose spike                                                   | Static import and realtime vehicle snapshot foundations are fixture-proven; real sources remain disabled.                                                                        |
| Completed                | WP-00–09, WP-13, WP-16, WP-17, transit-data/PostGIS spike, Compose spike                                                          | Fixture-backed public API and mobile vehicle map/list/search/detail flow are integrated; catalog DB composition remains an explicit unavailable state until query adapters land. |
| Completed                | Vertical-slice integration and QA                                                                                                 | Fixture contract, stale/source-unavailable representation, Compose topology, mobile fixture mode, and payload construction are verified without provider calls.                  |
| Pending                  | WP-10 through WP-18                                                                                                               | Start only when their stated dependencies and Phase gates are met.                                                                                                               |
| Blocked / evidence-gated | WP-06 real TriMet access, optional sources, production-host deployment, physical RNMapbox proof, mobile RNTL/Vitest harness proof | Implement interfaces and fixtures without credentials; do not scrape.                                                                                                            |

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
- Passed: WP-09 mobile fixture mode, runtime Zod validation, ETag-aware remote repository boundary, foreground-only polling configuration, mode/freshness filters, exact-ID-first search, source/freshness detail, accessible list alternative, and one fleet plus selected-vehicle `ShapeSource` update path. `corepack pnpm --dir apps/mobile typecheck`, `test` (6 files/10 tests), and `expo config --type public` passed on 2026-07-23.
- Passed: fixture-only vertical-slice contract validation, OpenAPI structural validation and formatted generated-client drift check, public API/config Go tests, mobile Vitest (6 files/10 tests), Compose topology validation, and synthetic GeoJSON payload construction. At 1,000/1,500/3,000 vehicles, payloads measured 185,871/279,896/561,971 bytes and 0.93/1.11/2.51 ms build/stringify on this host; these are not native render-performance claims.
- Not passed / open: React Native Testing Library with this Vitest dependency pair reaches an untransformed React Native Flow entrypoint; pure logic remains in Vitest under ADR-0002. Native map, SQLite, Maestro, iOS, and Android device gates have not run. Caddy binary validation and systemd verification also require a suitable host environment.
- OpenAPI lint completes with warnings about component-example reference shape; structural validation and generated-client drift pass. Resolve those warnings before using examples as formal response-conformance fixtures.
- Passed in the approved network environment: pinned sqlc v1.30.0 generation and `go test ./...`; ordinary sandbox `make generate` cannot reach `proxy.golang.org` and is an environment network gate, not a generated-code drift failure.

## Next integration milestone

Phase 1 implementation and fixture QA are complete. Its required physical development-build demonstration (restricted public Mapbox token, Android and iOS devices, and Maestro) remains a blocking exit gate and cannot be claimed from this environment. Do not start Stage 6 until that gate has passed or an approved roadmap/ADR decision changes it.
