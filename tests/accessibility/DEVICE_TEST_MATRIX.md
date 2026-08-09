# Accessibility and device test matrix

Automated checks establish only source-level semantics. Native Mapbox,
SQLite, VoiceOver, and TalkBack behavior require a development build on a real
device or supported simulator/emulator.

| Area                              | Automated evidence                                                                 | Manual/device gate                                             | Status                        |
| --------------------------------- | ---------------------------------------------------------------------------------- | -------------------------------------------------------------- | ----------------------------- |
| Map-independent vehicle selection | RNTL semantic query proof; policy check requires list/semantics when Mapbox exists | VoiceOver/TalkBack selection and details                       | Pending device                |
| Map filters/search/freshness      | Future component tests and deterministic fixture mode                              | iOS and Android development builds, max text and reduce motion | Not implemented               |
| Native Mapbox layers              | GeoJSON/layer builder unit tests                                                   | ShapeSource rendering, press/filter behavior, attribution      | Pending approved token/device |
| SQLite migrations                 | Executor unit proof                                                                | Actual Expo SQLite migration and recovery                      | Pending device                |
| Maestro critical flow             | Bootstrap YAML when installed                                                      | Launch and map/list equivalence on both platforms              | Pending CLI and build         |

Before release, test VoiceOver and TalkBack, denied location, max text size,
reduce motion, contrast/color independence, keyboard/switch access where
available, and completion of every core map flow through an equivalent list and
detail interface. Record device/OS/build identifiers and findings in the
release evidence; do not mark a gate complete from unit tests alone.

## Configured iOS simulator coverage

Run `make ios-simulators` on the Intel macOS Sequoia/Xcode host to create these
targets. Runtime selection is intentionally major-version based so setup always
uses the highest installed patch in that release line.

| Target         | Runtime selection          | Launch command            | Evidence focus                                |
| -------------- | -------------------------- | ------------------------- | --------------------------------------------- |
| iPhone 13 mini | Highest installed iOS 26.x | `make ios-iphone-13-mini` | Compact viewport, max text, VoiceOver         |
| iPhone Air     | Highest installed iOS 26.x | `make ios-iphone-air`     | Current UI behavior, VoiceOver, reduce motion |

Simulator success does not replace physical-device testing of performance,
notifications, location, or platform-specific hardware behavior.

Run `make android-simulators` on the Intel macOS host to create the Android AVD
coverage. These use Google APIs images and custom dimensions; Motorola and Sony
OEM behavior remains a physical-device gate.

| Target              | Runtime             | Launch command                    | Evidence focus                               |
| ------------------- | ------------------- | --------------------------------- | -------------------------------------------- |
| Motorola Razr 2024  | Android 16 / API 36 | `make android-motorola-razr-2024` | Foldable layout, compact posture, TalkBack   |
| Google Pixel 10 Pro | Android 16 / API 36 | `make android-pixel-10-pro`       | Current platform UI, TalkBack, reduce motion |
| Sony Xperia 1 II    | Android 12 / API 31 | `make android-sony-xperia-1-ii`   | Tall viewport, minimum supported behavior    |
