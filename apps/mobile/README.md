# Tabi mobile compatibility spike (WP-08)

This is an Expo development-build application foundation. Vehicle screens use
the normalized Tabi API only; they never call a transit provider directly.

The primary application icon is `assets/images/tabi-app-icon.png`. Expo uses
it for the shared app icon, iOS icon, Android adaptive icon foreground, and
web favicon. Re-run a native prebuild/development build to see icon changes on
an installed device.

## Vehicle vertical slice (WP-09)

The map tab defaults to deterministic fixture mode and renders the same
normalized vehicle data through a Mapbox `ShapeSource` (when a public Maps SDK
token is configured) and an accessible list/detail flow. It supports mode and
freshness filters, exact-ID-first vehicle search, source/age display, and an
explicit stale/source-unavailable state. Polling is foreground-only and remote
snapshot requests use ETags when supplied.

Set `EXPO_PUBLIC_API_MODE=remote` and a valid `EXPO_PUBLIC_API_BASE_URL` only
for a Tabi development API; invalid/missing remote configuration fails at the
mobile boundary. Fixture mode is the safe default and uses no network.

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
- RNTL/Vitest compatibility is not currently a pass claim. Native Mapbox
  rendering remains a Maestro/device responsibility regardless of the harness.
- The installed RNTL/Vitest dependency pair currently reaches React Native's
  untransformed Flow entrypoint under Vitest. Component `.tsx` tests are kept
  outside the passing pure-Vitest command until the harness ADR gate is
  resolved; repository, filter, freshness, and GeoJSON tests run in Vitest.
- `maestro/launch.yaml` is a bootstrap flow. It requires an installed development build and Maestro CLI.
- Physical iOS/Android development builds, Mapbox style/location-puck rendering, ShapeSource press/filter validation, New Architecture validation, actual SQLite migration, and Maestro launch are unrun evidence gates until devices and an approved token are available.

## Ownership and next step

WP-09 consumes the `src/maps/` adapter and replaces synthetic data with normalized API data after WP-07. Keep router files thin and preserve the equivalent accessible list/detail flow.

## Fixture notifications foundation (WP-12)

Notification settings are an accessible, local fixture surface for service-alert
and departure-reminder subscriptions. They validate scope, IANA quiet-hour time
zones, lead time, and expiry, and only use opaque IDs in notification links.
The push adapter is explicitly unavailable: it requests neither OS permission
nor an Expo push token, performs no network call, and does not deliver a real
notification. Device permission, token registration/rotation, push receipts,
and end-to-end delivery remain unverified device/backend gates. Mobile
background policy only allows deferrable static/cache maintenance and
subscription reconciliation—never location monitoring or continuous vehicle
polling.

## Trip planning boundary (WP-11)

The Plan tab is a provider-independent planning shell. It offers deterministic
stop endpoint pickers, constraints, an accessible timeline that mirrors
map-friendly leg geometry, safe opaque-ID planning links, and an explicit
location-denied path. Fixture mode remains deterministic. In remote mode, the
repository calls only Tabi's normalized `/v1/journeys/plan` endpoint for opaque
TriMet stop IDs; it never calls Mapbox or TriMet directly. Local map pins and
saved locations remain local-only until an explicit coordinate contract is
implemented. Planning links never include precise coordinates or search text,
and no endpoint or location data is persisted by this foundation.
