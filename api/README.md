# API contract

`openapi.yaml` is the executable source of truth for the public `/v1` API. It
contains only normalized Tabi models, never provider payloads or credentials.
`codegen.json` declares the generated TypeScript output. Go generation remains
deliberately deferred until its generator and compatibility policy are pinned
by the architecture owner.

`examples/` contains deterministic, sanitized JSON payloads that mirror the
embedded OpenAPI examples and can be used by contract and mobile fixture tests.

## Commands

Run the dependency-free structural check:

```sh
ruby api/scripts/validate-openapi.rb api/openapi.yaml
```

Run the full lint once the repository-pinned Redocly CLI is installed:

```sh
api/scripts/lint.sh
```

Generate the TypeScript client (owned by `packages/api-client` after WP-01
creates it) and check it for drift:

```sh
api/scripts/generate-typescript.sh
api/scripts/generate-check.sh
```

The generation scripts intentionally fail when their generator is unavailable;
they never report a synthetic success. Root `make generate` and CI should call
these scripts after pinning `@redocly/cli` and `openapi-typescript`.

## Generated-code ownership

`packages/api-client/src/generated/` is generated from this document. Do not
hand edit generated files. Change `openapi.yaml`, run generation, and commit
the resulting deterministic output in the same change. Backend and mobile code
consume generated types or deliberate adapters; neither is allowed to redefine
the wire contract.

## Contract rules

- IDs are source-qualified; API clients must treat them as opaque.
- Timestamps are ISO 8601 UTC instants; service dates are separate calendar
  values.
- Coordinate arrays are `[longitude, latitude]`; distances are meters.
- Clients tolerate additive fields and `unknown` enum values.
- Realtime data must include `freshness`; a stale status is never live data.
- ETags are supplied for snapshot/static endpoints and `304` has no body.
