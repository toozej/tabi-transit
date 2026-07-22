---
doc_id: TAB-PLAN-010
title: "Linux Server and Docker Compose Infrastructure"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["infra-agent", "security-agent", "sre-agent"]
depends_on: ["TAB-PLAN-002", "TAB-PLAN-006", "TAB-PLAN-008"]
---

# Linux Server and Docker Compose Infrastructure

## Canonical production target

Deploy Tabi to one ordinary Linux server or VPS using:

- Docker Engine and the Docker Compose plugin;
- immutable images from GHCR;
- Caddy as the only internet-facing reverse proxy and TLS terminator;
- PostgreSQL with PostGIS in a private Compose network;
- separate API, realtime poller, notification worker, migration, and importer processes;
- systemd timers for one-shot Compose jobs;
- encrypted restic backups to a second provider or machine;
- optional observability services through a disabled-by-default Compose profile.

No cloud-specific infrastructure-as-code platform is required. The server must be replaceable from documented host bootstrap steps, Compose files, secrets, image digests, and backups.

## Capacity baseline

Start production with a measured baseline of:

- 2 virtual CPU cores;
- 4 GB RAM;
- 40 GB or more SSD;
- a supported 64-bit Linux distribution;
- stable IPv4/IPv6 and a domain name.

A 1 vCPU/2 GB host can be used for development or a very small pilot, but PostgreSQL, map API traffic, imports, and observability must be profiled before relying on it. Add swap only as an emergency buffer, not as normal working memory.

Capacity gates:

- database working set fits memory;
- GTFS import completes without OOM;
- all-system vehicle response meets latency budget;
- backup does not materially interrupt API;
- disk has at least 30% headroom after a representative retention period.

## Host layout

```text
/opt/tabi/
  compose.yaml
  compose.production.yaml
  Caddyfile
  .env
  release.env
  deployment/scripts/
  deployment/systemd/

/etc/tabi/secrets/
  postgres_password
  trimet_app_id
  mapbox_server_token
  installation_auth_key
  restic_password
  restic_environment

/var/lib/tabi/
  backups/
  feed-archive/
  static-artifacts/
```

Secret files are root-owned mode `0400` or `0600`. Application data normally uses Docker named volumes; backup/export directories use explicit bind mounts so operators can inspect and copy them safely.

## Compose services

### Always on

- `caddy`: publishes host ports 80 and 443; proxies only to `api`.
- `api`: stateless Go HTTP service; private database connection.
- `realtime-poller`: one active process with a database advisory lock.
- `postgres`: PostGIS image pinned by digest; no published host port.

### Profiles

- `notifications`: `notification-worker`.
- `observability`: metrics/log/dashboard containers selected by ADR.
- `jobs`: `migrate` and `gtfs-importer` one-shot services.
- `otp`: OpenTripPlanner only after its decision gate.

The production host uses:

```bash
docker compose   --env-file .env   --env-file release.env   -f compose.yaml   -f compose.production.yaml   up -d
```

## Networks and exposure

- `frontend`: Caddy and API only.
- `backend`: internal database network for PostgreSQL and backend processes.
- `egress`: outbound-only application network for API, poller, workers, and jobs to reach approved providers.
- PostgreSQL is attached only to `backend` and has no `ports` mapping.
- API has `expose`, not a public host port.
- Caddy is the only public container.
- Host firewall allows:
  - 80/tcp and 443/tcp from the internet;
  - SSH only from approved addresses or a VPN;
  - no database, metrics, or Docker daemon exposure.
- Disable password SSH login and interactive root login.
- Never expose the Docker socket to an application container.

## TLS and reverse proxy

Caddy receives the production hostname and automatically manages public HTTPS certificates. The Caddyfile:

- proxies to `api:8080`;
- forwards request IDs and standard proxy headers;
- enables gzip/zstd;
- applies safe response headers;
- writes structured logs to stdout;
- preserves `/health/live` and `/health/ready`;
- does not expose private metrics.

DNS A/AAAA records point directly to the host. Certificate state is persisted in `caddy_data` and included in backup, although certificates can be reissued.

## Persistent storage

Named volumes:

- `postgres_data`;
- `caddy_data`;
- `caddy_config`.

Bind mounts:

- `/var/lib/tabi/feed-archive`;
- `/var/lib/tabi/static-artifacts`;
- `/var/lib/tabi/backups`.

The API and poller containers are read-only except for explicitly mounted temporary directories. No important application state lives in an unnamed container filesystem.

## Secrets and configuration

