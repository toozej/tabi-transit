---
doc_id: TAB-PLAN-020
title: "Linux Docker Compose Deployment Runbook"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["infra-agent", "release-agent", "sre-agent"]
depends_on: ["TAB-PLAN-010", "TAB-PLAN-011", "TAB-PLAN-014"]
---

# Linux Docker Compose Deployment Runbook

## Outcome

A fresh Linux VPS can run Tabi with one public HTTPS endpoint, private PostGIS, durable feed/database storage, scheduled jobs, encrypted off-site backup, and image rollback.

## 1. Provision server and DNS

- Create a supported 64-bit Linux VPS with at least the measured baseline.
- Add an operator SSH key.
- Create a non-root `tabi-deploy` user.
- Disable password authentication and root SSH.
- Point the API domain A/AAAA record to the server.
- Configure provider reverse DNS only if needed for operational email; Tabi does not send mail by default.
- Record server provider, region, image, disk, IPs, renewal, and console-recovery procedure.

## 2. Harden host

- Apply updates.
- Set UTC system clock and enable time synchronization.
- Install Docker Engine from its official repository and the Compose plugin.
- Configure Docker `local` logging driver or explicit rotation.
- Configure firewall:
  - allow 80 and 443;
  - restrict SSH to VPN/operator addresses where possible;
  - deny all other inbound.
- Do not enable Docker TCP listening.
- Add disk, memory, load, and backup-age monitoring.
- Reboot and confirm Docker starts automatically.

## 3. Create directories and secrets

```bash
sudo install -d -m 0750 -o root -g tabi-deploy /opt/tabi
sudo install -d -m 0700 /etc/tabi/secrets
sudo install -d -m 0750 -o root -g tabi-deploy   /var/lib/tabi/backups   /var/lib/tabi/feed-archive   /var/lib/tabi/static-artifacts
```

Create random secrets with the required length/entropy. Store each as a file under `/etc/tabi/secrets`, never in Git. Give the deploy user only the access needed to invoke the approved Compose commands.

Required:

- `postgres_password`;
- `trimet_app_id`;
- `mapbox_server_token`;
- `installation_auth_key`;
- `restic_password`;
- provider-specific restic environment file.

## 4. Install deployment files

Copy:

- `deployment/compose.yaml`;
- `deployment/compose.production.yaml`;
- `deployment/Caddyfile`;
- `.env`;
- `release.env`;
- scripts;
- systemd units/timers.

Validate:

```bash
cd /opt/tabi
docker compose --env-file .env --env-file release.env   -f compose.yaml -f compose.production.yaml config --quiet
```

## 5. Initialize database

Start PostgreSQL only, wait for health, and run migration:

```bash
docker compose --env-file .env --env-file release.env   -f compose.yaml -f compose.production.yaml up -d postgres

docker compose --env-file .env --env-file release.env   -f compose.yaml -f compose.production.yaml   --profile jobs run --rm migrate
```

Run importer once and verify active feed:

```bash
docker compose --env-file .env --env-file release.env   -f compose.yaml -f compose.production.yaml   --profile jobs run --rm gtfs-importer
```

## 6. Start production

```bash
docker compose --env-file .env --env-file release.env   -f compose.yaml -f compose.production.yaml up -d
```

Verify:

- all required containers healthy;
- `https://API_DOMAIN/health/live`;
- `https://API_DOMAIN/health/ready`;
- known route/stop lookup;
- vehicle source freshness;
- Caddy certificate;
- PostgreSQL not reachable publicly.

## 7. Enable timers

Install and enable:

- `tabi-gtfs-import.timer`;
- `tabi-postgres-backup.timer`;
- `tabi-restic-check.timer`.

Use `systemctl list-timers` and manually run each service once. Timers must be persistent and mutually locked.

## 8. Deploy release

CI uploads `release.env.new` and calls:

```bash
sudo -u tabi-deploy /opt/tabi/deployment/scripts/deploy.sh   /opt/tabi/release.env.new
```

The script validates, locks, dumps DB, pulls, migrates, recreates, checks public health, promotes the release file, and records a deployment log.

## 9. Roll back application

- Set prior image digest in `release.env`.
- Recreate services.
- Verify health.
- Do not restore the database for a normal code rollback.
- If a new feature caused a failure, disable its server flag first.
- If a migration is incompatible, follow its explicit forward-fix plan.

## 10. Restore database

1. Stop API writes and notification worker.
2. Create a separate restore database/container.
3. Restore the custom-format dump.
4. Run integrity queries and application smoke tests.
5. Switch the API only after validation.
6. Re-run realtime pollers to rebuild current snapshots.
7. Document data loss window and incident.

Never overwrite the only database before proving the backup restores.

## 11. Replace a lost host

- Provision/harden a new host.
- Restore deployment files and secrets from secure operator storage.
- Restore the newest verified off-site database dump.
- Restore feed/static artifacts if needed.
- Start Compose and import any missing public data.
- Update DNS.
- Verify TLS, API, source freshness, notification tokens, and scheduled jobs.
- Rotate credentials if host compromise is possible.

## 12. Routine maintenance

Daily:

- source and backup age;
- disk/headroom;
- unhealthy/restarting containers.

Weekly:

- review errors and resource trends;
- remove obsolete local dumps/images without deleting current/previous release.

Monthly:

- host/container security updates;
- restic integrity check;
- restore a sample artifact;
- validate Caddy renewal and domain expiry.

Quarterly:

- full isolated database restore;
- full host-loss tabletop/rebuild;
- capacity and cost review.

## Required evidence

- sanitized bootstrap transcript;
- firewall port scan;
- Compose configuration output;
- backup IDs/checksums;
- restore report;
- deployment and rollback report;
- host resource baseline;
- timer status;
- domain/TLS verification.
