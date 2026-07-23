# GTFS-Realtime sanitized fixtures

The GTFS-Realtime parser and poller tests construct a minimal protobuf fixture
with synthetic identifiers (`2901`, `20`, and `trip-1`) using the official
generated message types. This keeps the binary wire fixture deterministic and
reviewable without retaining a provider payload. Tests also derive malformed,
deleted, stale, empty, incomplete-ID, and impossible-coordinate cases from it.

No real endpoint, credential, rider coordinate, or raw provider payload is
stored here.
