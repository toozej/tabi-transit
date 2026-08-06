---
doc_id: TAB-PLAN-005
title: "Maps, Search, Geocoding and Trip Planning"
status: implementation-ready
last_updated: 2026-08-05
intended_agents: ["maps-agent", "mobile-agent", "backend-agent"]
depends_on: ["TAB-PLAN-002", "TAB-PLAN-004", "TAB-PLAN-006", "TAB-PLAN-017"]
---


# Maps, Search, Geocoding and Trip Planning

## Mapbox credentials and compliance

Use separate environment tokens:

- Mobile public token with only native Maps SDK/style scopes.
- Browser public token with only browser Maps SDK/style scopes and an allowlist
  of Tabi's public origins. This token is intentionally exposed to the browser;
  it must never be a server, download, Search, or Geocoding token.
- CI/EAS SDK download token if required, stored as a secret.
- Backend Search/Geocoding token with minimum scopes.
- Never use a broad account default token in production.
- Configure Mapbox budgets, alerts, usage dashboards, and rotation.
- Retain required logo/attribution and expose the native SDK telemetry/anonymous usage control.

## Browser Mapbox GL integration

- Use the pinned `mapbox-gl` browser SDK behind `apps/web/src/maps/VehicleMap.tsx`.
- Configure its optional restricted public token as
  `VITE_MAPBOX_ACCESS_TOKEN`, and its approved style URL as
  `VITE_MAPBOX_STYLE_URL`. `VITE_*` values are public build-time values.
- Keep vehicle rendering to GeoJSON sources and style layers; do not create one
  DOM marker per vehicle. Keep the selected vehicle in its own source.
- Preserve the complete semantic vehicle list and detail flow when the token is
  absent or the map fails. Browser map rendering does not authorize Mapbox
  Search, Geocoding, Directions, or any provider call from the Tabi API.
- Before production enablement, complete D-020: public-origin restriction,
  approved style/attribution and telemetry treatment, budget alerts, and
  supported-browser/accessibility checks.

## Expo/RNMapbox integration

- Add `@rnmapbox/maps` config plugin to `app.config.ts`.
- Use Expo development builds; Expo Go is unsupported for this native module.
- Prefer Continuous Native Generation/prebuild; commit native projects only after an ADR.
- Use Mapbox native SDK v11 through a pinned compatible RNMapbox release.
- Configure foreground location permission through Expo Location.
- Validate iOS/Android and New Architecture in Phase 0.

## Map style/layers

Create light and dark Mapbox styles, and evaluate a high-contrast style. Runtime layers sit above the base map:

```text
route-shapes-source
  route-casing
  route-line

stops-source
  stop-circle
  station-symbol
  selected-stop

vehicles-source
  stale-vehicle
  bus-symbol
  max-symbol
  wes-symbol
  streetcar-symbol
  aerial-tram-symbol
  selected-halo
  selected-symbol
```

Rules:

- Use `ShapeSource` plus `SymbolLayer`/`CircleLayer`/`LineLayer`.
- Put mode, route, direction, source, freshness, and stable ID in properties.
- Use `MarkerView` only for a few interactive origin/destination pins.
- Separate selected state from the fleet collection.
- Batch updates, use stable feature IDs, query rendered features on press.
- Test at expected fleet, 1.5x headroom, and 3x stress.

## Vehicle interpolation and camera

Interpolation is presentation-only, short-window, and disabled for stale/unknown timestamp, impossible jumps, missing bearing, or reduce-motion. Detail timestamps always show source age.

Camera rules: default service-area bounds; one-time recenter on explicit location action; no forced continuous tracking; fit route shapes; pad for sheets; predictable search zoom; preserve tab camera; visible reset.

## Search/geocoding

Provider plan:

- Search Tabi stop, station, route, and vehicle entities first.
- Use Mapbox Search Box for POI/autocomplete if current terms and Portland quality pass the Phase 0 review.
- Use Mapbox Geocoding for address, street, intersection, and reverse geocoding.
- Proxy/normalize through the backend so the mobile contract is provider-neutral and quotas/cost are monitored.

Normalized result:

```ts
type PlaceResult = {
  id: string;
  source: "tabi" | "mapbox";
  kind: "stop" | "station" | "route" | "vehicle" | "address" | "poi" | "intersection" | "place";
  name: string;
  subtitle?: string;
  coordinate?: [number, number];
  bbox?: [number, number, number, number];
  stopId?: string;
  routeId?: string;
  vehicleId?: string;
  providerFeatureId?: string;
  attribution?: string;
};
```

Behavior:

1. Normalize input and detect exact stop/route/vehicle patterns.
2. Return exact Tabi matches first.
3. Debounce/cancel external autocomplete.
4. Bias to TriMet service area and current proximity.
5. Deduplicate equivalent results.
6. Show source attribution.
7. Apply provider session-token requirements.
8. Do not persist temporary Mapbox data beyond allowed use.
9. Rate-limit and meter cost.

## Nearby semantics

Inputs: lat/lon, radius, modes, overall limit, `limitPerMode`, accessibility, optional arrivals.

Backend:

- validate coordinates, limits, and service-area margin;
- use PostGIS `ST_DWithin` and spatial nearest ordering;
- document platform versus parent-station grouping;
- return straight-line distance first and label it;
- group per normalized mode.

Mobile: same entities feed list/map; location denied uses typed/map-center; stable stop metadata cached but raw user coordinates are not retained.

## Trip planning

### MVP adapter: TriMet Trip Planner

Backend converts normalized requests to the official planner, maps provider output to Tabi itinerary objects, attaches alerts/freshness, applies deterministic post-filters/ranking, and keeps provider details out of mobile.

### Filters/ranking

Evaluate depart/arrive compatibility, duration, transfer count, walking, selected modes, accessibility, alert impact, and realtime confidence. A hard `maxTransfers` excludes violations. When no result remains, show nearest alternatives and explain that the constraint could not be met.

### Geometry

Prefer provider geometry. Otherwise use a reliable GTFS shape slice for transit. Walking geometry may use Mapbox walking Directions only after terms/cost review; otherwise use a clearly approximate dashed connector. Mapbox driving/cycling is not a transit planner.

### OpenTripPlanner gate

Add OTP only when TriMet planner is proven insufficient for filters, accessibility, multi-agency, quality, latency, or availability and the team accepts Java/OSM/graph operations. Build regional OSM+approved GTFS graph, connect realtime, implement the same Tabi contract, dual-run/compare, then switch by feature flag.

## External deep links

Support Apple Maps, Google Maps, and documented compatible apps through installed-app detection and browser fallback. Label actions accurately, for example “Open walking directions.”

## Acceptance

- Physical iOS/Android rendering and location puck.
- Fleet headroom meets map budget.
- Filters use native layer expressions and avoid unnecessary refetch.
- Exact transit ID search outranks geocoding.
- Search/storage/attribution comply with Mapbox terms.
- Location-denied search and planning work.
- Trip filters are enforced or disclosed.
- Every map workflow has list/text completion.
