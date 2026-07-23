# Implementation Status

Last updated: 2026-07-23 (Phase 2 rider-information completion and Phase 6 local hardening integrated)

## Repository baseline

- Repository: newly initialized `main` branch with remote `origin` configured.
- Plan source: `tabi-implementation-plan.zip`, preserved at repository root.
- Extracted plan: `docs/implementation-plan/`.
- Existing approved ADRs: none found at assessment time.
- External account, credential, legal, and host status: unknown; no status is inferred.
- Workstation evidence decision (2026-07-23): the user has directed implementation to proceed without locally installed Android/iOS development builds or Maestro runs. Those device gates remain **unverified**, not passed, and do not authorize a production-release claim.

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

| State                      | Work packages                                                                                                                     | Notes                                                                                                                                                                                                                                            |
| -------------------------- | --------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Completed                  | WP-00–03, WP-08, WP-13, WP-16, WP-17, transit-data/PostGIS spike, Compose spike                                                   | WP-03 migrations/persistence and Phase 1 deployment/quality foundations are integrated.                                                                                                                                                          |
| Completed                  | WP-00–06, WP-08, WP-13, WP-16, WP-17, transit-data/PostGIS spike, Compose spike                                                   | Static import and realtime vehicle snapshot foundations are fixture-proven; real sources remain disabled.                                                                                                                                        |
| Completed                  | WP-00–09, WP-13, WP-14, WP-16, WP-17, transit-data/PostGIS spike, Compose spike                                                   | Public API uses PostgreSQL catalog/current-vehicle reads when configured; CI validates deterministic quality and contract drift without secrets.                                                                                                 |
| Completed                  | Vertical-slice integration and QA                                                                                                 | Fixture contract, stale/source-unavailable representation, Compose topology, mobile fixture mode, and payload construction are verified without provider calls.                                                                                  |
| Completed / evidence-gated | WP-10 Rider Information                                                                                                           | Timezone-aware, after-midnight arrivals; transactional current-state writers; durable bounded favorites/recents; and JSON offline-artifact default are integrated. Native SQLite evidence remains gated.                                         |
| Completed / evidence-gated | WP-11 search/planner foundation                                                                                                   | Provider-neutral OpenAPI, fail-closed API/application boundaries, and fixture-only accessible mobile planning UI are integrated; live planning/search remain gated.                                                                              |
| Completed                  | WP-15 non-host operations, WP-17 security/privacy guardrails                                                                      | Host/backup/metrics and legal evidence remain unverified.                                                                                                                                                                                        |
| Completed / evidence-gated | WP-12 Notifications                                                                                                               | Contract, fixture-only mobile settings, encrypted-token persistence, leased PostgreSQL delivery/receipt state, and a disabled worker are integrated. Real permission, token registration, provider delivery, and receipts remain evidence-gated. |
| Blocked / evidence-gated   | WP-18                                                                                                                             | Phase 5 is stopped at D-002/D-003: no canonical source, written rights, terms, or approval is recorded for Streetcar or Rose City.                                                                                                               |
| Blocked / evidence-gated   | WP-06 real TriMet access, optional sources, production-host deployment, physical RNMapbox proof, mobile RNTL/Vitest harness proof | Implement interfaces and fixtures without credentials; do not scrape.                                                                                                                                                                            |

## Decisions awaiting evidence

- D-001 TriMet AppID, terms, rate/cache/attribution requirements.
- D-002/D-003 Streetcar and Rose City source rights; both remain disabled.
- D-004 Mapbox Search/Geocoding terms, tokens, budgets, and storage rules.
- D-010 Expo/RNMapbox/native SDK compatibility on physical builds.
- D-011 Vitest and React Native Testing Library compatibility proof.
- D-012 vehicle JSON versus GeoJSON hot-path benchmark.
- D-016 Linux host/provider, DNS, backup target, and SSH/VPN policy.
- D-017 Expo project/push credential ownership and isolated delivery/receipt evidence.
- D-018 notification retention, consent copy, subscription limits, and deletion-policy approval.

## Phase 0 validation evidence

