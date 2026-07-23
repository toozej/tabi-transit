# ADR-0008: Defer optional platform and product expansions behind evidence gates

- Status: Accepted
- Date: 2026-07-22
- Owners: Architecture, Product, Security
- Evidence required: individual ADR amendments and phase-gate evidence

## Context

Several technologies/features can increase cost, privacy risk, or operational scope before the minimum vertical slice proves value.

## Decision

Do not deploy OpenTripPlanner, Redis, historical vehicle storage, notifications, specialist views, optional sources, broad observability profile, or Fly.io as prerequisites for the first vertical slice. Use Expo Push behind an abstraction only after notification scope is approved. Choose vehicle hot-path JSON versus GeoJSON, SQLite artifact format, exact RNMapbox matrix, and Vitest/RNTL support from measured Phase 0 evidence.

## Consequences

The first slice remains focused on normalized vehicle data, public API, accessible mobile map/list, and canonical Compose. Deferred features must not create hidden runtime dependencies.

## Rollback / forward fix

Enable later features through independently tested flags, migrations, contracts, and ADR evidence. Remove a feature by disabling it without affecting core transit reads.

## Validation

See the architecture decision register for open owners and decision gates.
