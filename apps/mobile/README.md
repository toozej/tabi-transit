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

## Intel Mac iOS simulators

The repository pins Xcode 26.3 in the root `.xcode-version`. This is the newest
Xcode release supported by macOS Sequoia. On the Intel MacBook Pro, install macOS Sequoia
15.6 or newer and select Xcode 26.3 before creating the targets:

```sh
sudo xcode-select --switch /Applications/Xcode.app/Contents/Developer
make ios-simulators
```

Every iOS simulator and device launch target first installs the locked mobile
workspace dependencies and then runs `make prereqs-ios-simulators`. On macOS,
that target runs Xcode's first-launch component installation and downloads the
universal iOS 26.2 Simulator runtime when an equal-or-newer patch is not
already installed. `make prereqs` additionally installs the repository's
pinned general-purpose tools and Android prerequisites; `make bootstrap`
remains as a compatibility alias for existing local workflows.

The simulator manifest is `ios-simulators.json`. Setup creates an iPhone 13
mini and an iPhone Air using the highest installed iOS 26.x runtime. Names
include the selected runtime version,
so installing a newer patch and rerunning setup creates a new target without
deleting the old one.

If automatic installation fails, install the newest version offered under
Xcode > Settings > Components > Add Platforms. For the current
Sequoia-compatible toolchain, the command-line equivalents for Intel are:

```sh
xcodebuild -downloadPlatform iOS -buildVersion 26.2 -architectureVariant universal
```

Build and launch the Expo development app on either exact simulator target:

```sh
make ios-iphone-13-mini
make ios-iphone-air
```

These commands run Expo prebuild as needed. The generated `ios/` directory
remains ignored and must not be committed. Simulator results are still native
device evidence and must be recorded separately in the accessibility matrix.

Compatibility and component-install commands follow Apple's
[Xcode system requirements](https://developer.apple.com/xcode/system-requirements/)
and
[additional component installation](https://developer.apple.com/documentation/xcode/downloading-and-installing-additional-xcode-components)
documentation.

## Intel Mac Android emulators

The Android emulator setup is also configured for the Intel macOS Sequoia
host. Homebrew must already be installed. Every Android emulator and device
launch target first installs the locked mobile workspace dependencies and runs
the Android prerequisite target, which uses `Brewfile.android` to install the
current stable Android Studio application, Android SDK Command-Line Tools, and
OpenJDK 21; prompts for Android SDK license acceptance; and installs or updates
the emulator, platform/build tools, and x86_64 Google APIs system images for
API 31 and 36:

The Android prerequisite target is also included in `make prereqs` on macOS
and is skipped on other operating systems. SDK packages are managed under `ANDROID_SDK_ROOT` or
`ANDROID_HOME` when either is set, otherwise under the standard
`~/Library/Android/sdk` location. The SDK manager refreshes installed package
revisions, so each API-level image receives the latest patch available for that
package.

Create the configured Android Virtual Devices (its prerequisites run automatically):

```sh
make android-simulators
```

Build, boot, and launch the Expo development app on a specific device with:

```sh
make android # Default: Motorola Razr 2024 / Android 16
make android-motorola-razr-2024
make android-pixel-10-pro
make android-sony-xperia-1-ii
```

Forcefully stop all running emulators when a simulator is hung or no longer
needed:

```sh
make stop-ios-simulators
make stop-android-simulators
make stop-simulators # Stop both platforms
```

The manifest in `android-emulators.json` models these targets:

| Target              | System image        | Display profile                       |
| ------------------- | ------------------- | ------------------------------------- |
| Motorola Razr 2024  | Android 16 / API 36 | 1080 × 2640 at 413 dpi, foldable base |
| Google Pixel 10 Pro | Android 16 / API 36 | 1280 × 2856 at 495 dpi                |
| Sony Xperia 1 II    | Android 12 / API 31 | 1644 × 3840 at 643 dpi                |

These are AVD compatibility profiles, not OEM firmware emulators. Google APIs
images do not reproduce Motorola Hello UX, Sony software, physical camera and
radio behavior, or the Razr cover display. The Razr target uses Android
Studio's foldable base profile when available and otherwise uses a Pixel Pro
base with the Razr main-display dimensions. Validate OEM behavior on physical
devices before release.

The setup follows Android's documentation for
[Android Studio installation](https://developer.android.com/studio/install),
[SDK Manager](https://developer.android.com/tools/sdkmanager),
[AVD Manager](https://developer.android.com/tools/avdmanager), and
[emulator acceleration](https://developer.android.com/studio/run/emulator-acceleration).

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
