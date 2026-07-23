# Tabi mobile compatibility spike (WP-08)

This is an Expo development-build application foundation. It deliberately has no backend calls or provider access.

## Local commands

After the workspace dependency install is available:

```sh
corepack pnpm --dir apps/mobile typecheck
corepack pnpm --dir apps/mobile test
corepack pnpm --dir apps/mobile start -- --dev-client
corepack pnpm --dir apps/mobile prebuild
```

Copy `.env.example` to a local ignored environment file only when a restricted public Maps SDK token is approved. Without it, the map route displays an accessible synthetic vehicle list and explicitly says that map rendering is disabled. Never use a server token, SDK-download token, provider credential, or signing material in this application.

## Phase 0 evidence and gates

- The map adapter uses one `ShapeSource` with 1,500 deterministic synthetic vehicle points and native `SymbolLayer`/`CircleLayer` layers; it does not use fleet `MarkerView`s.
- The SQLite migration unit proof verifies the transaction and version guard through the Expo SQLite executor interface. It is not a physical-device SQLite proof.
- The RNTL/Vitest test proves basic RNTL query wiring using a deliberately small React Native host-component test shim. It is not evidence that native Mapbox rendering works in Vitest; native behavior remains a Maestro/device responsibility.
- `maestro/launch.yaml` is a bootstrap flow. It requires an installed development build and Maestro CLI.
- Physical iOS/Android development builds, Mapbox style/location-puck rendering, ShapeSource press/filter validation, New Architecture validation, actual SQLite migration, and Maestro launch are unrun evidence gates until devices and an approved token are available.

## Ownership and next step

WP-09 consumes the `src/maps/` adapter and replaces synthetic data with normalized API data after WP-07. Keep router files thin and preserve the equivalent accessible list/detail flow.
