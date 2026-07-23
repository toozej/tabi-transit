# Security and privacy validation scaffold

Run `scripts/quality-security-check.sh` for the dependency-free policy checks.
It fails only on concrete repository violations: recognisable production-secret
material, raw coordinate values passed to logging calls, public Compose ports
for Postgres/admin/metrics, Docker socket mounts, missing sensitive-file ignore
rules, or a Mapbox implementation with no accessible list/semantic fallback.

The following remain required release gates, not claims made by this scaffold:

- secret, dependency, container, SBOM, and license scans after their tools are
  pinned;
- API input-abuse and log-redaction tests once handlers exist;
- staging DAST and deployment ownership/mode review;
- manual review of provider terms, Mapbox storage/telemetry obligations, store
  privacy disclosures, and credential rotation rehearsal.

Fixtures must be deterministic and sanitized. Never add live credentials,
precise user locations, push tokens, full searches, or production-provider
captures without an approved fixture sanitization review.
