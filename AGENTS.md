# Repository Guidelines

## Project Structure

- Module: github.com/toozej/tabi-transit
- Mobile app: apps/mobile/ is the Expo React Native client, with screens under app/, application code under src/, and simulator helpers under scripts/.
- Web app: apps/web/ is the Vite client. Shared workspace packages live under packages/, including the generated API client, design tokens, and transit-domain code.
- Go code: internal/ contains shared internal application packages; independently runnable services live under services/<service>/cmd/<service>/.
- API and database: api/ contains the OpenAPI contract and examples; db/ contains migrations, queries, fixtures, and sqlc configuration.
- Deployment and operations: deployment/ contains Compose, Fly, web, systemd, and secret-management configuration. Keep credentials and local environment files out of Git.
- Documentation and tests: docs/ contains architecture, implementation, runbook, and security documentation. Cross-component and specialized tests live under tests/.
- Generated API-client and sqlc output is checked in. Update the source contract or schema and use the appropriate Make target to regenerate it rather than editing generated files by hand.

## Build and Test Commands

Use mktemp -d with trap to automatically clean up any temporary directories needed for build and test work.

The root Makefile is the canonical entry point. Run make help to see the complete current target list. The main workflows are:

- Install pinned tools and dependencies: make prereqs
- Format and validate formatting: make format, make format-check
- Run static checks: make lint, make typecheck
- Run the deterministic unit suite: make test
- Run focused suites: make test-js, make test-web, make test-go, or make test-race
- Run integration and database checks: make test-integration, make test-db-migrations
- Regenerate and verify generated code: make generate, make generate-check
- Build all JavaScript packages and Go binaries: make build
- Run repository checks and the build: make all
- Install and run the complete pre-commit workflow: make pre-commit
- Run local services: make dev-api, make dev-poller, make dev-web, or make dev-mobile

Use make db-up to start the local PostgreSQL/PostGIS container and make db-migrate to apply migrations. The mobile simulator targets (make ios-simulators, make android-simulators, and the device-specific targets) require the corresponding local platform tooling. make test-e2e and make test-load require Maestro and k6 respectively. make test-trimet-live is an explicit authenticated smoke test and must only be run when a trusted local .env is present.

For changes that affect multiple layers, run the narrowest relevant checks first, then make format-check, make lint, make typecheck, and make test. Run make generate-check when changing OpenAPI, database schemas, queries, or generation configuration.

## Development and Data Safety

- Never commit .env files, production credentials, API tokens, private keys, generated local artifacts, coverage output, binaries, or tool caches.
- Use the repository-local pinned tool environments installed by make prereqs; do not modify the application Go module or system Python installation to install development tools.
- Treat fixtures and provider adapters as source-controlled test data. Use official provider interfaces through backend adapters and do not add direct provider credentials to clients.
- Keep database and deployment changes documented alongside their migrations or configuration, and use disposable local resources for tests where possible.

## Style and Tooling

- Keep Go code formatted and imports organized through the repository Make targets.
- Keep TypeScript and JavaScript formatted, linted, and type-checked with make format, make lint, and make typecheck.
- Prefer root Make targets over direct tool invocations so pinned versions and repository-local caches are used.
- Keep generated clients and persistence code current with make generate; do not hand-edit generated output.
- Before committing, use make pre-commit or at minimum the relevant formatting, lint, typecheck, generation, and test targets.
