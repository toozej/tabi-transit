# Fixture sanitization policy

Fixtures are deterministic test inputs, never captures of a rider, a local
environment, or a credentialed production session. This policy applies to
`tests/fixtures/`, `db/fixtures/`, and spike fixtures.

Allowed data is synthetic or provider data deliberately reduced and sanitized
for the test case. Transit stop and vehicle coordinates are allowed when they
represent public transit entities needed for a mapping or spatial test. Use
invented identifiers and bounded sample data where possible.

Do not commit:

- any `.env` file, token, private key, provider AppID, push token, signing
  material, or authenticated response;
- a user's current/device coordinates, address, or full free-text search;
- diagnostic output containing request headers or identifiers from a real app;
- a wholesale provider export when a small sanitized fixture proves the case.

When adding a fixture, record its source class and transformations in that
fixture directory's README. If the fixture originated from a provider response,
remove credentials, headers, account identifiers, and unrelated entities before
commit. Preserve only the fields necessary to test the normalization or error
case.

Run the local guard before handoff:

```sh
bash scripts/security/verify-fixture-policy.sh
bash tests/security/fixture_policy_test.sh
```

The guard detects recognizable credential signatures, public Mapbox tokens,
TriMet AppID assignments, push tokens, and explicit user search/location
fixtures. It is a regression check, not proof that all fixture data is safe;
reviewers still assess provenance and minimization.
