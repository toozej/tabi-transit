---
doc_id: TAB-PLAN-018
title: "Open Questions, Decision Gates and Risks"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["technical-lead", "product-owner", "legal-coordinator"]
depends_on: ["TAB-PLAN-001", "TAB-PLAN-002"]
---


# Open Questions and Decisions

Do not stall unrelated work. Apply the default while collecting evidence.

## External decisions

- **D-001 TriMet:** AppID, terms, rate, cache/retention/attribution, beta expectations. Default backend-only conservative official-source use.
- **D-002 Streetcar:** canonical feed/API, ownership, UmoIQ relationship, terms, overlap. Default disabled/no scraping.
- **D-003 Rose City Transit:** API/export, rights, rate, attribution/socials, history/advanced fields, change contact. Default optional reference only.
- **D-004 Mapbox:** Search Box vs Geocoding, storage, scopes, telemetry/attribution, offline/cache, budget. Default no unapproved persistent geocoder storage.

## Architecture decisions

- **D-010** Exact Expo/RNMapbox/native SDK matrix: Phase 0 physical builds; default current compatible stable/v11.
- **D-011** Vitest/RNTL harness: prove; no silent Jest suite; move native assertions to Maestro under ADR if needed.
- **D-012** Vehicle hot-path JSON vs GeoJSON: benchmark; freeze contract after.
- **D-013** Mobile static sync SQLite artifact vs JSON: profile; default JSON vertical slice.
- **D-014** Poller: separate task sharing packages/DB.
- **D-015** Push: Expo Push Service first, abstraction retained.
- **D-016** Linux hosting: select an affordable VPS/provider, supported distribution, server size, backup target, DNS, and SSH/VPN policy. Canonical runtime remains Docker Compose.
- **D-017** Analytics/crash: privacy/cost review; default minimal crash only.
- **D-019** Fly.io optional path: validate current trial/pricing, process-group sizing, database choice, and backup before use. Do not label it a free tier.
- **D-018** OTP: not deployed until a measured TriMet planner gap.

## Product decisions

- Minimum OS versions.
- Exact mode taxonomy and shuttle/replacement handling.
- Parent station versus platform semantics for nearest/per-mode grouping and default radius.
- Rider-facing adherence wording and specialist visibility.
- Notification MVP scope.
- Tabi/TriMet-inspired branding/trademark review.

## Risk register

| Risk | Impact | Mitigation |
|---|---|---|
| Expo/RNMapbox incompatibility | map blocked | Phase 0 device/version proof |
| Community wrapper maintenance | upgrade/security | adapter isolation, release monitoring, native fallback ADR |
| Upstream quota/schema | missing data | backend caches/adapters/fixtures/health/flags |
| No Streetcar API | scope gap | gate and pursue official source |
| Mapbox cost | financial | scopes, budgets, usage/cost metrics |
| Vitest RN instability | component gap | proof, pure logic separation, Maestro fallback ADR |
| GTFS↔RT ID mismatch | incorrect joins | mappings/diagnostics/unmatched preservation |
| stale shown as live | trust harm | freshness contract/UI/no aggressive interpolation |
| single-host outage | API downtime | off-site backup, scripted rebuild, external uptime, practiced restore |
| disk/log exhaustion | outage/corruption | rotated logs, quotas/headroom alarms, retention |
| store/privacy rejection | delay | early forms/minimal permissions/checklist |
| excessive infra scope | schedule/cost | single Linux Compose host; profiles; delay Redis/OTP/large dashboards |
| specialist over-scope | MVP delay | Phase 5 gates |
| location leakage | privacy | local history/no background/log redaction |

## Decision process

Owner writes ADR with context/options/evidence/decision/consequences/rollback; architecture/security/product approvals as relevant; update plans/contracts; close with tested evidence or reviewed terms.
