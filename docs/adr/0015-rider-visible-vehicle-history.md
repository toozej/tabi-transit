# ADR-0015: Rider-visible vehicle history contract

- Status: Accepted
- Date: 2026-07-28
- Owners: Product, Architecture, Mobile, Backend
- Evidence required: bounded API contract, persistence/query tests, accessible mobile proof, and performance measurements

## Context

ADR-0013 retains normalized vehicle observations for 30 days, but there is no
reader, API, generated client, or rider-visible history view.

## Decision

Provide a vehicle-history endpoint backed only by
`history.vehicle_observations`. It must validate opaque vehicle IDs, cap the
requested interval at 30 days, paginate or downsample observations, return
source/freshness metadata, and distinguish empty, stale, and unavailable
states. The mobile vehicle detail screen will show an accessible textual
timeline first; any map rendering is supplemental.

## Consequences

No raw provider payload, observation older than 30 days, or inferred adherence
may be exposed. Payload/query limits and storage/query measurements are
required before release.

## Rollback / forward fix

Feature-flag the endpoint and mobile panel. Existing retention pruning remains
in force even while the feature is disabled.

## Validation

Add persistence keyset-pagination/window tests, API validation/ETag/error
tests, OpenAPI/generated-client drift checks, mobile timeline loading/empty/
error/accessibility tests, and representative payload/performance evidence.
