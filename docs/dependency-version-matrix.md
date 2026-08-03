# Dependency and version-selection matrix

The table records selection criteria, not unverified current versions. A version becomes **Selected** only when its exact value is pinned in the named artifact and validation is recorded.

| Component                            | Status   | Selection criteria                                            | Pinning artifact                   | Required validation                      |
| ------------------------------------ | -------- | ------------------------------------------------------------- | ---------------------------------- | ---------------------------------------- |
| Node.js                              | Proposed | Supported LTS; compatible with Expo and pnpm                  | `.nvmrc` or `package.json` engines | `make doctor`, workspace install         |
| pnpm                                 | Proposed | Compatible with selected Node/workspaces                      | `packageManager`                   | frozen-lockfile install                  |
| Expo SDK / React Native              | Proposed | Supported stable pairing for development builds               | mobile `package.json`, lockfile    | iOS/Android development build            |
| Xcode / iOS Simulator                | Proposed | Latest Xcode supported by Sequoia; Intel-compatible runtimes  | `.xcode-version`, simulator config | Both configured simulator targets        |
| Android Studio / Emulator            | Proposed | Stable Studio; Intel x86_64 API 31, 36, and 37 images         | Brewfile, emulator config          | All three configured emulator targets    |
| `@rnmapbox/maps` / native Mapbox SDK | Proposed | Compatible with selected Expo/RN; maintained security posture | mobile manifest/lockfile           | ShapeSource build/device proof           |
| TypeScript / Vitest / RNTL           | Proposed | Strict TS and a proven RN-compatible test harness             | workspace manifests/lockfile       | typecheck and representative test        |
| Go                                   | Proposed | Current supported stable compatible with generators           | `go.mod`, `go.work`                | unit/race tests                          |
| Chi, pgx, sqlc, OpenAPI generators   | Proposed | Supported Go version; reproducible output                     | `go.mod`, generator config         | generate-check and tests                 |
| PostgreSQL / PostGIS                 | Proposed | PostGIS support; compatible image/dump/restore path           | Compose image digest               | migrations, spatial query, restore proof |
| Docker Engine / Compose              | Proposed | Linux supported releases; Compose spec support                | runbook/version evidence           | `docker compose config`                  |
| Caddy                                | Proposed | Supported stable; security update policy                      | image digest                       | config validation and TLS rehearsal      |
| restic                               | Proposed | Encrypted repository support and restore verification         | deployment docs/image/package pin  | backup/restore rehearsal                 |

Selection rule: choose supported, non-EOL releases with no known blocking compatibility issue; pin exact package versions in lockfiles and exact production image digests. Upgrade only with compatibility, security, rollback, and native-rebuild evidence appropriate to the component.
