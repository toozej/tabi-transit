# Set sane defaults for Make.
SHELL = bash
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

.DEFAULT_GOAL := help

GO_CACHE_DIR := $(CURDIR)/.cache/go-build

# Go-based build, development, and test tools are installed independently as
# package@version into a repository-local bin directory. Their module graphs do
# not affect the application module or one another.
TOOLS_DIR := $(CURDIR)/.tools
TOOLS_BIN := $(TOOLS_DIR)/bin
NVM_DIR := $(TOOLS_DIR)/nvm
NODE_RUN := $(CURDIR)/scripts/with-node.sh
PYTHON_TOOLS_VENV := $(TOOLS_DIR)/python
GO_TOOLS := $(CURDIR)/scripts/manage-go-tools.sh
GO_TOOL_MANIFEST ?= $(CURDIR)/tools/go-tools.tsv
PYTHON_TOOL_REQUIREMENTS := $(CURDIR)/tools/requirements.txt
# Use the character code for "#" because an unescaped # starts a Make comment.
GO_TOOL_NAMES := $(shell if test -f "$(GO_TOOL_MANIFEST)"; then awk -F '\t' 'NF && substr($$1, 1, 1) != sprintf("%c", 35) { print $$1 }' "$(GO_TOOL_MANIFEST)"; fi)
GO_TOOL_INSTALL_TARGETS := $(addsuffix -install,$(GO_TOOL_NAMES))
export PATH := $(TOOLS_BIN):$(PYTHON_TOOLS_VENV)/bin:$(PATH)
export PRE_COMMIT_HOME := $(CURDIR)/.cache/pre-commit
export SEMGREP_SETTINGS_FILE := $(CURDIR)/.cache/semgrep/settings.yml
export XDG_CACHE_HOME := $(CURDIR)/.cache

.PHONY: all clean help prereqs prereqs-web prereqs-mobile prereqs-ios-simulators prereqs-android-simulators bootstrap bootstrap-ios-simulators format format-check lint typecheck test test-unit test-js test-web test-web-preview test-go test-integration test-db-migrations test-e2e test-race test-load test-history-benchmark test-vehicle-payload-benchmark test-trimet-live generate generate-check db-up db-migrate dev-api dev-mobile dev-web ios-simulators ios-iphone-13-mini ios-iphone-air android android-simulators android-motorola-razr-2024 android-pixel-10-pro android-sony-xperia-1-ii stop-ios-simulators stop-android-simulators stop-simulators dev-poller build doctor nvm-install nvm-update node-check update-dependencies tools-install go-tools-install python-tools-install pre-commit-tools-install pre-commit-install-no-prereqs pre-commit-run pre-commit-run-no-generate pre-commit-update pre-commit licenses
.PHONY: $(GO_TOOL_INSTALL_TARGETS)

all: pre-commit-run build ## Run repository checks and build all applications

clean: ## Remove local repository tool and cache artifacts
	@rm -rf "$(TOOLS_DIR)" "$(CURDIR)/.cache"

help: ## Display available Make targets
	@grep -E '^[a-zA-Z0-9_-]+ ?:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-24s\033[0m %s\n", $$1, $$2}'

prereqs: tools-install prereqs-mobile ## Install pinned dependencies, tools, and host-specific prerequisites
	@$(MAKE) prereqs-ios-simulators
	@$(MAKE) prereqs-android-simulators

nvm-install: ## Install nvm through Homebrew for repository Node.js tooling
	@command -v brew >/dev/null 2>&1 || { echo 'Homebrew is required to install nvm.' >&2; exit 1; }
	@if ! brew list --versions nvm >/dev/null 2>&1; then \
		brew install nvm; \
	fi
	@nvm_script="$$(brew --prefix nvm)/nvm.sh"; test -s "$$nvm_script" || { echo "Homebrew nvm script is unavailable: $$nvm_script" >&2; exit 1; }
	@mkdir -p "$(NVM_DIR)"

nvm-update: ## Update nvm through Homebrew and retain its repository-local Node data
	@command -v brew >/dev/null 2>&1 || { echo 'Homebrew is required to update nvm.' >&2; exit 1; }
	@brew update
	@if brew list --versions nvm >/dev/null 2>&1; then brew upgrade nvm; else brew install nvm; fi
	@$(MAKE) nvm-install

