# First vehicle vertical-slice validation

This runbook verifies the fixture-backed path from the executable API contract
to the accessible mobile list/detail representation. It intentionally does not
call a transit provider, require a credential, or assert native Mapbox output.

## Automated local evidence

Run:

```sh
make generate-check
make test-integration
make test-vehicle-payload-benchmark
```

`tests/integration/run.sh` checks the representative API fixtures, OpenAPI
structure and generated-client drift, black-box public API unit boundary,
mobile's pure Vitest boundary, and Compose topology. A generated-code mismatch
is a release-blocking contract failure: regenerate through the owned API
workflow; do not hand-edit generated code.

The payload script reports build/stringify time and payload bytes for 1x,
1.5x, and 3x synthetic fleets. It proves data-shape construction only; it does
not establish device frame rate, memory, or native `ShapeSource` performance.

## Device/Mapbox gate (not passed by these commands)

With an approved restricted public Maps SDK token and an installed Expo
development build, run:

```sh
maestro test tests/e2e/maestro/vehicle-fixture-flow.yaml
```

The required assertion is that fixture mode opens the vehicle experience,
filters/searches an exact vehicle ID, selects it from the accessible list, and
shows normalized source and freshness. Then repeat on a physical Android and
iOS device with Mapbox configured, measuring source update, selection/filter
latency, frame behavior, and memory at 1x, 1.5x, and 3x fleets.

Do not call this a passing device, native-map, accessibility, or production
deployment test until it actually runs. The current Vitest/RNTL evidence gate
remains governed by ADR-0009.

## Failure expectations

- A 503 `source_unavailable` response must leave the prior valid snapshot
  intact; it must never be labelled live.
- A stale entity must present its freshness status and age, not animation or a
  live claim.
- PostgreSQL stays private in Compose; Caddy is the sole public-port service.
- Never place coordinates, provider credentials, or tokens in test output.
