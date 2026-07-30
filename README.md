# Tabi

Tabi is a Portland-area transit application for iOS and Android. It uses an Expo React Native client and Go services backed by PostgreSQL/PostGIS.

The implementation source of truth is in [docs/implementation-plan](docs/implementation-plan/00_README.md). Current delivery state is tracked in [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md).

## Development status

Phase 0 is in progress. The repository foundation, API contract, mobile compatibility, transit-data, and deployment spikes are being established before production feature work begins.

## Safety and source policy

Tabi uses official sources through backend adapters. Portland Streetcar coverage and Rose City Transit-inspired presentation use TriMet's official feeds and Arrivals V2 interface; Tabi has no direct Streetcar/PBOT/UmoIQ or Rose City integration. Never add provider credentials, production tokens, or `.env` files to Git.
