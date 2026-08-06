---
doc_id: TAB-PLAN-022
title: "Responsive Web Application Implementation Plan"
status: implementation-ready
last_updated: 2026-08-05
intended_agents: ["web-agent", "design-system-agent", "accessibility-agent", "qa-agent"]
depends_on: ["TAB-PLAN-001", "TAB-PLAN-002", "TAB-PLAN-003", "TAB-PLAN-004", "TAB-PLAN-005", "TAB-PLAN-012", "TAB-PLAN-017"]
---

# Responsive Web Application Implementation Plan

## Objective

Deliver Tabi as a responsive, accessible web application at a stable public
URL. It must provide the same rider-facing capabilities and truthful
freshness/offline behavior as the native apps on current desktop and mobile
browsers. It is a first-class client of the Tabi API; it never calls TriMet,
Mapbox Search, Geocoding, or other transit providers directly. It may use the
Mapbox browser Maps SDK solely to render the approved vehicle-map style.

The web app complements rather than replaces the Expo application. Native-only
capabilities (native map SDK behavior, OS notification registration, device
SQLite, and installed-app deep links) retain platform adapters. Feature parity
means equivalent rider outcomes and explanations, not identical widgets or a
claim that browser permissions/background execution behave like native apps.

## Chosen delivery shape

Add `apps/web/` as a TypeScript React single-page application built with Vite.
It will use a web router, TanStack Query, Zod, the generated `@tabi/api-client`,
and browser-native accessibility semantics. The production build is immutable
static assets served by the existing public edge/Caddy deployment; API requests
remain same-origin under `/v1` in production. This avoids privileged browser
credentials, simplifies CORS, and permits independent web releases without
changing native binaries.

Do not attempt to make the current React Native screens render on the web as
the primary implementation. `@rnmapbox/maps`, native navigation conventions,
and desktop information density would create an unreliable compromise. Instead,
share pure contracts and presentation rules, then build platform-appropriate
DOM and React Native views from the same feature specifications.

## Shared-client architecture

```text
                    packages/api-client + packages/transit-domain
                                      |
              shared query keys / validation / formatting / fixtures
                         |                              |
                 apps/mobile                     apps/web
          Expo Router + native UI          React router + semantic HTML
          RNMapbox + SQLite                Web-map adapter + IndexedDB
                         \                 /
                          \               /
                         Tabi /v1 API and public static artifacts
```

Before building parity screens, extract only platform-neutral code from
`apps/mobile/src` into workspace packages: domain models/policies, API request
schemas and query-key factories, freshness/formatting rules, deep-link payload
parsers, and deterministic fixtures. Keep rendering, navigation, map, storage,
permission, external-link, and notification code behind client-specific
adapters. Do not move React Native components into a shared package.

Create `packages/design-tokens/` for color, type scale, spacing, radius,
elevation, motion, icon, and semantic-state tokens. Generate CSS custom
properties for web and a typed native token map from one source. Tokens and a
small component-behavior specification are the source of the shared visual
feel; each platform owns its components.

## Web application structure

```text
apps/web/
├── src/
│   ├── app/{router,providers,queryClient,config,errorBoundary}/
│   ├── features/{nearby,vehicles,planner,riderInfo,saved,notifications}/
│   ├── components/{layout,feedback,forms,list,details}/
│   ├── maps/{WebMapAdapter,layers,geojson,controls}/
│   ├── platform/{storage,location,externalLinks,notifications}/
│   ├── styles/{tokens,global,utilities}/
│   └── test/
├── public/
├── index.html
├── vite.config.ts
└── package.json
```

Routes use durable, shareable, validated URLs:

| Rider surface | Web route | Equivalent native surface |
|---|---|---|
| Nearby | `/nearby` | Nearby tab |
| Vehicle map and search | `/map`, `/vehicles/:vehicleId` | Map tab/detail |
| Plan a trip | `/plan` | Plan tab |
| Alerts | `/alerts`, `/alerts/:alertId` | Alerts tab/detail |
| Saved items | `/saved` | Saved tab |
| Stop | `/stops/:stopId` | Stop detail |
| Route | `/routes/:routeId` | Route detail |
| Settings, credits, privacy | `/settings/*`, `/credits` | Settings/credits |

Path/query parameters are Zod-validated and only opaque IDs or explicit,
non-sensitive choices are serialized. Never place precise coordinates, search
text, push tokens, or provider session tokens in shareable URLs, analytics, or
logs.

