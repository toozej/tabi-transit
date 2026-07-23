import type {
  Alert,
  NearbyStops,
  RouteDetail,
  RouteShape,
  Schedule,
  StaticManifest,
  Stop,
} from "@/domain/riderInfo";
import type { Arrival } from "@/domain/riderInfo";

const now = "2026-07-22T16:30:02Z";
export const riderFreshness = {
  source: "fixture-normalized-static",
  sourceUpdatedAt: now,
  fetchedAt: now,
  processedAt: now,
  status: "fresh" as const,
  ageSeconds: 4,
  isRealtime: false,
};
export const fixtureStops: Stop[] = [
  {
    id: "fixture:stop:101",
    name: "Burnside & 10th",
    coordinate: [-122.681, 45.522],
    modes: ["bus"],
    routeIds: ["fixture:route:20"],
    wheelchairAccessible: true,
  },
  {
    id: "fixture:stop:102",
    name: "Pioneer Square North",
    coordinate: [-122.677, 45.518],
    modes: ["light_rail"],
    routeIds: ["fixture:route:blue"],
    wheelchairAccessible: true,
  },
];
export const fixtureNearby: NearbyStops = {
  distanceType: "straight_line",
  freshness: riderFreshness,
  groups: [
    { mode: "bus", stops: [{ ...fixtureStops[0]!, distanceMeters: 180 }] },
    {
      mode: "light_rail",
      stops: [{ ...fixtureStops[1]!, distanceMeters: 320 }],
    },
  ],
};
export const fixtureArrivals: Arrival[] = [
  {
    id: "fixture:arrival:1",
    stopId: "fixture:stop:101",
    routeId: "fixture:route:20",
    headsign: "Gresham",
    scheduledAt: "2026-07-22T16:36:00Z",
    estimatedAt: "2026-07-22T16:38:00Z",
    status: "estimated",
    freshness: {
      ...riderFreshness,
      isRealtime: true,
      source: "fixture-normalized-arrivals",
    },
  },
];
export const fixtureRoute: RouteDetail = {
  route: {
    id: "fixture:route:20",
    mode: "bus",
    shortName: "20",
    longName: "Burnside/Stark",
    color: "15803D",
  },
  directions: [
    { directionId: 0, headsign: "Beaverton" },
    { directionId: 1, headsign: "Gresham" },
  ],
  staticFeedVersion: "fixture-v1",
  freshness: riderFreshness,
};
export const fixtureAlerts: Alert[] = [
  {
    id: "fixture:alert:1",
    revision: "1",
    header: "Minor delay",
    description: "Allow extra travel time.",
    effect: "significant_delays",
    severity: "warning",
    periods: [{ startAt: now }],
    source: "fixture-alerts",
    freshness: {
      ...riderFreshness,
      isRealtime: true,
      source: "fixture-alerts",
    },
  },
];
export const fixtureManifest: StaticManifest = {
  staticFeedVersion: "fixture-v1",
  publishedAt: now,
  artifacts: [
    {
      name: "stops",
      version: "fixture-v1",
      sha256: "a".repeat(64),
      mediaType: "application/json",
      sizeBytes: 1024,
    },
    {
      name: "schedules",
      version: "fixture-v1",
      sha256: "b".repeat(64),
      mediaType: "application/json",
      sizeBytes: 4096,
    },
  ],
  freshness: riderFreshness,
};
export const fixtureSchedule: Schedule = {
  stopId: "fixture:stop:101",
  serviceDate: "2026-07-22",
  staticFeedVersion: "fixture-v1",
  schedule: [
    {
      tripId: "fixture:trip:night",
      routeId: "fixture:route:20",
      stopId: "fixture:stop:101",
      serviceDate: "2026-07-22",
      serviceDaySeconds: 90_060,
      headsign: "Gresham",
    },
  ],
};
export const fixtureRouteShape: RouteShape = {
  type: "FeatureCollection",
  staticFeedVersion: "fixture-v1",
  features: [
    {
      type: "Feature",
      geometry: {
        type: "LineString",
        coordinates: [
          [-122.69, 45.52],
          [-122.68, 45.521],
          [-122.67, 45.522],
        ],
      },
      properties: {
        shapeId: "fixture:shape:20",
        routeId: "fixture:route:20",
        directionId: 1,
      },
    },
  ],
};
