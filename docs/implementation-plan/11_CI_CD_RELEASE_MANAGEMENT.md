---
doc_id: TAB-PLAN-011
title: "CI/CD and Release Management"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["cicd-agent", "mobile-release-agent", "backend-release-agent"]
depends_on: ["TAB-PLAN-003", "TAB-PLAN-010", "TAB-PLAN-012"]
---

# CI/CD and Release Management

## Pull request workflow

Parallel jobs:

1. repository policy and changed-files;
2. format/lint/typecheck;
3. Vitest and coverage;
4. React Native component tests;
5. Go unit/race/fuzz smoke;
6. OpenAPI lint and generated drift;
7. PostGIS migrations/integration;
8. upstream fixture contracts;
9. `docker compose config` for base/production/profile combinations;
10. Dockerfile/Compose/security lint;
11. container build without push;
12. dependency/license/secret scan;
13. Expo iOS/Android preview builds when native/config changes;
14. Maestro smoke for release-critical/native changes.

Path filtering must never skip shared contracts, database migrations, deployment files, or generated-code validation.

## Container images

Build one backend runtime image containing separate Go entrypoints for API, poller, importer, migration, worker, and healthcheck. Compose selects the command per service.

Requirements:

- multi-stage build;
- minimal non-root runtime;
- read-only-compatible filesystem;
- pinned base image digest;
- OCI labels with source/commit/version;
- SBOM and provenance;
- vulnerability scan;
- immutable GHCR digest;
- current and previous image retained on the host.

PostGIS and Caddy images are also pinned by digest in `release.env`.

## Main workflow

Revalidate → build/sign backend image → generate SBOM/provenance → push GHCR → deploy development/preview where configured → run API/database/source smoke → publish image digest and commit metadata → trigger internal EAS builds as needed.

## Linux server deployment workflow

Use a protected GitHub environment with:

- server hostname;
- unprivileged deploy username;
- SSH private key;
- known-host fingerprint;
- optional WireGuard/Tailscale connection material;
- no root password and no Docker registry password stored on the server when a scoped pull token is sufficient.

Deployment:

1. create `release.env.new` with immutable digests;
2. copy it and the deployment script to `/opt/tabi`;
3. execute `deployment/scripts/deploy.sh`;
4. script takes a local database dump;
5. validate Compose configuration;
6. authenticate to GHCR using a read-only token;
7. pull;
8. run migrations through the `jobs` profile;
9. recreate changed services;
10. wait for health;
11. perform public smoke tests;
12. atomically promote the release file;
13. retain previous release metadata.

Use GitHub environment approvals for production. Limit SSH `authorized_keys` with a dedicated user and least privilege. Do not expose the Docker TCP API.

## Staging release

- Deploy a separate Compose project/host.
- Use an independent PostGIS volume/database and provider credentials.
- Build EAS staging iOS/Android.
- Run full Maestro.
- Run load, migration, backup, source-outage, and degraded-state tests.
- Produce a release report with server capacity and disk impact.

## Production mobile

Automate version/build numbers, release notes, privacy/credits/licenses, EAS signed AAB/IPA, Play internal/closed and TestFlight submissions, validation, protected promotion, and GitHub release with backend compatibility.

## Scheduled workflows

- low-rate upstream contract smoke;
- dependency and vulnerability review;
- GTFS/source health;
- backup-age verification;
- remote restic repository check;
- certificate/domain expiration checks;
- stale docs/links;
- Fly.io trial configuration validation when that optional target is maintained.

The production GTFS import and database dump schedules run on host systemd timers, not GitHub Actions, so they continue during GitHub outages.

## EAS profiles

`development`, `simulator`, `preview`, and `production`, with environment-specific API and public Mapbox settings. Native SDK download credentials stay in EAS secrets.

## Versioning

Mobile semantic version plus iOS build/Android code. Backend immutable commit digest and release tag. `/v1/config` advertises API build and minimum supported app. Database and static-feed versions remain independent.

## Expo Updates

Only compatible JavaScript/assets. Runtime-version gating, staged QA, rollback rehearsal, and no native/config/privacy bypass. Native changes require a binary.

## Database ordering

Expand → deploy code supporting old/new → backfill → switch reads → later contract. Feed activation remains separate. Deployment script never automatically runs an irreversible destructive migration without an explicit approved flag and fresh verified backup.

## Release channels

Development, preview/staging, beta/TestFlight/closed testing, phased public, GA. Feature flags disable optional provider and notification functions.

## Quality gates

Backend:

- required CI green;
- Compose validation;
- staging soak;
- reviewed migration;
- verified backup age;
- no P0/P1;
- scan policy;
- source health;
- enough host disk/memory headroom.

Mobile:

- physical devices;
- critical Maestro;
- accessibility;
- privacy/credits/Mapbox controls;
- backend compatibility.

## Rollback

Backend:

- restore prior `release.env`;
- pull/recreate prior images;
- keep additive database schema compatible;
- disable source/feature flags;
- reactivate prior GTFS feed;
- use database restore only for data corruption, not ordinary application rollback.

Mobile:

- halt phased release;
- corrected binary;
- compatible OTA rollback;
- server flag disable.

## Release record

Commit, image digest, SBOM/signature/provenance, OpenAPI hash, migrations, Compose configuration hash, host/deployment target, EAS/store versions, tests, feed versions, backup ID/age, notes/issues/rollback.

## Acceptance criteria

- PR validation is deterministic.
- Images are immutable and traceable.
- Deployment uses restricted SSH and protected environments.
- Production has no resident CI runner by default.
- Compose and migration failures stop safely.
- Image rollback and host rebuild are rehearsed.
- EAS works without local signing.
- Contract drift blocks merge.
