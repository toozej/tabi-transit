# MVP threat model and control register

This register records current implementation evidence. A control marked as a
gate is not a claim that it has been deployed, legally reviewed, or tested in a
production environment.

| Threat | Current control/evidence | Remaining gate |
| --- | --- | --- |
| Credential exposure | `.gitignore`, safe examples, `_FILE` configuration support, repository/fixture signature checks | Secret scanning in CI, rotation rehearsal, production secret ownership review |
| Precise location or search leakage | Raw-coordinate logging check; fixture policy rejects explicit user-location/search fixtures | Request-log and analytics review once production observability is selected |
| Public database/admin exposure | Compose policy check rejects published Postgres/admin/metrics ports and Docker socket mounts | Actual host firewall, ownership, and Compose runtime inspection |
| API abuse or malformed input | Boundary validation, request-size/rate-limit tests in component packages | Staging abuse/DAST evidence and capacity limits |
| Malformed or stale transit feeds | Fixture parsers reject invalid data; freshness contract prevents replacement of last valid snapshot | Approved upstream smoke and operational alert thresholds |
| False live-map claims | Freshness status is part of the contract and fixture flow | Android/iOS native rendering and accessibility evidence |
| Mapbox account/cost misuse | Native token guidance uses a separate least-privilege public token; no token is committed | Account scopes, budgets, telemetry, attribution, and terms review (D-004) |
| Supply-chain/license risk | Pinned project dependencies and local checks | Pinned dependency, license, image, SBOM, and provenance scanning in CI |
| Optional-source rights violation | ADR-0005 keeps sources disabled and forbids scraping | Written source/terms/attribution approval |

## Evidence review cadence

Review this register when adding a provider, public endpoint, persistence of
user/device data, analytics, maps/search capability, or a production deployment.
Every new externally sourced fixture must comply with the fixture policy.

The following remains intentionally unresolved: D-001 (TriMet terms/AppID),
D-004 (Mapbox product/terms/storage/telemetry/budget), D-016 (host/backup/DNS),
and physical Android/iOS/Maestro evidence. None is inferred from local source
files or a configured local environment.