node-check: nvm-install ## Check that nvm selects a supported Node.js runtime
	@$(NODE_RUN) bash -c 'command -v corepack >/dev/null 2>&1 || { echo "Corepack is required by the Node.js version in .nvmrc." >&2; exit 1; }; node_version="$$(node --version)"; node_without_prefix="$${node_version#v}"; node_major="$${node_without_prefix%%.*}"; node_minor_and_patch="$${node_without_prefix#*.}"; node_minor="$${node_minor_and_patch%%.*}"; if [ "$$node_major" -eq 24 ] && [ "$$node_minor" -ge 19 ]; then :; else echo "Unsupported Node.js version $$node_version; require >=24.19.0 and <25 (see .nvmrc)." >&2; exit 1; fi'

update-dependencies: node-check ## Update JavaScript, Go, Python, and pre-commit dependency/toolchain pins
	@$(MAKE) nvm-update
	@CI=true $(NODE_RUN) corepack pnpm install --no-frozen-lockfile
	@CI=true $(NODE_RUN) corepack use pnpm@latest
	@CI=true $(NODE_RUN) corepack pnpm update --recursive --latest
	@$(NODE_RUN) corepack pnpm add --save-dev --workspace-root typescript@5.9.3 @types/node@24
	@$(NODE_RUN) corepack pnpm --dir apps/mobile exec expo install --fix
	@$(NODE_RUN) corepack pnpm --dir apps/mobile add --save-dev @react-native/metro-config@0.86.2 react-dom@19.2.3 react-test-renderer@19.2.3
	@$(NODE_RUN) corepack pnpm --dir apps/web add react@19.2.3 react-dom@19.2.3
	@GOWORK=off GOTOOLCHAIN=auto go get -u ./...
	@GOWORK=off GOTOOLCHAIN=auto go mod tidy
	@cd spikes/transit-data && GOWORK=off GOTOOLCHAIN=auto go get -u ./... && GOWORK=off GOTOOLCHAIN=auto go mod tidy
	@TOOLS_BIN="$(TOOLS_BIN)" $(GO_TOOLS) update
	@$(MAKE) go-tools-install
	@scripts/update-python-tools.sh
	@$(MAKE) python-tools-install
	@GOWORK=off pre-commit autoupdate
	@$(MAKE) generate

prereqs-mobile: node-check ## Install locked mobile workspace dependencies
	@$(NODE_RUN) corepack pnpm install --frozen-lockfile

prereqs-web: node-check ## Install web dependencies and generate its shared token CSS
	@$(NODE_RUN) corepack pnpm install --frozen-lockfile
	@$(NODE_RUN) corepack pnpm --filter @tabi/design-tokens build

prereqs-ios-simulators: prereqs-mobile ## Install mobile dependencies, Xcode components, and configured iOS runtimes on macOS
	@if test "$$(uname -s)" = Darwin; then \
		$(NODE_RUN) corepack pnpm --dir apps/mobile ios:prereqs; \
	else \
		echo 'Skipping iOS Simulator prerequisites (requires macOS)'; \
	fi

prereqs-android-simulators: prereqs-mobile ## Install mobile dependencies, Android Studio, SDK tools, and configured runtimes on Intel macOS
	@if test "$$(uname -s)" = Darwin; then \
		$(NODE_RUN) corepack pnpm --dir apps/mobile android:prereqs; \
	else \
		echo 'Skipping Android Studio prerequisites (configured for Intel macOS)'; \
	fi

# Compatibility aliases for existing local workflows.
bootstrap: prereqs
bootstrap-ios-simulators: prereqs-ios-simulators

format: node-check ## Format repository source files
	@$(NODE_RUN) corepack pnpm format

format-check: node-check ## Verify repository source formatting
	@$(NODE_RUN) corepack pnpm format:check

lint: node-check ## Run repository linters
	@$(NODE_RUN) corepack pnpm lint

typecheck: node-check ## Run TypeScript type checks
	@$(NODE_RUN) corepack pnpm typecheck
	@$(NODE_RUN) corepack pnpm --dir apps/web typecheck

test-js: node-check ## Run JavaScript and mobile unit tests
	@$(NODE_RUN) corepack pnpm test
	@$(NODE_RUN) corepack pnpm --dir apps/mobile test
	@$(MAKE) test-web

test-web: node-check ## Run deterministic web unit tests
	@$(NODE_RUN) corepack pnpm --dir apps/web test
	@tests/quality/web_deployment_policy_test.sh

test-web-preview: node-check ## Build the web app and start a local immutable-asset preview
	@$(NODE_RUN) corepack pnpm --dir apps/web build
	@$(NODE_RUN) corepack pnpm --dir apps/web preview:smoke

test-go: ## Run Go unit tests
	@GOCACHE=$(GO_CACHE_DIR) go test ./...

test-unit: test-js test-go ## Run JavaScript, mobile, and Go unit tests

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

test-vehicle-payload-benchmark: node-check ## Compare compact vehicle JSON and GeoJSON serialization locally
	@$(NODE_RUN) node tests/performance/vehicle_geojson_payload.mjs

