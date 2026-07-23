# Architecture decision register

| ID               | Decision                                       | Status                   | Owner                          | Gate / evidence still required                                                          |
| ---------------- | ---------------------------------------------- | ------------------------ | ------------------------------ | --------------------------------------------------------------------------------------- |
| ADR-0001         | Contract-first modular monolith                | Accepted                 | Architecture                   | OpenAPI lint/generation and conformance tests                                           |
| ADR-0002         | Expo development builds; adapter boundaries    | Accepted                 | Architecture/Mobile            | RNMapbox device matrix; RNTL/Vitest proof                                               |
| ADR-0003         | Backend normalization and explicit freshness   | Accepted                 | Architecture/Backend           | Source fixture and stale/recovery tests                                                 |
| ADR-0004         | Single-host Compose/Caddy canonical deployment | Accepted                 | Architecture/Infra/Security    | Host, firewall, restore, and rollback evidence                                          |
| ADR-0005         | Official source first; optional-source gates   | Accepted                 | Architecture/Product/Security  | Terms, rights, credentials, and attribution review                                      |
| ADR-0006         | Evidence-based version selection/pinning       | Accepted                 | Architecture/Repository        | Recorded versions, lockfiles, compatibility checks                                      |
| ADR-0007         | Fly.io optional image adapter                  | Accepted                 | Architecture/Infra             | Current pricing, capacity, database, backup review                                      |
| ADR-0008         | Defer optional expansions                      | Accepted                 | Architecture/Product/Security  | Feature-specific ADR evidence                                                           |
| D-010            | Exact Expo/RNMapbox/native SDK matrix          | Deferred                 | Mobile                         | Physical development-build evidence                                                     |
| ADR-0009 / D-011 | Vitest/RNTL compatibility harness              | Accepted evidence gate   | Mobile/QA                      | RNTL/Vitest proof remains required; pure Vitest plus Maestro/device fallback documented |
| ADR-0010         | Defer workstation device evidence              | Accepted                 | Integration / Mobile / Quality | Native/device acceptance remains unverified and release-gated                           |
| D-012            | Vehicle payload JSON vs GeoJSON                | Deferred                 | API/Mobile                     | Synthetic fleet payload/render benchmark                                                |
| ADR-0012 / D-013 | JSON static artifact pending SQLite evidence   | Accepted bounded default | Mobile/Backend                 | Native indexed SQLite, startup, update, and migration profiling                         |
| D-016            | VPS/provider, DNS, SSH/VPN, backup target      | Deferred                 | Operations/Security            | Operator selection and documented evidence                                              |
| D-017            | Expo Push project/credential evidence          | Deferred                 | Product/Mobile/Security        | Project, credential ownership, real token, isolated delivery/receipt test               |
| D-018            | Notification retention and user-facing policy  | Deferred                 | Product/Privacy                | Retention, consent copy, subscription limits, deletion/export policy                    |
| D-019            | Fly.io deployment                              | Deferred                 | Operations                     | Current cost and recovery approval                                                      |

No external account, credential, source terms, deployment host, or legal approval has been verified by this register.
