# TriMet Web Services boundary

This adapter is feature-disabled by default. It uses only sanitized fixtures in
tests and makes no provider calls unless `TRIMET_ENABLED=true` and a valid
`TRIMET_APP_ID_FILE` (preferred) or `TRIMET_APP_ID` are configured.

Before a real smoke test can be enabled, record D-001 evidence for a registered
AppID, the applicable Web Services terms, rate limits, cache/retention rules,
required attribution, and beta-endpoint expectations. Configure an HTTPS
`TRIMET_BASE_URL` whose host is in `TRIMET_ALLOWED_HOSTS` (default:
`ws.trimet.org`), a bounded `TRIMET_TIMEOUT`, and set
`TRIMET_PLANNER_ENABLED=true` only after planner-specific review. The AppID
must remain in a local secret file or Docker secret, never in source, logs, or
client configuration.

The source-facing DTOs are private. New provider fields/endpoints require a
sanitized deterministic fixture and mapper test before feature enablement.
