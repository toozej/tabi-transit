# ADR-0004: Canonical single-host Docker Compose deployment

- Status: Accepted
- Date: 2026-07-22
- Owners: Architecture, Infrastructure, Security
- Evidence required: Compose validation, deployment/rollback rehearsal, and restore report before production

## Context

The reference deployment must be operable at early-stage scale without cloud-specific orchestration.

## Decision

Use one Linux host with Docker Engine and Docker Compose. Caddy is the only public reverse proxy/TLS endpoint. PostgreSQL with PostGIS, metrics, admin services, jobs, and workers remain private; PostgreSQL has no host port. Use immutable image digests, Compose secrets with `_FILE` configuration, systemd one-shot timers, logical dumps, and encrypted off-site backup.

## Consequences

Single-host loss is an accepted initial availability risk mitigated by tested restore and host-replacement runbooks. No Terraform, Kubernetes, or required cloud platform is introduced.

## Rollback / forward fix

Roll application images back by prior digest; use expand/migrate/contract database changes so image rollback remains safe. Restore first into an isolated database, never over the only live database.

## Validation

Production acceptance requires the runbook evidence listed in plan 20; no host, domain, or backup target is currently verified.
