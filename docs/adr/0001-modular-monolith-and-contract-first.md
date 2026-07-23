# ADR-0001: Contract-first modular monolith

- Status: Accepted
- Date: 2026-07-22
- Owners: Architecture
- Evidence required: OpenAPI lint/generation and module-boundary tests as packages are added

## Context

The MVP needs independently runnable API, importer, poller, and worker processes without premature service decomposition. Public clients need a stable, testable boundary.

## Decision

Use a Go modular monolith with separate binaries for `transit-api`, `gtfs-importer`, `realtime-poller`, and later `notification-worker`. Define public HTTP behavior OpenAPI-first under `/v1`; the executable OpenAPI document becomes the client/server contract once introduced.

## Consequences

Shared domain, configuration, source, and persistence packages reduce duplication. Long-running imports stay out of HTTP requests. Contract changes require compatibility review; generated code is not hand edited.

## Rollback / forward fix

Components can be extracted only after measured ownership, scaling, or reliability evidence. Preserve API compatibility during any extraction.

## Validation

WP-02 supplies a linted initial specification and deterministic generation; later packages supply conformance tests.
