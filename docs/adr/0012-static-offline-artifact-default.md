# ADR-0012: Keep versioned static artifacts as validated JSON pending native SQLite evidence

- Status: Accepted
- Date: 2026-07-23
- Owners: Architecture / Mobile / Data
- Decision gate: D-013

## Context

Phase 2 needs a safe offline representation for static transit data. Expo
SQLite is available behind an adapter, but this workstation cannot run the
physical Android/iOS development builds needed to measure native migration,
indexed query, update, and startup behavior.

The deterministic artifact proxy at
`apps/mobile/scripts/measure-static-artifacts.mjs` generates a representative
synthetic feed and compares compact JSON with a normalized-row representation.
Its results are documented in `apps/mobile/STATIC_ARTIFACT_MEASUREMENT.md`.
They are Node.js serialization/parse proxies, not native SQLite performance.

## Decision

Use a versioned, validated JSON static artifact as the default Phase 2 offline
format. Validate and hash it before atomic replacement. Keep the SQLite local
storage boundary for bounded rider-owned records only; do not claim it stores
or has device-validated static schedules.

Revisit the artifact format only after supported iOS and Android development
builds measure indexed `stop_id` lookup, cold/warm startup, atomic download and
replacement, and native migration behavior. D-013 remains an evidence gate for
that future switch.

## Consequences

JSON keeps fixture mode and degraded behavior simple, inspectable, and
portable. It may be less efficient than indexed SQLite for larger feeds, so it
is deliberately a bounded default rather than a claim of optimal device
performance. Static artifacts contain only transit data; rider location,
search text, credentials, and provider-private payloads stay out of them.

## Rollback / forward fix

Disable or discard a failed JSON artifact by version/hash and retain the prior
validated artifact. A future SQLite migration must support side-by-side
read/validation, atomic activation, and rollback to the prior JSON artifact.

## Validation

Passed locally: `node apps/mobile/scripts/measure-static-artifacts.mjs`, which
reported equivalent serialized-size proxies for the synthetic JSON and
normalized-row forms. Not yet run: physical iOS/Android indexed SQLite,
startup, update, and migration measurements.
