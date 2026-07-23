# ADR-0010: Defer Local Device and Maestro Evidence While Continuing Non-Device Work

- Status: Accepted
- Date: 2026-07-23
- Owners: Integration / Mobile / Quality

## Context

This workstation cannot install Android or iOS development builds or run Maestro. The mobile Mapbox public token is configured locally, but `.env` files are not read or committed by the implementation process.

## Decision

Proceed with non-device Phase 1 completion and subsequent dependency-ready backend, contract, infrastructure, and pure-TypeScript work. Do not inspect, print, modify, or commit local environment files. Treat Android/iOS development-build, native RNMapbox, physical SQLite, and Maestro checks as explicitly deferred evidence rather than successful checks.

## Consequences

The roadmap's local execution gate no longer blocks implementation progress on this workstation. It still blocks any claim of native-device acceptance, app-store readiness, or production release. A capable CI/device environment must run the documented flows before those claims are made.
