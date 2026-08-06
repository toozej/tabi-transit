# Current Blockers and Evidence Gates

Last updated: 2026-08-05

This register distinguishes work that is prohibited until evidence exists from
work that can continue using deterministic fixtures and disabled adapters.
None of the items below is inferred to be approved, configured, or available.

The validated vehicle `bbox` API filter, route-stop direction selector,
external walking-directions handoff, and request-ID hardening added on
2026-07-30 introduce no new evidence gate. The documented vehicle-direction
filter remains a non-blocking implementation follow-up because the normalized
GTFS-Realtime current-vehicle record has no direction value to filter yet; it
must receive reviewed static-trip enrichment rather than silently ignoring the
parameter.

| ID             | Gate                                             | Affected work                                         | Current handling                                                                                                                                                                                                                              | Required evidence to clear                                                                                                                                                                              |
| -------------- | ------------------------------------------------ | ----------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| D-004          | Mapbox Search/Geocoding terms and budget         | Address/POI search and persistent geocoder results    | No provider search calls or persistent geocoder storage. Native and browser map rendering do not use Mapbox Search or Geocoding.                                                                                                              | Product/legal review of Search/Geocoding product, token scopes, attribution, telemetry, storage/cache rules, and a usage budget.                                                                        |
| D-020          | Browser Mapbox public-token and release evidence | Enabling the web vehicle map outside fixtures         | `apps/web` has a Mapbox GL JS vehicle-map adapter and falls back to the accessible list when `VITE_MAPBOX_ACCESS_TOKEN` is unset or the map fails to load. No browser token, permitted public origin, or browser-map verification is assumed. | A minimum-scope, URL-restricted public browser Maps SDK token; approved style and attribution/telemetry treatment; Mapbox usage budget/alerts; and supported-browser, keyboard, and attribution checks. |
| D-010          | Physical Expo/RNMapbox compatibility             | Production map release                                | Local Android/iOS development-build evidence is waived by the user for this workstation, but remains unverified.                                                                                                                              | Successful development builds and physical-device map checks on supported Android and iOS versions.                                                                                                     |
| D-011          | RNTL with Vitest compatibility                   | Native component-test evidence                        | Pure TypeScript logic is tested with Vitest; native harness remains unproven because RNTL's CommonJS React Native require reaches the pinned RN Flow entrypoint.                                                                              | A proven compatible RNTL/Vitest transform setup, or an approved alternative test strategy with its dependencies available and approved.                                                                 |
| D-013          | Native SQLite offline-artifact evidence          | Future SQLite static schedules/cache                  | ADR-0012 accepts validated JSON artifacts as the default. Native SQLite is not selected or claimed as device-validated.                                                                                                                       | Supported iOS/Android indexed SQLite lookup, startup, update, and migration measurements.                                                                                                               |
| D-016          | Linux host and operations policy                 | Production Compose deployment                         | Compose, scripts, and topology are validated locally; no host, DNS, backup target, or credentials are assumed.                                                                                                                                | Selected host/provider, DNS, SSH/VPN policy, off-site backup target, and a practiced restore.                                                                                                           |
| D-017          | Expo push project and credential evidence        | Live push delivery, receipts, and token rotation      | Phase 4 uses deterministic fixtures and a disabled push gateway only. No Expo project ID, push credential, device token, or push request is assumed.                                                                                          | An approved Expo project configuration, platform credential ownership/rotation policy, a real device token, and an isolated delivery/receipt test.                                                      |
| D-018          | Notification retention and user-facing policy    | Subscription limits, delivery retention, consent copy | Schema and UI may enforce conservative defaults, but no retention duration, notification copy, or jurisdiction-specific consent claim is treated as approved.                                                                                 | Product/privacy approval of retention duration, subscription limits, consent language, and deletion/export expectations.                                                                                |
| Device/Maestro | Local device and Maestro execution               | Mobile release evidence                               | Explicitly skipped on this workstation at the user's direction; not marked passed.                                                                                                                                                            | Run the documented device and Maestro flows in a capable environment.                                                                                                                                   |

## Resolved implementation gates

- **D-001 TriMet production enablement (2026-07-28):** Cleared by the product
  owner. Official TriMet access, the configured AppID, attribution, and the
  30-day normalized retention policy in ADR-0013 are approved for Tabi's
  backend adapters. This authorizes implementation and production composition;
  it does not claim that every planned TriMet-facing API or mobile screen is
  already implemented.

- **D-002 TriMet-provided Streetcar coverage (2026-07-26):** Cleared as a
  separate-source gate. TriMet's official Arrivals V2 documentation states
  that Streetcar results are included by default, and TriMet publishes the
  GTFS schedule and GTFS-Realtime endpoints. Streetcar data is therefore
  handled only as a `streetcar` mode within the TriMet source boundary; Tabi
  does not call Portland Streetcar, PBOT, or UmoIQ directly, and does not
  scrape them. D-001 is also resolved; normal runtime configuration still
  controls whether the provider adapter is composed.

- **D-003 Rose City Transit representation (2026-07-26):** Cleared as a
  separate-source gate. Rose City Transit is treated as a presentation/reference
  for data supplied through the official TriMet interfaces; Tabi does not call,
  scrape, store data from, or imply a partnership with Rose City. All machine
  data remains within the approved TriMet source boundary.

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
  fields. The initial credentialed smoke attempt could not open the opaque
  secret file in this workspace. On 2026-07-24, the repeatable local smoke test
  (`make test-trimet-live`) reached TriMet and revealed numeric `locid`/`route`
  fields; the adapter now normalizes those opaque IDs and the subsequent
  read-only smoke test passed. No AppID or provider response body was printed.

## Phase 3 consequence

Phase 3 may proceed with provider-independent contracts, deterministic
fixtures, and mobile UI that does not call a provider directly. The mobile
planner may call Tabi's normalized journey endpoint for opaque stop IDs; it
does not call TriMet or Mapbox. Live address/POI search remains
feature-disabled until D-004 is cleared. The web vehicle map can be enabled
independently once D-020 is cleared; its public token is intentionally optional
so fixture and preview builds do not carry a provider credential. The TriMet
planner is composed only
when validated server-side configuration is enabled; otherwise its public
endpoint fails closed.

## Phase 4 consequence

Phase 4 may add validated installation/subscription contracts, fixture-only
mobile settings, a transactional delivery model, and a disabled push gateway.
It must not request notification permission on launch, send a push, register a
real token, or claim delivery/receipt behavior until D-017 is cleared. Policy
defaults remain subject to D-018. The local worker persistence and receipt
state machine remain deliberately inert without an approved provider gateway.

## Phase 5 consequence

The TriMet-backed Streetcar and Rose City-inspired presentation slices may be
implemented and tested using the existing TriMet GTFS/GTFS-Realtime and
Arrivals boundaries. They must retain TriMet source/freshness semantics and
must not introduce a direct Streetcar, PBOT, UmoIQ, or Rose City dependency.
ADR-0013 accepts 30-day normalized vehicle-history retention. ADR-0015 now
provides the bounded history API and textual rider timeline. ADR-0016 retains
historical trip-update evidence and classifies it only when sufficient; no
rider-visible adherence claim is made from vehicle GPS, and no public aggregate
is exposed until that contract is separately reviewed.
