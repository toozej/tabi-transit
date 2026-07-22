---
doc_id: TAB-PLAN-017
title: "Initial API Contract"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["api-agent", "backend-agent", "mobile-agent"]
depends_on: ["TAB-PLAN-001", "TAB-PLAN-002"]
---


# Initial API Contract

`api/openapi.yaml` becomes executable source of truth in WP-02.

## Conventions

`/v1`; JSON UTF-8; ISO 8601 UTC instants; `YYYY-MM-DD` service dates; longitude/latitude; meters/seconds; opaque cursors; source-qualified IDs; safe `unknown` enum; request ID; ETag on snapshots/static data.

## Common freshness

```json
{
  "source": "trimet-gtfsrt-vehicle-positions",
  "sourceUpdatedAt": "2026-07-22T16:30:00Z",
  "entityUpdatedAt": "2026-07-22T16:29:55Z",
  "fetchedAt": "2026-07-22T16:30:02Z",
  "processedAt": "2026-07-22T16:30:02Z",
  "status": "fresh",
  "ageSeconds": 7,
  "isRealtime": true
}
```

Status: `fresh | aging | stale | unknown`.

## Error

```json
{
  "error": {
    "code": "source_unavailable",
    "message": "Vehicle positions are temporarily unavailable.",
    "requestId": "req_...",
    "retryAfterSeconds": 30,
    "source": "trimet-gtfsrt-vehicle-positions",
    "details": []
  }
}
```

Codes: `validation_error`, `not_found`, `conflict`, `rate_limited`, `source_unavailable`, `temporarily_unavailable`, `internal_error`.

## `GET /v1/config`

API/build/minimum app, enabled features/sources, polling display recommendations, stale thresholds, service bounds, static feed manifest/version, support/status/privacy/credits URLs. No secrets.

## Stops

### `GET /v1/stops/nearby`

Query: `lat`, `lon`, bounded `radiusMeters`, `modes`, `limit`, `limitPerMode`, `includeArrivals`, `wheelchairAccessible`.

Response includes `distanceType: straight_line`, mode groups, stop IDs/names/coordinates/distance/parent/routes, and freshness.

### `GET /v1/stops/{id}`

Stable details and routes.

### `GET /v1/stops/{id}/arrivals`

Minutes, route/direction filters, scheduled inclusion. Arrival contains route/headsign, scheduled/estimated, status, vehicle/trip/sequence, realtime, freshness, alert refs.

### `GET /v1/stops/{id}/schedule`

Service date, route, direction, cursor.

## Routes

- `GET /v1/routes` with mode/date/query.
- `GET /v1/routes/{id}`.
- `GET /v1/routes/{id}/shape` as GeoJSON plus feed version.
- `GET /v1/routes/{id}/stops`.
- `GET /v1/routes/{id}/vehicles`.

## Vehicles

### `GET /v1/vehicles`

Filters: modes, routes, directions, sources, freshness, optional bbox, `format=json|geojson`. Include snapshot/ETag/collection freshness.

```json
{
  "id": "trimet:vehicle:2901",
  "sourceVehicleId": "2901",
  "mode": "bus",
  "routeId": "trimet:route:20",
  "tripId": "trimet:trip:...",
  "blockId": "trimet:block:...",
  "directionId": 1,
  "headsign": "Gresham",
  "coordinate": [-122.67, 45.52],
  "bearing": 90,
  "speedMetersPerSecond": null,
  "currentStopId": null,
  "nextStopId": "trimet:stop:...",
  "scheduleDeviationSeconds": 120,
  "inService": true,
  "freshness": {}
}
```

- `GET /v1/vehicles/search?q=`.
- `GET /v1/vehicles/{id}` enriched current detail.

## Alerts

- `GET /v1/alerts` filters route/stop/mode/effect/active/updated.
- `GET /v1/alerts/{id}`.

Fields: ID/revision, header/description, cause/effect/severity if sourced, periods, informed entities, URL, source, freshness.

## Search

`GET /v1/search` with query/types/proximity/bbox/language/session token semantics. Combines Tabi transit and Mapbox places with source/attribution.

`GET /v1/geocode/reverse` with terms-aware caching/storage.

## Journeys

### `POST /v1/journeys/plan`

```json
{
  "origin": {"type": "coordinate", "coordinate": [-122.68, 45.52], "label": "Current location"},
  "destination": {"type": "stop", "stopId": "trimet:stop:8334"},
  "time": {"mode": "depart_at", "value": "2026-07-22T17:00:00-07:00"},
  "preferences": {
    "modes": ["bus", "light_rail", "streetcar"],
    "maxTransfers": 2,
    "wheelchairAccessible": false,
    "optimize": "fewer_transfers"
  }
}
```

Return plan ID/expiry, normalized endpoints, itineraries, applied/unsupported preferences, provider/freshness/alerts. Itinerary contains times, duration, transfer count, walking, accessibility, score/reason, and `walk|transit|wait|transfer` legs with route/trip/stops/realtime/geometry/alerts.

## Installations/subscriptions

- `POST /v1/installations` returns ID and one-time installation credential.
- `PUT /v1/installations/{id}/push-token`.
- `DELETE /v1/installations/{id}`.
- `GET/POST /v1/subscriptions`.
- `DELETE /v1/subscriptions/{id}`.

Authenticated by installation credential. Subscription types service alert, departure reminder, and later arrival threshold; scopes, quiet hours, time zone, expiry.

## Headers

Request: request ID, app version/build/platform, locale, ETag. Response: request ID, ETag, cache, retry-after, API version, static-feed version.

## Contract tests

OpenAPI request/response, golden examples, backward diff, generated client compile, unknown enum/extra field tolerance, error/freshness, ETag/304.
