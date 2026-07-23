# ADR-0003: Backend normalization with explicit freshness

- Status: Accepted
- Date: 2026-07-22
- Owners: Architecture, Backend
- Evidence required: contract and source-fixture tests

## Context

Provider payloads vary and realtime data can be delayed, malformed, or absent. Presenting stale information as live harms riders.

## Decision

Normalize approved provider data in the backend. Every realtime entity and response carries source/fetch/entity timestamps and a derived freshness state. A failed, invalid, or unexpectedly empty poll never replaces the last valid snapshot. Mobile displays freshness and does not animate stale data as current.

## Consequences

Source-specific schemas stay private. Pollers need source-health state and safe snapshot replacement. Static GTFS remains usable during realtime failures.

## Rollback / forward fix

Freshness thresholds are configuration, not API-breaking constants. Add sources via adapters and contract additions; disable a source by feature flag if evidence fails.

## Validation

Fixture tests must cover valid, invalid, empty, stale, and recovery cases before a provider is enabled.
