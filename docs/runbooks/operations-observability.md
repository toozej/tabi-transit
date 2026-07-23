# Operations and observability

This is the non-host operational baseline for the single-host Compose design.
It is not evidence that a host, domain, dashboard, backup repository, or alert
receiver has been configured.

## Signals and safe fields

Use JSON logs with event name, level, request/job/run ID, HTTP route/status,
duration, source ID, feed/snapshot version, freshness status/age, and classified
error code. Never log authorization headers, tokens, passwords, AppIDs, raw
request bodies, free-text searches, or precise coordinates. The repository
policy check rejects recognizable credential and location logging patterns.

The implemented API readiness endpoint is the primary local operational probe:

```bash
curl --fail --silent --show-error https://API_DOMAIN/health/live
curl --fail --silent --show-error https://API_DOMAIN/health/ready
```

Readiness requires a reachable database and a successful source-health record;
it is not merely a process-liveness signal. Inspect normalized response
freshness (`source`, `fetchedAt`, `processedAt`, `status`, and `ageSeconds`)
when investigating a realtime incident. A stale or unavailable upstream must
be communicated as such, never represented as live.

## Local validation and failure-safe checks

From a host copy of the deployment directory:

```bash
deployment/scripts/ops-preflight.sh runtime
deployment/scripts/ops-preflight.sh backup
deployment/scripts/verify-backup-age.sh 6
deployment/scripts/verify-restore-candidate.sh /var/lib/tabi/backups/postgres/tabi-....dump
```

The scripts fail when deployment files, local dumps/checksums, restore tooling,
or required secret files are absent. They do not print secret values, initiate a
backup, contact restic, or restore a database. `verify-backup-age.sh` reports
off-site status as unverified unless restic secret files are present; then run
`restic-check.sh` for repository verification. A real restore still uses
`restore-postgres-isolated.sh` with a distinct Compose project and integrity/API
smoke checks before any switch.

## Alert/runbook catalog

| Signal | Severity | First action |
| --- | --- | --- |
| `/health/ready` fails | Page | Check database, source-health, recent deploy; roll back only application image if needed. |
| Source freshness exceeds its threshold | Ticket/Page by rider impact | Disable the source if invalid; preserve the last valid snapshot and show its freshness. |
| Backup age/checksum fails | Page | Stop claiming recoverability; inspect dump job/disk, then verify an isolated restore. |
| Disk, memory, or container restart trend | Ticket/Page at exhaustion risk | Preserve current/previous images and dumps; free space only under the host maintenance procedure. |
| Caddy/TLS failure | Page | Validate DNS/certificate/container logs; never expose API, metrics, Postgres, or Docker as a workaround. |
| Credential/billing suspicion | Page | Disable affected adapter, rotate through operator storage, and review safe logs. |

## Deferred observability profile

No `observability` Compose profile is enabled yet. The backend does not expose a
private metrics endpoint, and adding collectors/dashboards without measured
memory/disk budget and an approved observability ADR would create an unsupported
deployment claim. When this changes, metrics must remain on the internal backend
network, Caddy must not proxy them, and the profile must be validated alongside
host firewall and retention evidence.

## Incident record

Record detection time, user impact, source/release identifiers, freshness,
actions, recovery evidence, and a regression test/runbook improvement. Do not
attach secrets, raw coordinates, full queries, provider payloads, or user data.
