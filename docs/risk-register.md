# Risk register

| ID    | Risk                                      | Impact                            | Mitigation / trigger                                                | Owner            | Status |
| ----- | ----------------------------------------- | --------------------------------- | ------------------------------------------------------------------- | ---------------- | ------ |
| R-001 | Expo/RNMapbox incompatibility             | Blocks map release                | Adapter isolation; Phase 0 build/device proof                       | Mobile           | Open   |
| R-002 | Vitest/RNTL harness instability           | Weak mobile component coverage    | Prove harness; use pure Vitest plus Maestro fallback                | Mobile/QA        | Open   |
| R-003 | Upstream quota/schema/availability change | Missing or incorrect transit data | Backend adapters, fixtures, source health, flags                    | Backend          | Open   |
| R-004 | GTFS-to-realtime ID mismatch              | Incorrect joins                   | Diagnostics; preserve unmatched records; mapping tests              | Backend          | Open   |
| R-005 | Stale data presented as live              | Rider trust/safety harm           | Mandatory freshness contract and stale UI tests                     | Backend/Mobile   | Open   |
| R-006 | Location or search-text leakage           | Privacy incident                  | On-device defaults; redacted logs; no background location           | Security         | Open   |
| R-007 | Mapbox cost or terms violation            | Financial/legal exposure          | Token restriction, budget, approved storage policy                  | Product/Security | Open   |
| R-008 | Single-host outage/disk exhaustion        | Service outage/data loss          | Off-site backups, rotated logs, headroom alerts, restore rehearsal  | Operations       | Open   |
| R-009 | Supply-chain or CI secret compromise      | Credential/code compromise        | Pinning, scans, protected environments, no secrets in untrusted PRs | Security         | Open   |
| R-011 | Fly cost/recovery assumptions             | Unplanned spend/outage            | Optional only; current pricing and recovery review                  | Operations       | Open   |

Risk review is required before enabling an external provider, production deployment, notifications, or persistent geocoder storage.
