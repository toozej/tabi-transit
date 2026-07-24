# Current Blockers and Evidence Gates

Last updated: 2026-07-23

This register distinguishes work that is prohibited until evidence exists from
work that can continue using deterministic fixtures and disabled adapters.
None of the items below is inferred to be approved, configured, or available.

| ID             | Gate                                            | Affected work                                         | Current handling                                                                                                                                                                                                                                                                                                          | Required evidence to clear                                                                                                                                          |
| -------------- | ----------------------------------------------- | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| D-001          | TriMet Web Services terms and live verification | Production TriMet arrivals and A-to-B planning        | An AppID is reported in uninspected local secret files. The adapter now supports the repository `TABI_TRIMET_*` names and the documented Arrivals V2 host/schema, but remains feature-disabled. The 2026-07-23 opaque-secret smoke attempt could not open the secret file in this workspace, before any provider request. | A successful read-only Arrivals V2 smoke test plus reviewed terms covering rate limits, caching/retention, attribution, beta expectations, and planner constraints. |
| D-002          | Portland Streetcar source rights                | Optional Streetcar integration                        | Disabled; no scraping or unofficial fallback.                                                                                                                                                                                                                                                                             | Written canonical-source and usage approval.                                                                                                                        |
| D-003          | Rose City Transit source rights                 | Optional Rose City integration                        | Disabled; no scraping or unofficial fallback.                                                                                                                                                                                                                                                                             | Written API/export, rights, rate, attribution, and retention approval.                                                                                              |
| D-004          | Mapbox Search/Geocoding terms and budget        | Address/POI search and persistent geocoder results    | No provider search calls or persistent geocoder storage. Map rendering remains separately configured through the public mobile token.                                                                                                                                                                                     | Product/legal review of product, token scopes, attribution, telemetry, storage/cache rules, and a usage budget.                                                     |
| D-010          | Physical Expo/RNMapbox compatibility            | Production map release                                | Local Android/iOS development-build evidence is waived by the user for this workstation, but remains unverified.                                                                                                                                                                                                          | Successful development builds and physical-device map checks on supported Android and iOS versions.                                                                 |
| D-011          | RNTL with Vitest compatibility                  | Native component-test evidence                        | Pure TypeScript logic is tested with Vitest; native harness remains unproven.                                                                                                                                                                                                                                             | A proven compatible RNTL/Vitest transform setup, or an approved alternative test strategy.                                                                          |
| D-013          | Native SQLite offline-artifact evidence         | Future SQLite static schedules/cache                  | ADR-0012 accepts validated JSON artifacts as the default. Native SQLite is not selected or claimed as device-validated.                                                                                                                                                                                                   | Supported iOS/Android indexed SQLite lookup, startup, update, and migration measurements.                                                                           |
| D-016          | Linux host and operations policy                | Production Compose deployment                         | Compose, scripts, and topology are validated locally; no host, DNS, backup target, or credentials are assumed.                                                                                                                                                                                                            | Selected host/provider, DNS, SSH/VPN policy, off-site backup target, and a practiced restore.                                                                       |
| D-017          | Expo push project and credential evidence       | Live push delivery, receipts, and token rotation      | Phase 4 uses deterministic fixtures and a disabled push gateway only. No Expo project ID, push credential, device token, or push request is assumed.                                                                                                                                                                      | An approved Expo project configuration, platform credential ownership/rotation policy, a real device token, and an isolated delivery/receipt test.                  |
| D-018          | Notification retention and user-facing policy   | Subscription limits, delivery retention, consent copy | Schema and UI may enforce conservative defaults, but no retention duration, notification copy, or jurisdiction-specific consent claim is treated as approved.                                                                                                                                                             | Product/privacy approval of retention duration, subscription limits, consent language, and deletion/export expectations.                                            |
| Device/Maestro | Local device and Maestro execution              | Mobile release evidence                               | Explicitly skipped on this workstation at the user's direction; not marked passed.                                                                                                                                                                                                                                        | Run the documented device and Maestro flows in a capable environment.                                                                                               |

## Resolved implementation gates

- **Arrival timezone/current-state foundation (2026-07-23):** GTFS agency
  timezones, calendar-aware after-midnight arrival derivation, transactional
  vehicle/trip-update current-state writers, and disabled-by-default poller
  composition are implemented and fixture-tested. The disposable PostGIS
  migration suite applied migration `000007` twice successfully. This does not
  verify a real source endpoint; D-001 and source configuration/approval rules
  still apply.

- **TriMet Arrivals V2 boundary correction (2026-07-23):** The adapter now
  defaults to the documented `https://developer.trimet.org` host, explicitly
  requests JSON, enforces TriMet's 60-minute maximum, preserves documented
  epoch-millisecond query/arrival times, and tolerates additive response
  fields. Deterministic fixture tests pass. The credentialed smoke test did
  not run because this workspace denied the adapter access to the opaque secret
  file; no AppID or response body was printed and no provider request occurred.

## Phase 3 consequence

Phase 3 may proceed only with provider-independent contracts, deterministic
fixtures, disabled adapter boundaries, and mobile UI that does not call a
provider directly. Live address/POI search and live A-to-B planning remain
feature-disabled until D-004 and D-001 respectively are cleared.

## Phase 4 consequence

Phase 4 may add validated installation/subscription contracts, fixture-only
mobile settings, a transactional delivery model, and a disabled push gateway.
It must not request notification permission on launch, send a push, register a
real token, or claim delivery/receipt behavior until D-017 is cleared. Policy
defaults remain subject to D-018. The local worker persistence and receipt
state machine remain deliberately inert without an approved provider gateway.

## Phase 5 consequence

WP-18 and the Phase 5 Streetcar/specialist-view scope are blocked by D-002 and
D-003. No Streetcar or Rose City adapter, source fixture presented as provider
data, deduplication policy, history/adherence screen, or public credit claim
may be implemented or enabled until the applicable written source approval is
recorded. The only allowed preparatory work is evidence collection and a
source-contract review; it must not contact, fetch from, or scrape a provider.
