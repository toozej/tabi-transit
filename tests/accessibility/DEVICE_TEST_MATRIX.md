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
