# ADR-0016: Historical adherence from retained trip-update evidence

- Status: Accepted
- Date: 2026-07-28
- Owners: Product, Transit Domain, Backend, Security
- Evidence required: retained delay model, schedule matching tests, classification policy, and rider-language review

## Context

Vehicle-position history cannot truthfully establish whether service was early,
on time, or late. Current trip updates are current-state projections and are
overwritten, so their delay/timing evidence is not retained historically.

## Decision

Before exposing adherence, retain normalized historical trip-update observations
for the same 30-day policy. Preserve schedule-deviation/delay evidence and the
minimum source timestamps needed to match it to a scheduled trip. Classify only
when the evidence is sufficient; otherwise report unknown. Define thresholds
and rider-facing wording in the contract rather than deriving claims from GPS.

## Consequences

This adds retention volume and schedule-matching complexity. It prevents false
precision and requires explicit handling for canceled, replaced, stale,
unmatched, and after-midnight service.

## Rollback / forward fix

Keep adherence unavailable while retaining the existing vehicle-history feature.
Disable historical trip-update writing independently if the policy changes.

## Validation

Add migration/writer retention-boundary tests; schedule matching fixtures for
early, on-time, late, unknown, canceled, and after-midnight cases; API/mobile
disclaimer tests; and query/storage performance evidence.
