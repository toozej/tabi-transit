---
doc_id: TAB-PLAN-004
title: "Mobile Application Implementation Plan"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["mobile-agent", "mobile-accessibility-agent", "qa-agent"]
depends_on: ["TAB-PLAN-001", "TAB-PLAN-002", "TAB-PLAN-003", "TAB-PLAN-017"]
---


# Mobile Application Plan

## Objective and stack

One production iOS/Android application using Expo development builds, TypeScript, Expo Router, TanStack Query, Zustand, Zod, React Hook Form, Expo SQLite, SecureStore, Location, Notifications, Background Task, `@rnmapbox/maps`, Vitest, React Native Testing Library, and Maestro.

## M0 compatibility spike

Before feature screens:

1. Create Expo app and EAS development profiles.
2. Configure RNMapbox plugin and tokens.
3. Build physical iOS/Android apps.
4. Render style/location puck.
5. Update a `ShapeSource` with at least 1,500 synthetic vehicle points.
6. Validate symbol press/filter expressions.
7. Confirm New Architecture compatibility.
8. Prove SQLite migrations.
9. Prove Vitest pure-domain tests.
10. Prove RNTL under the chosen Vitest harness.
11. Prove Maestro launches both builds.
12. Pin versions and record caveats in ADR.

Exit: reproducible builds and spike tests pass on both platforms.

## Structure

```text
apps/mobile/
├── app/
│   ├── (tabs)/{nearby,map,plan,alerts,saved}.tsx
│   ├── stop/[stopId].tsx
│   ├── route/[routeId].tsx
│   ├── vehicle/[vehicleId].tsx
│   ├── itinerary/[itineraryId].tsx
│   ├── schedules/
│   ├── settings/
│   └── credits.tsx
├── src/{bootstrap,domain,data,features,maps,platform,ui,test}/
├── assets/
├── app.config.ts
└── eas.json
```

## Boot

Initialize privacy-safe error reporting → load config/flags → open/migrate SQLite → load local preferences/favorites → create QueryClient → configure Mapbox public token → register notification handlers without requesting permission → render shell → background refresh. Database failure gets a recovery screen and preserves durable user data when possible.

## Tabs and deep links

Tabs: Nearby, Map, Plan, Alerts, Saved. Settings include privacy, map telemetry, location status, cache/history deletion, data source freshness, credits/licenses. Deep links for stop/route/vehicle/alert/itinerary/subscription validate through Zod.

## State/data rules

### TanStack Query

Central typed keys, screen/app-aware polling, classified retries, cancellation, visible stale cache, and terms-aware persistence exclusions.

### SQLite

Tables for metadata, static manifest, stops/routes/shapes/schedules, favorites, recents, saved places/itineraries, bounded cached alert/vehicle details, subscription mirror, and preferences. Repositories return domain models.

### Zustand

Small stores for map filters/selection/camera, planner draft, transient connectivity/banners. Never duplicate server entities.

## Screens

### Nearby

Permission-neutral entry; contextual location request; typed/map-center alternatives; mode chips; `limitPerMode`; grouped accessible list and map; loading/empty/error/offline/stale states; save action. Tests include denied location and nearest two per selected mode.

### Stop

Name/ID, save, route/direction-grouped arrivals, scheduled/realtime text distinction, refresh, alert banners, map, schedule link, walking deep link, freshness explanation.

### Route

Direction selector, shape/stops, active vehicles, alerts, schedule, planner prefill.

### Vehicle map

Full-screen Mapbox plus bottom sheet; mode/route/direction/source/freshness filters; vehicle-ID search; one vehicle `ShapeSource`; native layers by mode/state; separate selected layer; optional low-zoom clustering only after usability tests; foreground-only polling; preserved camera; accessible list alternative; conservative/reduced-motion interpolation.

### Vehicle detail

Only sourced fields; freshness/source; trip/block/stop links; locate action; no “live” claim when stale.

### Alerts

Filter/sort, route/mode/stop relevance, detailed active periods/source, changed-content handling, optional subscription flag.

### Planner

Origin/destination/swap, current/map/saved/stop choices, depart/arrive controls, mode/transfer/walking/accessibility preferences, explicit ranking reason, timeline/map, replan, safe share/deep link.

### Schedules

Service date, route/direction/stop hierarchy, feed version/offline state, after-midnight GTFS time handling, accessible list/table.

### Saved/settings/credits

Favorites/recents, reorder/delete, offline indicators, theme/time/units/accessibility, notifications, Mapbox telemetry, location settings, clear data, sources/updates, privacy, licenses, full credits.

## Local lifecycle

Transactional/idempotent migrations; public cache replaceable; favorites/preferences durable. Static sync uses version/checksum/temp file/atomic swap. Corruption recovery recreates public cache while preserving user records. No migration depends on network.

## Network

One generated client adding request ID, app/build/platform/locale. Respect `Retry-After`; classify offline/timeout/validation/rate/source/server errors. Backoff only eligible background work. No raw coordinates or search strings in production logs.

## Accessibility

Semantic headings/roles; arrival announcements include route/headsign/estimate/freshness; map features mirrored and announced in list/sheet; maximum font scale tests; contrast/icon/text redundancy; reduce motion; manual VoiceOver/TalkBack release script.

## Mobile security

SecureStore installation ID; no TriMet AppID/server Mapbox token; minimal environment-specific public Mapbox token; no background location MVP; redacted logs; validated notification payloads; privacy-safe crash reporting.

## Deliverables

Reproducible builds, architecture, MVP screens, SQLite migrations/repos, generated API integration, Mapbox adapter/layers, accessibility report, Maestro suite, EAS config, store metadata/privacy/credits skeleton.
