# Implementation Status

Last updated: 2026-07-22

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

| State | Work packages | Notes |
|---|---|---|
| Active | WP-00, WP-01, WP-02, WP-08, transit-data spike, Docker Compose spike | Phase 0 parallel work with non-overlapping ownership. |
| Pending | WP-03 through WP-18 | Start only when their stated dependencies and Phase 0 gates are met. |
| Blocked / evidence-gated | WP-06 real TriMet access, optional sources, production-host deployment, physical RNMapbox device proof | Implement interfaces and fixtures without credentials; do not scrape. |

## Decisions awaiting evidence

- D-001 TriMet AppID, terms, rate/cache/attribution requirements.
- D-002/D-003 Streetcar and Rose City source rights; both remain disabled.
- D-004 Mapbox Search/Geocoding terms, tokens, budgets, and storage rules.
- D-010 Expo/RNMapbox/native SDK compatibility on physical builds.
- D-011 Vitest and React Native Testing Library compatibility proof.
- D-012 vehicle JSON versus GeoJSON hot-path benchmark.
- D-016 Linux host/provider, DNS, backup target, and SSH/VPN policy.

## Next integration milestone

Integrate and validate the Phase 0 foundation: reproducible repository bootstrap, executable OpenAPI contract, compilable mobile shell, documented RNMapbox status, GTFS/GTFS-RT/PostGIS proofs, and Compose configuration validation. Then record evidence-based ADRs and commit the coherent foundation.
