# Local development

## Required tools

The Phase 0 baseline pins Node.js `>=20.19.0 <23`, pnpm `10.13.1` through
Corepack, and Go `1.26.x`. Docker Engine with the Compose plugin is needed for
database, service integration, and deployment checks. The exact Expo, React
Native, Mapbox, PostGIS, Caddy, and restic versions are selected by the
architecture/version-matrix work package; do not infer them from this runbook.

Run `make doctor` first. It exits non-zero if a required local prerequisite is
unavailable; it does not claim success for optional tools it cannot verify.

## Prerequisites and routine checks

```sh
make prereqs
make format-check lint typecheck test
```

`make format` writes formatting changes. `make format-check` is the non-mutating
check used by CI. `make test` runs TypeScript unit tests and `go test ./...`.

## Web application

The web app is runnable without provider credentials in deterministic fixture
mode:

```sh
make prereqs-web
make dev-web
```

`prereqs-web` installs the locked workspace dependencies and generates the
shared design-token CSS that Vite imports. `dev-web` repeats that safe setup so
a fresh checkout does not fail on a missing generated token artifact.

To exercise the Tabi API locally, copy `apps/web/.env.example` to the ignored
`apps/web/.env.local`, set `VITE_TABI_DATA_MODE=remote`, and keep
`VITE_TABI_API_BASE_URL=/v1`. Set the Vite-only proxy target in that same file:

```dotenv
TABI_WEB_API_PROXY=http://127.0.0.1:8080
```

Then start the local API separately after its database/configuration is ready.
The proxy value is read only by Vite's Node process; browser requests remain
same-origin at `/v1`.

Browser maps are disabled pending D-004. There is intentionally no Mapbox
credential setting for the web app today. Never put a Mapbox secret, Search
token, or any privileged provider credential in `VITE_*`: those values are
compiled into the public browser bundle. A future approved browser map may use
only a domain-restricted public token in `apps/web/.env.local`; sensitive
credentials remain server-side secret files.

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

On the configured Intel macOS host, `make prereqs` installs the iOS and Android
development components. Use `make ios-simulators` and `make android-simulators`
to create the configured virtual devices; device-specific launch targets are
listed by `make help` and documented in `apps/mobile/README.md`.

## Service and database checks

After the relevant packages land, use `make db-up`, `make db-migrate`,
`make dev-api`, and `make dev-poller`. PostgreSQL and metrics must remain private
to the Compose network. Use sanitized fixtures for normal test runs instead of
calling public transit sources.