## Responsive interaction model

At narrow widths, use a compact top bar and five primary destinations matching
the native tab order: Nearby, Map, Plan, Alerts, Saved. Primary content is
single-column; map detail/filter panels are modal dialogs or bottom sheets with
focus management.

At wide widths, show a persistent sidebar for the same destinations, use
two-column layouts for planner/stop/route detail where useful, and keep map
filters and selected details in an adjacent resizable panel. The URL always
identifies the selected entity, so refresh, browser history, a new tab, and a
shared link preserve context. Breakpoints are design tokens validated at 320,
375, 768, 1024, 1440, and 200%-zoom effective widths rather than assumptions
about a specific device.

All controls work with pointer, keyboard, and touch. Dialogs, sheets, popovers,
map focus, skip links, visible focus, Escape behavior, and focus restoration
are explicit acceptance criteria. Hover may enrich desktop UI but cannot expose
required information or actions.

## Functional parity scope

### Nearby, stops, routes, schedules, alerts, and saved data

Use the existing normalized endpoints and contracts for mode-filtered nearby
stops, stops/arrivals, route directions/shapes/stops, schedules, alerts, and
favorites/recents. Reproduce loading, empty, error, source-unavailable,
offline, and stale states with the same freshness wording and no fabricated
realtime state. Browser location is requested only after the rider chooses a
location-based action; typed stop ID, map-center, and saved-place flows work
after denial.

Persist favorites, recents, preferences, and approved replaceable static/cache
artifacts in IndexedDB through a repository interface. Session-only fallback is
explicit if storage is unavailable. Browser data clearing is granular (recents,
saved data, replaceable cache) and documented in settings. The first web MVP
does not claim full offline operation: a service worker/static-artifact strategy
is added only after cache versions, update/rollback behavior, quotas, and
provider terms are tested.

### Vehicles and maps

`apps/web/src/maps/VehicleMap.tsx` implements the `WebMapAdapter` with pinned
Mapbox GL JS (`mapbox-gl` 3.15.0). It receives the optional restricted public
browser token and approved style URL from the validated web configuration,
owns map lifecycle/source updates/cleanup, and preserves Mapbox attribution.
It renders no browser map without `VITE_MAPBOX_ACCESS_TOKEN`; on a missing
token or load failure, the complete accessible vehicle list/search/detail flow
remains available. Production enablement is tracked by D-020, independently
from the Search/Geocoding gate D-004.

Render the fleet as one source plus style layers (and a separate selected
feature), never as one DOM marker per vehicle. Keep textual search exact-ID
first, filters, freshness/source labels, selected detail/history, and a fully
equivalent accessible list. Stale observations are not animated or labelled
live. Map rendering and keyboard map interactions are enhancements; every map
task must be completable from the list/detail UI.

### Planning and external directions

Build the same endpoint picker, constraints, ranking disclosure, itinerary
timeline, alerts/freshness, and opaque-ID share link as native. Use browser
location only by rider action. A walking handoff opens a clearly labelled
external web-map URL in a new browsing context after the rider invokes it; do
not prefetch or disclose coordinates before that action. Search/geocoding stays
behind the existing Tabi backend and D-004 decision gate.

### Notifications and installability

Notification settings display existing subscriptions where the authenticated or
anonymous-installation contract supports them, but browser push is a separate
explicit opt-in and is not silently treated as an Expo token. It remains off
until D-017/D-018 and a browser-specific consent, service-worker, token,
delivery, and deletion design are approved. Native deep links and browser URLs
must both resolve to the relevant public route.

The initial release is a responsive web application, not an installable PWA
promise. Add a web app manifest/service worker only when the offline decision
above passes; installation must not change data-retention or notification
consent behavior.

## API, security, and operations work

1. Add web public-origin configuration, production same-origin proxying, local
   development CORS allowlists, and contract tests. Never use wildcard CORS
   with credentials.
2. Validate all API/deep-link/storage input with the shared schemas. Keep
   provider credentials and privileged Mapbox/Search tokens server-side; allow
   only the separately restricted public browser Maps SDK token in the client.
3. Define a restrictive CSP, `frame-ancestors`, `nosniff`, referrer policy,
   HTTPS/HSTS, asset cache headers, and an explicit third-party map/style
   allowlist. CSP/report-only rollout precedes enforcement.
