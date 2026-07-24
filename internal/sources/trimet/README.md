# TriMet Web Services boundary

This adapter is feature-disabled by default. It uses only sanitized fixtures in
tests and makes no provider calls unless `TRIMET_ENABLED=true` and a valid
`TRIMET_APP_ID_FILE` (preferred) or `TRIMET_APP_ID` are configured.

For local development, the repository-standard `TABI_TRIMET_*` names are also
accepted (including `TABI_TRIMET_APP_ID` and `TABI_TRIMET_APP_ID_FILE`). The
Compose service uses unprefixed `TRIMET_*` names. A `.env` file is deliberately
not loaded implicitly by Go binaries; inject it through the local process
runner without printing it.

Before an application feature can be enabled, record D-001 evidence for a
registered AppID, the applicable Web Services terms, rate limits,
cache/retention rules, required attribution, and beta-endpoint expectations.
Configure an HTTPS
`TRIMET_BASE_URL` whose host is in `TRIMET_ALLOWED_HOSTS` (default:
`developer.trimet.org`), a bounded `TRIMET_TIMEOUT`, and set
`TRIMET_PLANNER_ENABLED=true` only after planner-specific review. The AppID
must remain in a local secret file or Docker secret, never in source, logs, or
client configuration.

The source-facing DTOs are private. New provider fields/endpoints require a
sanitized deterministic fixture and mapper test before feature enablement.

Planning is separately gated by `TRIMET_PLANNER_ENABLED=true`; it accepts only
normalized place references and bounded mode, transfer, walking, and
accessibility preferences. The adapter returns source-neutral itinerary/leg
summaries and freshness, never raw provider payloads or unreviewed geometry.
No planner request is enabled until D-001 specifically records planner terms,
rate limits, cache/retention, attribution, and supported constraint semantics.

The authenticated live smoke test is intentionally opt-in and uses only the
Arrivals V2 endpoint. It reads a local secret file without printing it:

```sh
TRIMET_LIVE_SMOKE=1 TRIMET_ENABLED=true \
TRIMET_BASE_URL=https://developer.trimet.org \
TRIMET_APP_ID_FILE=deployment/secrets/trimet_app_id \
go test ./internal/sources/trimet -run TestLiveArrivalsSmoke -count=1
```

It does not enable an application feature, invoke the planner, retain a raw
response, or replace deterministic fixture tests.