- Passed: `make format-check`, `make lint`, `make typecheck`, `make test`, `make test-race`, `make build`, `make doctor`, OpenAPI structural validation/generation/drift checks, Compose topology validation, Go GTFS/GTFS-RT/PostGIS spike tests, and shell syntax checks.
- Passed: `corepack pnpm --dir apps/mobile typecheck`. Expo printed the expected public config with a feature-disabled empty public Mapbox token, but its command exited when this environment denied creation of `$HOME/.expo`; rerun it in a writable developer/CI home.
- Passed: an isolated real PostGIS geography query returned nearest bus and light-rail rows with `limitPerMode=1`; GTFS service time `25:15:30` remained `90930` seconds.
- Passed: WP-09 mobile fixture mode, runtime Zod validation, ETag-aware remote repository boundary, foreground-only polling configuration, mode/freshness filters, exact-ID-first search, source/freshness detail, accessible list alternative, and one fleet plus selected-vehicle `ShapeSource` update path. `corepack pnpm --dir apps/mobile typecheck`, `test` (6 files/10 tests), and `expo config --type public` passed on 2026-07-23.
- Passed: fixture-only vertical-slice contract validation, OpenAPI structural validation and formatted generated-client drift check, public API/config Go tests, mobile Vitest (6 files/10 tests), Compose topology validation, and synthetic GeoJSON payload construction. At 1,000/1,500/3,000 vehicles, payloads measured 185,871/279,896/561,971 bytes and 0.93/1.11/2.51 ms build/stringify on this host; these are not native render-performance claims.
- Passed: Rider Information OpenAPI contract (zero Redocly warnings), PostgreSQL-backed nearby/stop/route/shape/route-vehicle endpoints, deterministic mobile rider-information fixtures/screens (9 files/17 pure tests), hardened non-root runtime image build, Compose secrets wiring, CI workflow validation, and full root validation. Arrivals and schedules return an explicit source-unavailable state until service-calendar/trip-update storage is implemented.
- Passed: GTFS calendar/calendar-date preservation, after-midnight `90930` service-time migration proof, safe GTFS-RT trip-update/alert parsing, calendar-aware static schedule API, alert listing/detail API with explicit unknown freshness, disabled TriMet planner adapter, fixture privacy guard, and fail-safe backup/restore operations checks. Arrival estimates remain unavailable until agency timezone metadata and transactional trip-update writes exist.
- Passed: Phase 3 provider-independent foundations: executable search, reverse-geocode, and journey-plan contracts; generated-client drift checks; explicit feature-disabled API responses and config flags; fail-closed application gateways with deterministic constraint/ranking tests; and fixture-only accessible mobile picker, itinerary timeline, location-denied state, and opaque-ID deep links. No provider request, token, precise-location persistence, device build, or Maestro run occurred.
- Passed: Phase 4 notification foundations: OpenAPI-first installation, push-token, and subscription contracts with a generated client; disabled-before-decoding API handlers and config flag; validated installation/subscription/quiet-hour/expiry logic; an additive encrypted-token/delivery/receipt schema; unique dedupe and leased worker-claim queries; AES-GCM token boundary; deterministic no-op worker policy; and an accessible fixture-only mobile settings flow with safe deep links and background-work limits. No OS permission, real token, push request, receipt, provider credential, device build, or Maestro run occurred.
- Passed: Phase 4 provider-free completion: PostgreSQL-backed leased delivery claims, guarded state transitions, encrypted-token decryption only after a claim, invalid-token receipt handling, and fail-closed runtime validation are covered by unit tests. There is still no provider gateway, real receipt poll, OS permission, real token, Compose activation, or delivery claim.
- Passed: Phase 2 completion: GTFS agency IANA timezone validation/persistence; calendar-exception-aware, after-midnight schedule/arrival derivation; transactional vehicle and trip-update current-state writers with invalid/empty preservation; disabled-by-default allowlisted poller composition; durable bounded local favorites/recents with explicit session-only fallback; and ADR-0012's validated JSON static-artifact default. The isolated PostGIS suite applied all migrations including `000007` twice and exercised sanitized fixtures. No real source endpoint, credential, device SQLite execution, or production database was used.
- Passed: Phase 6 locally provable hardening: candidate release images must be immutable; deployment promotion retains the active release file until readiness passes; a deterministic mocked regression test verifies failed-health rollback and successful promotion. This is not a host, DNS, backup, restore, signing, or production deployment rehearsal.
- Not passed / open: React Native Testing Library with this Vitest dependency pair reaches an untransformed React Native Flow entrypoint; pure logic remains in Vitest under ADR-0002. Native map, SQLite, Maestro, iOS, and Android device gates have not run. Caddy binary validation and systemd verification also require a suitable host environment.
- Passed: OpenAPI lint, structural validation, and generated-client drift checks complete without warnings.
- Passed in the approved network environment: pinned sqlc v1.30.0 generation and `go test ./...`; ordinary sandbox `make generate` cannot reach `proxy.golang.org` and is an environment network gate, not a generated-code drift failure.

## Next integration milestone

All currently identified non-device, non-provider, non-host implementation work in Phases 2, 3, 4, and 6 is integrated. Phase 3 live providers remain gated by D-001 and D-004; Phase 4 real notification permission, delivery, and receipts remain gated by D-017 and D-018. Phase 5 is blocked at D-002/D-003; only a source-contract evidence review is permitted, with no provider contact, fetch, fixture, or scraper. Phase 6 host deployment remains gated by D-016. See `docs/BLOCKERS.md`. The physical development-build demonstration (restricted public Mapbox token, Android and iOS devices, and Maestro) is explicitly waived as a local-workstation blocker by the user, but remains unverified and required before any production-release claim.
