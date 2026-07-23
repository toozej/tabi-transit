# Test layout

- `unit/` contains framework-independent TypeScript tests.
- `contract/` is reserved for OpenAPI contract tests (WP-02/WP-07).
- `integration/` is reserved for database and service integration tests.
- `e2e/maestro/` is reserved for device flows (WP-08/WP-16).
- `load/` is reserved for k6 scenarios (WP-16).
- `fixtures/upstream/` contains only sanitized deterministic provider fixtures.

Tests must not routinely call production providers or contain credentials.
