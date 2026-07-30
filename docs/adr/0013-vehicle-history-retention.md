# ADR-0013: Thirty-day normalized vehicle-history retention

- Status: Accepted
- Date: 2026-07-28
- Owners: Product, Security, Architecture
- Evidence required: migration and retention-boundary tests

## Context

Specialist status, adherence, and history views need bounded vehicle
observations. Keeping only current state prevents those features; retaining raw
provider payloads or unbounded location history creates unnecessary privacy,
cost, and operational risk.

## Decision

Retain normalized vehicle observations for exactly 30 days. Each accepted
realtime snapshot records its normalized vehicle projection in
`history.vehicle_observations`; raw upstream payloads are never stored. The
same transaction prunes observations with `processed_at` older than 30 days.
ADR-0015 defines the separately bounded public history API.

## Consequences

The retention/data-rights decision for the Phase 5 history foundation is
cleared. History/adherence presentation remains subject to a later API/UI
contract and applicable TriMet source controls. The bounded table needs
storage monitoring and cannot recover observations older than 30 days.

## Rollback / forward fix

Disable the history insert path in a follow-up migration or delete the history
table after stopping dependent readers. Changes to the duration require a new
ADR and matching migration/test update.

## Validation

`make test` proves the 30-day policy constant. `make test-db-migrations`
applies migrations twice and proves that a row older than 30 days is removed
while the boundary row is retained.