4. Treat browser local storage as device-local and untrusted. Store no precise
   location, raw search text, access token, or notification secret by default.
5. Send privacy-safe web telemetry only after the product consent decision;
   include release/version, route class, and failure category, not sensitive
   route parameters. Add web RUM/error sampling, map/API error dashboards, and
   alert thresholds before broad rollout.
6. Publish immutable, content-hashed assets with a release identifier and a
   rollback procedure that can restore the previous web artifact without
   changing API contracts. API and web deployments use a compatibility window.

## Delivery phases and exit criteria

### W0 — decisions and foundation

- Configure the restricted browser map token/style and verify the
  supported-browser matrix, public origin, attribution, and usage controls
  required by D-020; decide analytics/consent, offline/PWA scope, and
  browser-push scope.
- Add `apps/web`, Vite, strict TypeScript, lint/test/build commands, config
  validation, static hosting preview, and no-secret fixture mode.
- Extract shared domain/API/query/token primitives without changing native
  behavior; record package ownership and migration tests.
- Establish responsive visual references for all five tabs and the key detail
  states.

Exit: fixture build deploys to a preview URL; routes, keyboard shell, tokens,
and generated-client boundary pass CI; browser-map capability has a documented
enabled or accessible-list fallback decision.

### W1 — usable responsive shell and rider information

- Implement app shell, navigation, errors, freshness/state components,
  preferences, saved/recents repository, Nearby, Stop, Route, Schedules, and
  Alerts.
- Add desktop/sidebar and narrow/mobile layouts, URL state, location-denied
  flow, cache-clearing controls, credits, privacy, and attribution.

Exit: the web app can complete the Phase 2 rider-information journeys with
fixture and remote-contract data on desktop and mobile browsers; all map-only
information has a text/list path.

### W2 — live map and planning parity

- Implement vehicle list/search/detail/history, feature-layer map adapter,
  filters, reduced-motion behavior, and map fallback.
- Implement planning picker, constraints, timelines, deep links, and external
  walking handoff.

Exit: vehicle and planner acceptance flows match the native product outcomes,
including stale/source-unavailable and location-denied states. Fleet rendering
meets the agreed browser performance budget on the supported matrix.

### W3 — release hardening

- Complete automated accessibility, visual/responsive, browser E2E, security
  headers/CSP, monitoring, deploy/rollback, and failure-injection coverage.
- Perform manual screen-reader, keyboard, touch, slow-network, storage-denied,
  offline/degraded, and map-attribution checks. Resolve legal/provider gates.

Exit: release gate below passes and the web release is documented alongside the
native release, without claiming unimplemented PWA or browser-push features.

## Quality and release gate

CI runs typecheck, lint, unit/component tests, generated-client drift checks,
production build, and preview smoke tests. Playwright covers Chromium, Firefox,
and WebKit for the critical flows; mobile-browser runs cover iOS Safari and
Chrome for Android on real devices or a documented equivalent. Automated axe
checks complement, never replace, manual VoiceOver/NVDA/JAWS and keyboard
testing. Map behavior is tested through adapter/source builders plus browser
E2E/manual checks, not DOM-marker assumptions.

The web release is blocked until the five primary destinations and detail/deep
links work at supported breakpoints; keyboard and screen-reader paths work;
freshness/offline/error states remain truthful; source/map attribution is
visible; no secret or sensitive URL/storage/logging path exists; CSP and
headers are verified; and deployment rollback is rehearsed. Native device and
browser evidence are tracked independently.

## Work-package sequence

| Package | Scope | Depends on |
|---|---|---|
| WP-W0 | decisions, web bootstrap, tokens, shared-core extraction | D-019/D-020, existing API client |
| WP-W1 | app shell, responsive navigation, shared states, storage | WP-W0 |
| WP-W2 | nearby/stops/routes/schedules/alerts/saved | WP-W1, rider-information endpoints |
| WP-W3 | vehicle map/list/detail/history | WP-W1; map production enablement requires D-020 |
| WP-W4 | planner/search/external browser links | WP-W1, planner/search gates |
| WP-W5 | accessibility, E2E, security headers, deploy/rollback | WP-W2–WP-W4 |

Do not begin a browser-push/PWA work package until the separate approval gate
is resolved. Any divergence from the shared product behavior requires an ADR
or an explicit platform limitation in the applicable feature specification.
