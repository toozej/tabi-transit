---
doc_id: TAB-PLAN-021
title: "Optional Fly.io Deployment"
status: optional
last_updated: 2026-07-22
intended_agents: ["infra-agent", "release-agent"]
depends_on: ["TAB-PLAN-010", "TAB-PLAN-011"]
---

# Optional Fly.io Deployment

## Important pricing constraint

Fly.io currently documents that it has no standing free account/free tier. It offers a free trial and may provide credits or allowances, but usage can become billable and allowances do not cap spending. Verify current pricing immediately before deployment.

This is therefore a **trial or optional low-cost target**, not the canonical production plan.

## Portability goal

The same OCI backend image used by Docker Compose must deploy to Fly.io without provider-specific application code. Fly configuration is an adapter around the image:

- Dockerfile/image is canonical;
- environment variables and secret names remain the same;
- OpenAPI, migrations, health endpoints, and database schema remain identical;
- no Fly-only business logic;
- Compose remains the recovery and self-hosting path.

## Preferred Fly topology

Build one backend image containing all Go binaries and use Fly process groups:

```toml
[processes]
web = "/app/transit-api"
poller = "/app/realtime-poller"
worker = "/app/notification-worker"

[http_service]
internal_port = 8080
force_https = true
processes = ["web"]
```

Scale:

- `web=1`;
- `poller=1`;
- `worker=0` until notifications are enabled.

Process groups run in separate Machines and can be sized/scaled independently. Secrets are app-wide to all Machines in that Fly App. If least-privilege separation is more important, deploy separate API, poller, and worker Fly Apps instead.

## Why not use Compose as Fly's primary interface

Fly deploys Docker images through `fly.toml`. Its multi-container Compose support has constraints, including one buildable service and app/Machine-level secret limitations. Keep the production Compose files for Linux self-hosting and maintain small explicit `fly.toml` files for Fly.

## Database options

In preference order:

1. an affordable external managed PostgreSQL service with required PostGIS support and verified network/security/backup;
2. Fly Managed Postgres when its cost is acceptable;
3. self-managed PostgreSQL on a Fly Volume for experimentation only.

Self-managed Fly Postgres is operator-managed, not a substitute for a fully managed database. Fly Volumes are local to Machines and daily snapshots should not be the primary backup. Continue encrypted off-site logical dumps.

Do not colocate production PostgreSQL inside the web Machine or depend on ephemeral Machine storage.

## Static/feed storage

Prefer reconstructable source data plus database state. For files that must persist:

- use a Fly Volume attached only to the process that owns it; or
- move archives/static artifacts to an external object store.

Volumes cannot be treated as a shared filesystem between multiple Machines. API and poller should stay stateless except for database/object-store access.

## Import and scheduled jobs

Options:

- scheduled GitHub Actions invokes a one-shot Fly Machine with the importer command;
- a dedicated lightweight process group runs an internal scheduler;
- an external scheduler calls an authenticated administrative job endpoint only if that endpoint is private and strongly protected.

Keep importer idempotent and database-locked. Do not expose job triggers publicly.

## Secrets

Use `fly secrets` for runtime secrets. The same names used in Compose may be supplied as ordinary environment secrets on Fly. Because Fly app secrets are available to every Machine in an app, separate apps when a secret should not be shared.

CI uses a scoped Fly deploy token in a protected GitHub environment.

## Deployment workflow

1. CI validates/tests the canonical image.
2. Push immutable digest to GHCR or Fly registry.
3. Run database backup.
4. Run release-command/one-shot migration.
5. `fly deploy` the exact image/config.
6. Scale process groups explicitly.
7. Verify public health and source freshness.
8. Record Fly app, region, Machine IDs/sizes, image digest, database endpoint, and cost estimate.
9. Roll back to prior image if health fails.

## Cost controls

- Check Fly Cost Explorer/upcoming invoice frequently.
- Keep one region unless resilience requirements justify more.
- Disable worker until used.
- Use the smallest measured Machine sizes.
- Avoid unnecessary persistent volumes and inter-region traffic.
- Set an external billing reminder/monitor because a free allowance is not a hard spending cap.
- Keep the Linux Compose exit path continuously tested.

## Trial acceptance criteria

- No application code changes are needed.
- The same image digest runs under Compose and Fly.
- Web/poller process groups remain healthy.
- Poller does not auto-stop.
- Database has PostGIS and a verified off-site restore.
- Secrets are not over-shared without an accepted risk.
- Monthly cost is estimated and approved.
- Documentation never calls the deployment a free tier.
- A Fly outage or cost decision can be handled by redeploying to the Linux Compose host.