# Explicit opt-in only: sources the user's trusted local .env without printing
# it, then performs one read-only TriMet Arrivals V2 compatibility request.
test-trimet-live: ## Run one authenticated TriMet Arrivals V2 smoke test
	@test -f .env || { echo 'test-trimet-live requires a local .env file'; exit 1; }
	@bash -c 'set -a; . ./.env; set +a; exec env TRIMET_LIVE_SMOKE=1 TABI_TRIMET_ENABLED=true TABI_TRIMET_BASE_URL=https://developer.trimet.org go test -v ./internal/sources/trimet -run TestLiveArrivalsSmoke -count=1'

generate: node-check sqlc-install ## Generate OpenAPI client and sqlc code
	@test -f api/openapi.yaml || { echo 'OpenAPI source missing: api/openapi.yaml (WP-02)'; exit 1; }
	@test -d packages/api-client || { echo 'API client package missing: packages/api-client (WP-02)'; exit 1; }
	@$(NODE_RUN) corepack pnpm --filter @tabi/api-client generate
	@test -f db/sqlc.yaml || { echo 'sqlc configuration missing: db/sqlc.yaml (WP-03)'; exit 1; }
	@sqlc generate -f db/sqlc.yaml

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

dev-mobile: node-check ## Run the Expo mobile development client
	@test -f apps/mobile/package.json || { echo 'mobile app missing: apps/mobile/package.json (WP-08)'; exit 1; }
	@$(NODE_RUN) corepack pnpm --dir apps/mobile start -- --dev-client

dev-web: prereqs-web ## Run the Vite web application
	@test -f apps/web/package.json || { echo 'web app missing: apps/web/package.json'; exit 1; }
	@$(NODE_RUN) corepack pnpm --dir apps/web dev

ios-simulators: prereqs-ios-simulators ## Install iOS prerequisites and create the configured iOS 26 simulators
	@$(NODE_RUN) corepack pnpm --dir apps/mobile ios:simulators

ios-iphone-13-mini: prereqs-ios-simulators ## Install iOS prerequisites, then build and run mobile on iPhone 13 mini with latest installed iOS 26
	@$(NODE_RUN) corepack pnpm --dir apps/mobile ios:iphone-13-mini

ios-iphone-air: prereqs-ios-simulators ## Install iOS prerequisites, then build and run mobile on iPhone Air with latest installed iOS 26
	@$(NODE_RUN) corepack pnpm --dir apps/mobile ios:iphone-air

android-simulators: prereqs-android-simulators ## Install Android prerequisites and create the configured Android 12 and 16 emulators
	@$(NODE_RUN) corepack pnpm --dir apps/mobile android:simulators

android: prereqs-android-simulators ## Install Android prerequisites, then build and run mobile on the default Motorola Razr 2024 profile
	@$(NODE_RUN) corepack pnpm --dir apps/mobile android

android-motorola-razr-2024: prereqs-android-simulators ## Install Android prerequisites, then build and run mobile on Motorola Razr 2024 with Android 16
	@$(NODE_RUN) corepack pnpm --dir apps/mobile android:motorola-razr-2024

android-pixel-10-pro: prereqs-android-simulators ## Install Android prerequisites, then build and run mobile on Pixel 10 Pro with Android 16
	@$(NODE_RUN) corepack pnpm --dir apps/mobile android:pixel-10-pro

android-sony-xperia-1-ii: prereqs-android-simulators ## Install Android prerequisites, then build and run mobile on Sony Xperia 1 II with Android 12
	@$(NODE_RUN) corepack pnpm --dir apps/mobile android:sony-xperia-1-ii

stop-ios-simulators: ## Forcefully stop every running iOS Simulator
	@if test "$$(uname -s)" = Darwin; then \
		xcrun simctl shutdown all || true; \
		pkill -KILL -x Simulator || true; \
	else \
		echo 'Skipping iOS Simulator shutdown (requires macOS)'; \
	fi

stop-android-simulators: ## Forcefully stop every running Android emulator
	@if test "$$(uname -s)" = Darwin; then \
		adb_path="$${ANDROID_SDK_ROOT:-$${ANDROID_HOME:-$$HOME/Library/Android/sdk}}/platform-tools/adb"; \
		if test ! -x "$$adb_path"; then echo "adb is unavailable; run make prereqs-android-simulators" >&2; exit 1; fi; \
		"$$adb_path" devices | awk '$$1 ~ /^emulator-/ && $$2 == "device" { print $$1 }' | while IFS= read -r serial; do \
			"$$adb_path" -s "$$serial" emu kill || true; \
		done; \
		pkill -KILL -f 'qemu-system.*-avd' || true; \
	else \
		echo 'Skipping Android emulator shutdown (requires macOS)'; \
	fi