- Non-secret settings live in root-owned `.env`.
- Immutable image references live in `release.env`.
- Secrets are mounted through Compose `secrets`.
- Go configuration supports `_FILE` variables such as:
  - `DATABASE_PASSWORD_FILE`;
  - `TRIMET_APP_ID_FILE`;
  - `MAPBOX_SERVER_TOKEN_FILE`;
  - `INSTALLATION_AUTH_KEY_FILE`.
- Do not put secrets in image layers, Git, shell history, Compose output, or health endpoints.
- Rotate deploy and provider credentials through a documented runbook.

## Scheduling

Use host systemd timers to invoke idempotent one-shot jobs:

- GTFS import: daily and manual;
- database dump: every six hours;
- encrypted off-site backup: after successful dump and daily;
- restic integrity check: monthly;
- retention/pruning: daily;
- restore rehearsal: quarterly in an isolated Compose project.

Timers use `flock` or equivalent to prevent overlap and `Persistent=true` so missed runs execute after reboot.

## Backup design

### Local dump

- `pg_dump --format=custom` to `/var/lib/tabi/backups/postgres/`.
- Write to a temporary filename, fsync where practical, then rename.
- Keep a short local window for fast restore.
- Hash dumps and log duration/size.

### Off-site backup

Use restic with a repository hosted by a different failure domain, for example an S3-compatible bucket, Backblaze B2, an SFTP server, or a second machine.

Back up:

- database dumps;
- feed archive subject to source terms;
- static artifacts;
- Compose/Caddy configuration;
- encrypted copies of non-recreatable secret material according to the key-management runbook.

Do not back up live PostgreSQL data-directory files as a substitute for database-aware dumps.

### Restore

Restore into a new database/container first, run integrity and application smoke checks, then switch the API. The runbook must support both full-host loss and accidental data corruption.

Initial targets:

- RPO: six hours for installation/subscription data;
- RTO: four hours for a practiced host replacement;
- public GTFS/realtime cache can be reconstructed from upstream and archived static feeds.

## Logging and observability

Use Docker's `local` logging driver or explicit rotated `json-file` settings so logs cannot fill the disk. Backend logs are structured JSON.

Lean production baseline:

- `/health` checks;
- `/metrics` bound to the backend network;
- external HTTPS uptime check;
- disk, memory, load, certificate, source freshness, backup age, and container-restart alerts.

The optional `observability` profile may run Prometheus-compatible storage, Grafana, and a log collector only after memory/disk impact is measured. A small host may instead export metrics/logs to a low-cost external service.

## Deployment and rollback

1. CI tests and builds one immutable backend image containing all Go binaries.
2. CI pushes the image and SBOM to GHCR.
3. A protected deploy job connects through a dedicated unprivileged SSH account.
4. Upload a new `release.env` containing immutable image digests.
5. Acquire a deployment lock.
6. Run a fresh database dump.
7. Pull images.
8. Run the one-shot migration service.
9. Recreate changed services with Compose.
10. Check container health and public API smoke tests.
11. On failure, restore the previous `release.env` and recreate prior images.
12. Database changes use expand/migrate/contract so image rollback remains safe.

Do not install a GitHub Actions runner on the production host by default. Use a GitHub-hosted runner over restricted SSH. If a self-hosted runner is ever used, isolate it from production credentials and Docker control.

## Environments

### Local and CI

Use the same base Compose file with mock upstreams and an ephemeral PostGIS volume.

### Preview/staging

Low-cost options, in preference order:

1. a second small VPS;
2. a separate Compose project on the same host with distinct networks, volumes, domains, credentials, and resource limits;
3. a temporary Fly.io deployment.

Never point staging jobs or mobile preview builds at the production database.

### Production

One Compose project and one database. Single-host failure is an accepted early-stage tradeoff mitigated by off-site backups, documented rebuild, status communication, and optional later migration to two hosts or a managed database.

## Host maintenance

- Apply Linux and Docker security updates on a defined schedule.
- Reboot after kernel updates during a maintenance window.
- Monitor disk and inode use.
- Prune unused images only after retaining the current and previous releases.
- Test Docker daemon restart and host reboot.
- Pin image digests and verify supply-chain signatures where available.
- Run `docker compose config` before every deployment.
- Keep the Docker API on its local Unix socket only.

## Acceptance criteria

- A clean Linux server can be bootstrapped from the runbook.
- `docker compose config` validates base and production files.
- Only 80/443 and restricted SSH are reachable.
- PostgreSQL is unreachable from the public network.
- A reboot restores all always-on services and persistent data.
- Scheduled imports and backups execute after missed timers.
- Off-site backup and isolated restore pass.
- Deployment and image rollback are rehearsed.
- Disk-filling logs are prevented.
- Production runs without any cloud-specific orchestration service.
