# Local development

## Required tools

The Phase 0 baseline pins Node.js `>=20.19.0 <23`, pnpm `10.13.1` through
Corepack, and Go `1.26.x`. Docker Engine with the Compose plugin is needed for
database, service integration, and deployment checks. The exact Expo, React
Native, Mapbox, PostGIS, Caddy, and restic versions are selected by the
architecture/version-matrix work package; do not infer them from this runbook.

Run `make doctor` first. It exits non-zero if a required local prerequisite is
unavailable; it does not claim success for optional tools it cannot verify.

## Bootstrap and routine checks

```sh
make bootstrap
make format-check lint typecheck test
```

`make format` writes formatting changes. `make format-check` is the non-mutating
check used by CI. `make test` runs TypeScript unit tests and `go test ./...`.

The following commands deliberately fail until their owning work package has
provided its runnable implementation instead of returning a false successful
placeholder:

- `make test-integration` requires WP-13 Compose and an integration runner.
- `make test-e2e` requires the Maestro CLI plus WP-08/WP-16 flows.
- `make test-load` requires k6 plus WP-16 scenarios.
- `make generate` and `make generate-check` require the WP-02 OpenAPI source
  and generated API client package.
- `make db-up` requires the WP-13 Compose topology; `make db-migrate` requires
  the WP-03 migration runner.
- `make dev-api` and `make dev-poller` require their service entrypoints;
  `make dev-mobile` requires the WP-08 Expo application.

## Local configuration

Copy `.env.example` to `.env` only for local use. `.env` is ignored by Git. Use
the `_FILE` form for secrets where supported, particularly in Docker Compose.
Never commit provider AppIDs, Mapbox tokens, signing keys, backups, or production
database URLs. Provider functionality remains disabled unless its documented
credential and terms gate is satisfied.

## Mobile and device checks

Mobile uses Expo development builds, not Expo Go. After WP-08 adds the app and
native setup, use `make dev-mobile`. Physical iOS/Android validation and Maestro
require an available simulator/emulator or physical device; a build-time check is
not evidence that Mapbox native rendering works on a device.

## Service and database checks

After the relevant packages land, use `make db-up`, `make db-migrate`,
`make dev-api`, and `make dev-poller`. PostgreSQL and metrics must remain private
to the Compose network. Use sanitized fixtures for normal test runs instead of
calling public transit sources.
