# Tabi

Tabi is a Portland-area transit application for iOS and Android. It uses an Expo React Native client and Go services backed by PostgreSQL/PostGIS.

The implementation source of truth is in [docs/implementation-plan](docs/implementation-plan/00_README.md). Current delivery state is tracked in [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md).

## Development status

Phase 0 is in progress. The repository foundation, API contract, mobile compatibility, transit-data, and deployment spikes are being established before production feature work begins.

## Safety and source policy

Tabi uses official sources through backend adapters. Portland Streetcar coverage and Rose City Transit-inspired presentation use TriMet's official feeds and Arrivals V2 interface; Tabi has no direct Streetcar/PBOT/UmoIQ or Rose City integration. Never add provider credentials, production tokens, or `.env` files to Git.

## Repository tooling

Install tooling and locked dependencies with `make prereqs`. Go-based build,
development, and test tools are version-pinned in `tools/go-tools.tsv` and
installed independently into the ignored `.tools/bin` directory; Python-based
tooling is pinned in `tools/requirements.txt` and installed into
`.tools/python`. These isolated tool environments do not modify the application
Go module or the system Python installation.

Run `make pre-commit` to install the local Git hook and execute the complete
repository check suite. Later runs can use `make pre-commit-run`, while
`make pre-commit-install` only installs the hook and its pinned dependencies.

Run `make pre-commit-update` to update pinned Go tools and pre-commit hook
revisions, regenerate checked-in clients, and validate the resulting toolchain.
A scheduled GitHub Actions workflow runs the same target and opens an auto-merge
PR when updates are available. Configure repository secrets named
`AUTOUPDATE_APP_ID` and `AUTOUPDATE_APP_PRIVATE_KEY`, matching the updater GitHub
App installed for this repository, so CI is triggered for its pull requests.
Without both secrets, the workflow is skipped with a warning. Dependabot keeps
application, JavaScript, Python-tool, Actions, and container dependencies current.
