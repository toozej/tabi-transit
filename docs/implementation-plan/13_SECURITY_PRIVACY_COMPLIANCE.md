---
doc_id: TAB-PLAN-013
title: "Security, Privacy, Licensing and Compliance"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["security-agent", "privacy-agent", "legal-coordinator", "all-agents"]
depends_on: ["TAB-PLAN-001", "TAB-PLAN-002", "TAB-PLAN-010"]
---


# Security, Privacy, Licensing and Compliance

## Classification

Public transit data remains subject to terms/attribution. User/device data includes installation ID, push token, subscriptions, request coordinates, local favorites/recents, diagnostics. Secrets include TriMet AppID, Mapbox server/download tokens, cloud/DB/push/store/signing credentials.

## Privacy defaults

No account, advertising SDK, sale of data, or background location in MVP. Foreground location only for explicit operations. Raw coordinates/search text excluded from access logs. History local unless a subscription requires server state. Clear local data and delete installation. Analytics minimal and privacy-reviewed.

## Threats

Credential/quota theft, API abuse/DoS, malformed feeds, supply-chain compromise, malicious deep links, notification spam, location/token logs, CI theft, drift, stale/false display, Mapbox cost attack.

## Mobile controls

Minimum public Mapbox token; SecureStore; Zod deep-link/persistence validation; no arbitrary WebView core flow; clear permission strings; redacted production logs; no transport exceptions; Expo Updates only within controlled runtime/release policy.

## API controls

TLS, size/time/rate limits, strict validation, no arbitrary URL fetch, outbound allowlist, circuit breakers, parameterized SQL, safe errors, runtime secrets, private admin/metrics, dependency/image scans.

Reads may be anonymous. Installation/subscription mutations use an issued installation credential stored safely and verified server-side.

## Linux host controls

- Only Caddy publishes public ports.
- PostgreSQL, Docker API, metrics, and admin endpoints remain private.
- SSH uses keys, a dedicated deploy user, and restricted source addresses or VPN.
- Secret files are root-owned and mounted read-only.
- Containers drop capabilities and use `no-new-privileges` where compatible.
- API/poller/worker run as non-root with read-only filesystems and bounded temporary mounts.
- Docker logs rotate and disk/backup age are alerted.
- Host updates and reboot recovery are tested.

## CI controls

Branch protection/reviews, protected production environments, restricted SSH deploy identity, pinned known-host fingerprint, EAS credential management, image signing/provenance/SBOM, secret scan, pinned action SHAs where practical, and no secrets for untrusted PRs.

## Mapbox obligations

Review current terms; retain logo/attribution; expose telemetry control; document data attribution; restrict/rotate tokens; budgets/quotas; enforce Search/Geocoding storage rules; validate offline/cache use; disclose privacy.

## TriMet obligations

Register/review AppID terms, rates, cache/retention/attribution, beta behavior, GTFS/GIS terms, developer notices. AppID server-side.

## Direct-source obligations

Written API/source ownership, rate, license/redistribution, attribution, history rights, and contact. Do not imply partnership or scrape public map. Deduplicate overlapping sources. Streetcar and Rose City-inspired presentation supplied by TriMet follow the TriMet source boundary and D-001 review.

## Open source/store

Automated JS/Go/native/container license inventory and third-party notices; accurately call RNMapbox community-supported; review restricted/copyleft licenses. Prepare Apple privacy manifest/reason APIs, privacy labels, Google Data Safety, location/notification declarations, policy/support URLs, and re-review after dependency changes.

## Incidents

Runbook for credential rotation, token restriction, source/feature disable, installation invalidation, evidence, required notification, and postmortem.

## Acceptance

Threat model reviewed; no privileged secret in mobile; redaction tests; deletion purges push data; terms/attribution registered; telemetry setting reachable; store forms accurate; licenses pass; restricted SSH, non-public database, non-public Docker API, root-owned secret files, and least-privilege production.
