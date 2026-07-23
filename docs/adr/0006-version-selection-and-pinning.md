# ADR-0006: Evidence-based dependency selection and pinning

- Status: Accepted
- Date: 2026-07-22
- Owners: Architecture, Repository
- Evidence required: lockfiles, tool directives, compatibility reports, and container image digests

## Context

The selected stack spans JavaScript/native/mobile, Go, database, and deployment tooling. Floating production versions create reproducibility and compatibility risk.

## Decision

Select current compatible stable versions during Phase 0 using the criteria in the dependency-version matrix. Pin Node, pnpm, Go, Expo/RN, RNMapbox/native SDK, PostgreSQL/PostGIS, Docker/Compose, Caddy, generators, and production images in their appropriate tool files, lockfiles, or digests. Do not claim a version is selected until it is recorded and validated.

## Consequences

WP-01 and WP-08 must provide evidence before the matrix can change from proposed to selected. Updates need compatibility tests and a rollback path.

## Rollback / forward fix

Revert lockfile/tool directive or image digest where compatible; use additive migrations and documented native rebuild requirements.

## Validation

`make doctor`, lockfile checks, mobile compatibility evidence, and Compose image-digest validation are pending implementation.
