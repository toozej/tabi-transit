# CI and release validation

The GitHub Actions workflows in this repository use no repository, provider,
deployment-host, store, EAS, or signing secrets on pull requests. They check
out with persisted credentials disabled and run only deterministic fixture and
repository validation.

## Merge validation

`Quality and contract validation` runs format, lint, strict TypeScript, Go
unit/race tests, OpenAPI and sqlc generated-code drift, secret/privacy policy,
Compose topology validation, and deterministic API/mobile fixture integration.
It does not contact transit providers or start a production deployment.

Run the same local preflight when the toolchain is installed:

```bash
bash scripts/ci/validate-workflows.sh
make format-check lint typecheck test test-race generate-check
bash scripts/quality-security-check.sh
bash deployment/scripts/validate-compose.sh
tests/integration/run.sh
```

`actionlint` is optional locally; when absent the validator still parses every
workflow and checks its required trigger/job structure. The CI workflow runs a
pinned `actionlint` release for expression-level validation.

## Release boundary

`Release input validation` is manual and accepts only an immutable lowercase
`registry/image@sha256:<64 hex>` reference. It never pushes, signs, deploys, or
contacts a registry. It intentionally fails until a reviewed backend
`Dockerfile` exists, preventing a green release workflow that has not built an
image.

Container build, vulnerability scanning, SBOM/provenance generation, signing,
GHCR publishing, protected-environment deployment, and rollback rehearsal are
disabled pending a reviewed Dockerfile and the required GitHub environment,
registry, signing, host, and backup evidence. The local deployment-script
regression check proves only that an unhealthy candidate does not replace the
active `release.env`; it is not a host rollback rehearsal. Do not add those
credentials to workflow YAML, repository variables, or pull-request contexts.

EAS build/submit and Maestro device runs are also disabled in CI until EAS
credentials and supported iOS/Android development-build runners are available.
The workstation waiver does not replace the physical-device release gate.

When those gates are satisfied, add a separately reviewed workflow that uses
protected environments, immutable image digests, short-lived/scoped registry
credentials, pinned action revisions, SBOM/provenance artifacts, and an
explicit deployment approval. The canonical deployment remains Docker Compose;
Fly.io is not a release prerequisite.
