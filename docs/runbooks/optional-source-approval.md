# Optional-source approval review

Use this review before enabling a direct optional external source. It is an
evidence checklist, not authorization to
contact a provider, fetch a feed, or scrape a public site.

## Required record for each source

- Written approval from the source owner or a documented canonical public API.
- Source owner, agency/provider relationship, technical contact, and review
  date.
- Feed/API documentation, authentication method, rate limits, expected
  freshness, and outage behavior.
- License and redistribution rights, including static, realtime, and historical
  data separately where applicable.
- Retention, caching, attribution, and credit requirements.
- Explicit answer on whether the source overlaps TriMet data and the approved
  deduplication identity/rule.
- Sanitized deterministic fixtures approved for repository use.
- Security/privacy review of credentials, logs, and any location/history data.
- Feature-flag, rollback, source-health, test, and runbook plan.

## Decision record

Record the evidence in a new or amended ADR and update `docs/BLOCKERS.md`
before implementation. The ADR must name the applicable source, scope,
retention, attribution, rate/freshness policy, overlap/deduplication rules, and
the owner who approved it.

Until that ADR is accepted, preserve the disabled configuration response and do
not add an adapter, provider-derived fixture, specialist history/adherence
view, or source credit claim. This runbook does not apply to Streetcar or Rose
City-inspired presentation data
already supplied through TriMet's official GTFS, GTFS-Realtime, or Arrivals V2
interfaces; that coverage stays within the approved TriMet boundary and must not add a
direct Streetcar/PBOT/UmoIQ/Rose City dependency.
