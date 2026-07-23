# Linux Docker Compose deployment

This directory implements the accepted single-host deployment shape (ADR-0004): Caddy is the only public service; API, jobs, workers, and PostGIS are private. It is a Phase 0 topology, not evidence of a deployed host or runnable backend image.

## Host preparation

Follow [the Linux Compose runbook](../docs/implementation-plan/20_LINUX_DOCKER_COMPOSE_RUNBOOK.md). Create `/opt/tabi`, `/etc/tabi/secrets`, and the `/var/lib/tabi` bind-mount directories with the documented ownership/modes. Copy `.env.example` to `/opt/tabi/.env`, and have CI install a `0600` `release.env` containing verified immutable image digests. Do not publish PostgreSQL, metrics, admin endpoints, or the Docker socket.

The images and binaries are intentionally not available during Phase 0. `TABI_BACKEND_IMAGE` must point to the future non-root backend image with `/app/transit-api`, `/app/realtime-poller`, `/app/notification-worker`, `/app/gtfs-importer`, `/app/tabi-migrate`, and `/app/tabi-healthcheck`.

## Validation

`scripts/validate-compose.sh` creates placeholder secret files in a temporary directory and runs Compose interpolation/config validation. It neither contacts providers nor starts containers. Run it after Docker Compose is installed:

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

`deploy.sh` validates a candidate release file, backs up, runs migrations, starts services, checks public readiness, and restores the previous application image set on a failed health check. Database migrations must remain expand/migrate/contract compatible for image rollback.

## Optional Fly adapter

`fly/fly.toml.example` is deliberately separate and not a production plan. It makes no free-tier claim. Use it only after pricing, region, PostGIS, backup, and secret-isolation evidence satisfies ADR-0007; the same OCI image and Compose recovery path remain required.
