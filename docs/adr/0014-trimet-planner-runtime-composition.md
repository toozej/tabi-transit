# ADR-0014: TriMet planner runtime composition

- Status: Accepted
- Date: 2026-07-28
- Owners: Architecture, Backend, Product
- Evidence required: normalized mapper, API contract, fixture/live-safe integration tests, and mobile repository proof

## Context

D-001 is resolved and the TriMet planner adapter has sanitized fixture tests,
but the public journey endpoint still fails closed. The TriMet client uses
provider-specific request/response types and cannot be assigned directly to
the application planner gateway.

## Decision

Create a dedicated mapper from application journey requests to the private
TriMet planner client and from provider results to normalized itineraries. At
startup, compose the mapper only from validated server-side TriMet
configuration. Replace the fail-closed journey handler with validated request
parsing and the application planning gateway. The mobile planner must use the
normal API repository while retaining deterministic fixture mode for tests.

## Consequences

The API must preserve provider source/freshness, disclose unsupported
constraints, and never expose AppIDs or provider DTOs. Unavailable or malformed
upstream results must remain safe errors rather than fabricated itineraries.

## Rollback / forward fix

Disable the composed planner through configuration while preserving the
provider-neutral contract and fixture mode. Extend only through mapper tests
when TriMet adds supported fields.

## Validation

Add mapper table tests, startup configuration tests, API request/response/error
tests, OpenAPI/client drift checks, and mobile remote/fixture repository tests.
