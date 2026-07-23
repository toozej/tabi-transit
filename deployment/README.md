# Linux Docker Compose deployment

This directory implements the accepted single-host deployment shape (ADR-0004): Caddy is the only public service; API, jobs, workers, and PostGIS are private. It is a Phase 0 topology, not evidence of a deployed host or runnable backend image.

## Host preparation

Follow [the Linux Compose runbook](../docs/implementation-plan/20_LINUX_DOCKER_COMPOSE_RUNBOOK.md). Create `/opt/tabi`, `/etc/tabi/secrets`, and the `/var/lib/tabi` bind-mount directories with the documented ownership/modes. Copy `.env.example` to `/opt/tabi/.env`, and have CI install a `0600` `release.env` containing verified immutable image digests. Do not publish PostgreSQL, metrics, admin endpoints, or the Docker socket.

The backend services are still source scaffolds, so this package intentionally contains no Dockerfile and no image-build claim. `TABI_BACKEND_IMAGE` must eventually point to a non-root image, pinned by lowercase SHA-256 digest, with `/app/transit-api`, `/app/realtime-poller`, `/app/notification-worker`, `/app/gtfs-importer`, `/app/tabi-migrate`, and `/app/tabi-healthcheck`. The deployment script rejects tags and malformed digests.

## Validation

`scripts/validate-compose.sh` creates placeholder secret files in a temporary directory and runs Compose interpolation/config validation. It verifies only Caddy publishes host ports, Postgres is confined to the internal backend network, and provider secrets are mounted with least privilege. It neither contacts providers nor starts containers. Run it after Docker Compose and `jq` are installed:

```bash
deployment/scripts/validate-compose.sh
```

Validate Caddy once the pinned Caddy image is available:

```bash
docker run --rm -e TABI_API_DOMAIN=api.example.invalid -v "$PWD/deployment/Caddyfile:/etc/caddy/Caddyfile:ro" caddy:VERSION caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
```

For production, use `docker compose --env-file .env --env-file release.env -f compose.yaml -f compose.production.yaml`. Enable only the necessary profiles: `jobs`, `notifications`, and (after an ADR) any observability profile.

## Jobs and recovery

Install the systemd service/timer pairs from `systemd/`, then manually run every service once and inspect `systemctl list-timers`. `backup-postgres.sh` creates logical custom-format dumps and checksums before optional restic backup. Restore rehearsals must target an isolated database first; no restore script is provided that overwrites the live database.

`deploy.sh` validates an immutable candidate release file, backs up, runs migrations, starts services, checks public readiness, and restores the previous application image set on a failed health check. Database migrations must remain expand/migrate/contract compatible for image rollback.

The scripts use `/run/lock` by default. Test harnesses may set `TABI_LOCK_DIR` to a private writable temporary directory; production must retain the host lock directory.

`restore-postgres-isolated.sh DUMP` first verifies the custom-format archive, then restores it only into a separate Compose project and performs a basic SQL connection check. It deliberately never touches the live project; operators must run documented integrity and application smoke checks before any switch.

## Optional Fly adapter

`fly/fly.toml.example` is deliberately separate and not a production plan. It makes no free-tier claim. Use it only after pricing, region, PostGIS, backup, and secret-isolation evidence satisfies ADR-0007; the same OCI image and Compose recovery path remain required.