stop-simulators: stop-ios-simulators stop-android-simulators ## Forcefully stop every running iOS Simulator and Android emulator

dev-poller: ## Run the local realtime poller
	@test -f services/realtime-poller/cmd/realtime-poller/main.go || { echo 'realtime poller entrypoint missing: services/realtime-poller/cmd/realtime-poller/main.go'; exit 1; }
	@go run ./services/realtime-poller/cmd/realtime-poller

build: node-check ## Build JavaScript packages and Go binaries
	@$(NODE_RUN) corepack pnpm build
	@GOCACHE=$(GO_CACHE_DIR) go build ./...

doctor: ## Check local development prerequisites
	@scripts/doctor.sh

tools-install: nvm-install go-tools-install python-tools-install ## Install every pinned repository tool

go-tools-install: $(GO_TOOL_INSTALL_TARGETS) ## Install every Go tool pinned in tools/go-tools.tsv

$(GO_TOOL_INSTALL_TARGETS): %-install:
	@TOOLS_BIN="$(TOOLS_BIN)" $(GO_TOOLS) install $*

python-tools-install: ## Install pinned Python-based repository tools into .tools/python
	@mkdir -p "$(TOOLS_DIR)" "$(dir $(SEMGREP_SETTINGS_FILE))"
	@if test ! -x "$(PYTHON_TOOLS_VENV)/bin/pre-commit" || \
		test ! -f "$(PYTHON_TOOLS_VENV)/.requirements.txt" || \
		! cmp -s "$(PYTHON_TOOL_REQUIREMENTS)" "$(PYTHON_TOOLS_VENV)/.requirements.txt"; then \
		python3 -m venv "$(PYTHON_TOOLS_VENV)"; \
		"$(PYTHON_TOOLS_VENV)/bin/python" -m pip install --disable-pip-version-check --requirement "$(PYTHON_TOOL_REQUIREMENTS)"; \
		cp "$(PYTHON_TOOL_REQUIREMENTS)" "$(PYTHON_TOOLS_VENV)/.requirements.txt"; \
	else \
		echo "Skipping Python tools (pinned requirements already installed)"; \
	fi

pre-commit-tools-install: tools-install ## Install pinned tools and pre-commit hook environments
	@mkdir -p "$(PRE_COMMIT_HOME)" "$(dir $(SEMGREP_SETTINGS_FILE))"
	@GOWORK=off pre-commit install-hooks

pre-commit-install: pre-commit-tools-install ## Install the local Git pre-commit hook
	@$(MAKE) pre-commit-install-no-prereqs

pre-commit-install-no-prereqs:
	@GOWORK=off pre-commit install
	@hook="$$(git rev-parse --git-path hooks/pre-commit)"; \
	if test -f "$$hook" && ! grep -q 'tabi-tools-path' "$$hook"; then \
		tmp="$$(mktemp)"; \
		{ head -n 1 "$$hook"; \
		  echo 'export PATH="$(TOOLS_BIN):$(PYTHON_TOOLS_VENV)/bin:$$PATH"  # tabi-tools-path'; \
		  echo 'export PRE_COMMIT_HOME="$(PRE_COMMIT_HOME)"'; \
		  echo 'export SEMGREP_SETTINGS_FILE="$(SEMGREP_SETTINGS_FILE)"'; \
		  echo 'export XDG_CACHE_HOME="$(XDG_CACHE_HOME)"'; \
		  tail -n +2 "$$hook"; } >"$$tmp"; \
		cat "$$tmp" >"$$hook"; \
		rm -f "$$tmp"; \
		echo "Prepended repository tooling to PATH in $$hook"; \
	fi

pre-commit-run: pre-commit-tools-install prereqs ## Run all pre-commit, vulnerability, and license checks
	@$(MAKE) pre-commit-run-no-generate

pre-commit-run-no-generate: node-check
	@$(NODE_RUN) env GOWORK=off pre-commit run --all-files
	@govulncheck ./...
	@$(MAKE) licenses

pre-commit-update: python-tools-install prereqs ## Update pinned Go tools and hook revisions, regenerate, and verify
	@TOOLS_BIN="$(TOOLS_BIN)" $(GO_TOOLS) update
	@$(MAKE) go-tools-install
	@GOWORK=off pre-commit autoupdate
	@$(MAKE) generate
	@$(MAKE) pre-commit-run-no-generate

pre-commit: pre-commit-install pre-commit-run ## Install and run the repository pre-commit workflow

licenses: go-licenses-install ## Report third-party Go dependency licenses
	@go-licenses report ./...
