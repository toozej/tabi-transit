import type {
  PublicConfig,
  Vehicle,
  VehicleCollection,
  VehicleHistory,
} from "@/domain/vehicleModels";

const processedAt = "2026-07-22T16:30:02Z";
const fresh = {
  source: "fixture-normalized-vehicle-positions",
  sourceUpdatedAt: "2026-07-22T16:30:00Z",
  fetchedAt: processedAt,
  processedAt,
  status: "fresh" as const,
  ageSeconds: 7,
  isRealtime: true,
};

const stale = { ...fresh, status: "stale" as const, ageSeconds: 240 };

export const fixtureVehicles: Vehicle[] = [
  {
    id: "fixture:vehicle:2901",
    sourceVehicleId: "2901",
    mode: "bus",
    routeId: "fixture:route:20",
    directionId: 1,
    headsign: "Gresham",
    coordinate: [-122.67, 45.52],
    bearing: 90,
    nextStopId: "fixture:stop:101",
    scheduleDeviationSeconds: 120,
    inService: true,
    freshness: fresh,
  },
  {
    id: "fixture:vehicle:77",
    sourceVehicleId: "77",
    mode: "light_rail",
    routeId: "fixture:route:blue",
    directionId: 0,
    headsign: "Hillsboro",
    coordinate: [-122.68, 45.53],
    bearing: 270,
    inService: true,
    freshness: fresh,
  },
  {
    id: "fixture:vehicle:900",
    sourceVehicleId: "900",
    mode: "streetcar",
    routeId: "fixture:route:a",
    coordinate: [-122.66, 45.51],
    inService: false,
    freshness: stale,
  },
];

export const fixtureCollection: VehicleCollection = {
  snapshotId: "fixture-snapshot-42",
  vehicles: fixtureVehicles,
  freshness: fresh,
};

/** Sanitized, normalized history fixture; it is not an upstream payload. */
export const fixtureVehicleHistory: Record<string, VehicleHistory> = {
  "fixture:vehicle:2901": {
    vehicleId: "fixture:vehicle:2901",
    retentionDays: 30,
    observations: [
      {
        coordinate: [-122.672, 45.519],
        observedAt: "2026-07-22T16:15:02Z",
        routeId: "fixture:route:20",
        mode: "bus",
        freshness: {
          status: "fresh",
          fetchedAt: "2026-07-22T16:15:02Z",
        },
      },
      {
        coordinate: [-122.67, 45.52],
        observedAt: processedAt,
        routeId: "fixture:route:20",
        mode: "bus",
        freshness: { status: "fresh", fetchedAt: processedAt },
      },
    ],
    freshness: {
      status: "historical",
      source: "normalized-vehicle-observations",
    },
  },
};

export const fixtureConfig: PublicConfig = {
  apiVersion: "0.1.0",
  minimumAppVersion: "0.1.0",
  features: { vehicleMap: { enabled: true } },
  sources: {
    fixtureNormalizedVehicles: { enabled: true },
    trimetStreetcar: { enabled: true },
  },
  pollingRecommendations: { vehiclesSeconds: 15 },
  staleThresholdSeconds: { vehicles: 90 },
  serviceBounds: { bbox: [-123, 45.3, -122.3, 45.8] },
  staticFeed: { version: "fixture-v1", publishedAt: processedAt },
};
