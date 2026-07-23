# ADR-0002: Expo development builds and isolated mobile adapters

- Status: Accepted
- Date: 2026-07-22
- Owners: Architecture, Mobile
- Evidence required: Phase 0 compatibility and physical-device reports

## Context

Tabi requires native Mapbox and SQLite capabilities while keeping mobile feature code testable and accessible.

## Decision

Use React Native with Expo development builds, TypeScript strict mode, Expo Router, TanStack Query for server state, Zustand for ephemeral UI state, Zod at boundaries, and Expo SQLite for local persistence. Isolate native/platform APIs and RNMapbox behind adapters. A map flow must provide an equivalent accessible list/detail flow.

## Consequences

Expo Go is not a supported development target. Feature code cannot call provider APIs directly. RNMapbox compatibility and the Vitest/RNTL harness remain separately gated.

## Rollback / forward fix

Replace a native adapter without changing feature/domain APIs. If the test harness cannot prove native assertions, keep pure logic in Vitest and move device interactions to Maestro after a documented ADR update.

## Validation

WP-08 must report build-time checks, SQLite proof, map `ShapeSource` spike, and any unrun device gate.
