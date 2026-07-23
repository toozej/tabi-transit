# Architecture decision register

| ID               | Decision                                       | Status                 | Owner                         | Gate / evidence still required                                                          |
| ---------------- | ---------------------------------------------- | ---------------------- | ----------------------------- | --------------------------------------------------------------------------------------- |
| ADR-0001         | Contract-first modular monolith                | Accepted               | Architecture                  | OpenAPI lint/generation and conformance tests                                           |
| ADR-0002         | Expo development builds; adapter boundaries    | Accepted               | Architecture/Mobile           | RNMapbox device matrix; RNTL/Vitest proof                                               |
| ADR-0003         | Backend normalization and explicit freshness   | Accepted               | Architecture/Backend          | Source fixture and stale/recovery tests                                                 |
| ADR-0004         | Single-host Compose/Caddy canonical deployment | Accepted               | Architecture/Infra/Security   | Host, firewall, restore, and rollback evidence                                          |
| ADR-0005         | Official source first; optional-source gates   | Accepted               | Architecture/Product/Security | Terms, rights, credentials, and attribution review                                      |
| ADR-0006         | Evidence-based version selection/pinning       | Accepted               | Architecture/Repository       | Recorded versions, lockfiles, compatibility checks                                      |
| ADR-0007         | Fly.io optional image adapter                  | Accepted               | Architecture/Infra            | Current pricing, capacity, database, backup review                                      |
| ADR-0008         | Defer optional expansions                      | Accepted               | Architecture/Product/Security | Feature-specific ADR evidence                                                           |
| D-010            | Exact Expo/RNMapbox/native SDK matrix          | Deferred               | Mobile                        | Physical development-build evidence                                                     |
| ADR-0009 / D-011 | Vitest/RNTL compatibility harness              | Accepted evidence gate | Mobile/QA                     | RNTL/Vitest proof remains required; pure Vitest plus Maestro/device fallback documented |
| D-012            | Vehicle payload JSON vs GeoJSON                | Deferred               | API/Mobile                    | Synthetic fleet payload/render benchmark                                                |
| D-013            | SQLite artifact vs JSON sync                   | Deferred               | Mobile/Backend                | Size, migration, and offline profiling                                                  |
| D-016            | VPS/provider, DNS, SSH/VPN, backup target      | Deferred               | Operations/Security           | Operator selection and documented evidence                                              |
| D-017            | Analytics/crash reporting                      | Deferred               | Product/Privacy               | Privacy, cost, and retention review                                                     |
| D-018            | OpenTripPlanner                                | Deferred               | Product/Architecture          | Measured TriMet planner gap and host capacity                                           |
| D-019            | Fly.io deployment                              | Deferred               | Operations                    | Current cost and recovery approval                                                      |

No external account, credential, source terms, deployment host, or legal approval has been verified by this register.
