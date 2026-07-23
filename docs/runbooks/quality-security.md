# Quality, security, privacy, and accessibility gates

Run the repository-local policy checks before integration:

```sh
bash scripts/quality-security-check.sh
bash tests/quality/quality_security_check_test.sh
git diff --check
```

The checker scans implementation/configuration/fixture areas, excluding
generated dependencies and environment templates. It blocks concrete leaked
credential signatures, raw coordinate logging, public Postgres/admin/metrics
ports in `deployment/compose.yaml`, Docker socket mounts, missing sensitive
file ignores, and a map-only mobile implementation. It prints `GATE` for
unavailable future components or evidence; a gate is not a pass claim.

For every release candidate, additionally run the pinned formatter, linters,
type checks, unit/integration/contract suites, Compose validation, image and
dependency scanning, and the applicable device matrix in
`tests/accessibility/DEVICE_TEST_MATRIX.md`. Exercise the failure catalogue in
`tests/quality/failure-injection-catalog.md` through component-owned tests and
record actual environment, command, result, and limitations.

Never place real provider captures, user coordinates, full search text, push
tokens, or credentials in fixtures or logs. Use synthetic/sanitized data and
record a credential, legal, device, or production-host dependency as pending
until it has actual evidence.
