# ADR-0009: Keep Native Component Assertions Behind the Vitest/RNTL Evidence Gate

- Status: Accepted
- Date: 2026-07-23
- Owners: Mobile / Quality
- Decision gate: D-011

## Context

The selected testing strategy requires Vitest for framework-independent TypeScript and a React Native Testing Library harness proven compatible with Vitest. The installed dependency pair reaches an untransformed React Native Flow entrypoint before component tests can run. This is not evidence that native components or Mapbox work.

## Decision

Keep repository, filtering, freshness, validation, and GeoJSON tests in the passing pure-Vitest suite. Do not add a duplicate Jest suite merely to bypass this incompatibility. Keep React Native component tests out of the passing claim until a minimal RNTL/Vitest proof runs with the selected dependency matrix. Validate native Mapbox behavior, accessibility interaction, SQLite, and navigation through Maestro on physical development builds once the required device and restricted public token are available.

## Consequences

The mobile vertical slice is fixture-tested at its domain and API boundary, but its RNTL and physical-device evidence gates remain open. Any package upgrade or harness configuration that resolves the Flow transform must add a representative component test and update D-011 before it is considered accepted evidence.
