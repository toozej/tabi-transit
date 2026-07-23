# Failure-injection catalog

This catalog is a test-plan scaffold, not evidence that these failures have
been exercised. Use deterministic fixtures and injected clocks/HTTP clients
instead of live provider calls.

| Failure                                                | Expected invariant                                              | Planned owner/evidence                 |
| ------------------------------------------------------ | --------------------------------------------------------------- | -------------------------------------- |
| GTFS/GTFS-RT timeout, malformed, empty, or stale input | last valid snapshot remains available and freshness is explicit | WP-04/WP-05 fixture tests              |
| Database unavailable/restart                           | safe readiness failure; no data corruption                      | WP-03/WP-07 integration test           |
| API offline/slow response                              | mobile shows an honest error/stale state                        | WP-09 RNTL/Maestro fixture mode        |
| Mapbox token/rate-limit failure                        | accessible list remains usable; no secret logged                | WP-09 device test                      |
| Disk full/backup destination failure                   | job fails visibly without claiming backup success               | WP-13 Compose/runbook rehearsal        |
| Caddy/TLS/container/host failure                       | health checks and rollback/recovery runbook are followed        | WP-13 host rehearsal                   |
| Clock skew/old static feed                             | timestamps/service-day rules remain explicit                    | WP-04/WP-05 unit and integration tests |

Add a deterministic regression test beside the component that owns the failure
mode. This catalog intentionally does not execute host, provider, or device
faults in Phase 0.
