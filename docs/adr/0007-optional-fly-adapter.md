# ADR-0007: Fly.io is an optional image deployment adapter

- Status: Accepted
- Date: 2026-07-22
- Owners: Architecture, Infrastructure
- Evidence required: current pricing, capacity, database, backup, and secret-isolation review before use

## Context

Fly.io can be useful for a trial or staging target but is not the required operating model.

## Decision

Keep Fly.io optional and separate from canonical Compose deployment. If used, deploy the same OCI image and health/migration behavior with no Fly-specific business logic. Do not describe Fly.io as a standing free tier and do not use it for production without current cost and recovery evidence.

## Consequences

Fly configuration is adapter-only. Secrets may be broadly available within a Fly app, so least-privilege requirements can require separate apps. Compose remains the recovery path.

## Rollback / forward fix

Roll back to the previous image or redeploy the same image to the Linux Compose host. Never rely on Fly volumes as the only backup.

## Validation

No Fly account, price, region, database, or cost approval is verified.
