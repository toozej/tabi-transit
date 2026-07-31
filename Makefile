.DEFAULT_GOAL := help
.PHONY: help bootstrap format format-check lint typecheck test test-unit test-integration test-db-migrations test-e2e test-race test-load test-history-benchmark test-vehicle-payload-benchmark test-trimet-live generate generate-check db-up db-migrate dev-api dev-mobile dev-poller build doctor
GO_CACHE_DIR := $(CURDIR)/.cache/go-build

help: ## Display available Make targets
	@grep -E '^[a-zA-Z0-9_-]+ ?:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-24s\033[0m %s\n", $$1, $$2}'

bootstrap: ## Install pinned JavaScript dependencies
	@corepack pnpm install --frozen-lockfile

format: ## Format repository source files
	@corepack pnpm format

format-check: ## Verify repository source formatting
	@corepack pnpm format:check

lint: ## Run repository linters
	@corepack pnpm lint

typecheck: ## Run TypeScript type checks
	@corepack pnpm typecheck

test-unit: ## Run JavaScript, mobile, and Go unit tests
	@corepack pnpm test
	@corepack pnpm --dir apps/mobile test
	@GOCACHE=$(GO_CACHE_DIR) go test ./...

test: test-unit ## Run the full deterministic unit-test suite

test-integration: ## Run deterministic cross-component integration checks
	@test -f deployment/compose.yaml || { echo 'integration tests require deployment/compose.yaml (WP-13)'; exit 1; }
	@test -x tests/integration/run.sh || { echo 'integration test runner missing: tests/integration/run.sh'; exit 1; }
	@# The runner renders Compose with a temporary placeholder-secret directory;
	@# do not require developer or production secrets for deterministic tests.
	@tests/integration/run.sh

test-db-migrations: ## Run disposable PostGIS migration and retention checks
	@db/test-migrations.sh

test-e2e: ## Run Maestro mobile end-to-end tests
	@command -v maestro >/dev/null || { echo 'Maestro CLI is required for mobile E2E tests; see docs/runbooks/local-development.md'; exit 1; }
	@test -d tests/e2e/maestro || { echo 'Maestro flows are not present yet (WP-08/WP-16)'; exit 1; }
	@maestro test tests/e2e/maestro

test-race: ## Run Go tests with the race detector
	@GOCACHE=$(GO_CACHE_DIR) go test -race ./...

test-load: ## Run k6 load tests
	@command -v k6 >/dev/null || { echo 'k6 is required for load tests; see docs/runbooks/local-development.md'; exit 1; }
	@test -d tests/load || { echo 'load test scripts are not present yet (WP-16)'; exit 1; }
	@k6 run tests/load/*.js

test-history-benchmark: ## Measure the maximum bounded vehicle-history API page locally
	@GOCACHE=$(GO_CACHE_DIR) go test ./internal/api -run '^$$' -bench '^BenchmarkVehicleHistoryMaximumPage$$' -benchmem -count=5

test-vehicle-payload-benchmark: ## Compare compact vehicle JSON and GeoJSON serialization locally
	@node tests/performance/vehicle_geojson_payload.mjs

# Explicit opt-in only: sources the user's trusted local .env without printing
# it, then performs one read-only TriMet Arrivals V2 compatibility request.
test-trimet-live: ## Run one authenticated TriMet Arrivals V2 smoke test
	@test -f .env || { echo 'test-trimet-live requires a local .env file'; exit 1; }
	@bash -c 'set -a; . ./.env; set +a; exec env TRIMET_LIVE_SMOKE=1 TABI_TRIMET_ENABLED=true TABI_TRIMET_BASE_URL=https://developer.trimet.org go test -v ./internal/sources/trimet -run TestLiveArrivalsSmoke -count=1'

generate: ## Generate OpenAPI client and sqlc code
	@test -f api/openapi.yaml || { echo 'OpenAPI source missing: api/openapi.yaml (WP-02)'; exit 1; }
	@test -d packages/api-client || { echo 'API client package missing: packages/api-client (WP-02)'; exit 1; }
	@corepack pnpm --filter @tabi/api-client generate
	@test -f db/sqlc.yaml || { echo 'sqlc configuration missing: db/sqlc.yaml (WP-03)'; exit 1; }
	@go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate -f db/sqlc.yaml

generate-check: ## Verify generated API and persistence code is current
	@before=$$(mktemp); after=$$(mktemp); trap 'rm -f "$$before" "$$after"' EXIT; \
	git diff -- packages/api-client/src/generated internal/persistence/sqlcgen > "$$before"; \
	$(MAKE) generate; \
	git diff -- packages/api-client/src/generated internal/persistence/sqlcgen > "$$after"; \
	diff -u "$$before" "$$after"

db-up: ## Start the local PostgreSQL/PostGIS database container
	@test -f deployment/compose.yaml || { echo 'Compose topology missing: deployment/compose.yaml (WP-13)'; exit 1; }
	@docker compose -f deployment/compose.yaml up -d db

db-migrate: ## Apply database migrations
	@test -x db/migrate.sh || { echo 'migration runner missing: db/migrate.sh (WP-03)'; exit 1; }
	@db/migrate.sh

dev-api: ## Run the local transit API
	@test -f services/transit-api/cmd/transit-api/main.go || { echo 'transit API entrypoint missing: services/transit-api/cmd/transit-api/main.go'; exit 1; }
	@go run ./services/transit-api/cmd/transit-api

dev-mobile: ## Run the Expo mobile development client
	@test -f apps/mobile/package.json || { echo 'mobile app missing: apps/mobile/package.json (WP-08)'; exit 1; }
	@corepack pnpm --dir apps/mobile start -- --dev-client

dev-poller: ## Run the local realtime poller
	@test -f services/realtime-poller/cmd/realtime-poller/main.go || { echo 'realtime poller entrypoint missing: services/realtime-poller/cmd/realtime-poller/main.go'; exit 1; }
	@go run ./services/realtime-poller/cmd/realtime-poller

build: ## Build JavaScript packages and Go binaries
	@corepack pnpm build
	@GOCACHE=$(GO_CACHE_DIR) go build ./...

doctor: ## Check local development prerequisites
	@scripts/doctor.sh
