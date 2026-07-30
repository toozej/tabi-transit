# ADR-0005: Official-source-first and optional-source gates

- Status: Accepted
- Date: 2026-07-22
- Owners: Architecture, Product, Security
- Evidence required: provider terms, credentials, licensing/attribution review, and fixture tests per source

## Context

Transit data, Mapbox services, and optional sources have operational, contractual, attribution, and privacy constraints.

## Decision

Use TriMet official GTFS, GTFS-Realtime, and Web Services through backend adapters only. Keep TriMet privileged credentials server-side. Treat Portland Streetcar schedule/realtime/arrival data included in those official TriMet sources as TriMet-provided `streetcar` coverage, not a separate provider integration; do not call or scrape Portland Streetcar, PBOT, or UmoIQ directly. Treat Rose City Transit as a TriMet-derived presentation/reference only: do not call, scrape, store data from, or imply a partnership with Rose City. Do not scrape public web pages. Do not persist Mapbox geocoder data until the selected product and storage rules are approved.

## Consequences

Missing credentials do not block unrelated work: adapters use typed configuration, sanitized fixtures, and feature flags. No partner relationship is implied.

## Rollback / forward fix

Disable any adapter by configuration and retain the last valid data under its freshness rules. Add an approved source through a new or amended ADR and contract review. TriMet Streetcar coverage stays within the approved TriMet adapter boundary and is enabled only by validated runtime configuration.

## Validation

The external-source register tracks all evidence gates; no external account or terms status is claimed as verified.
