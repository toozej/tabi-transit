import { z } from "zod";

import type { components } from "@tabi/api-client";

const coordinateSchema = z.tuple([z.number(), z.number()]);
export const freshnessSchema = z.object({
  source: z.string().min(1),
  sourceUpdatedAt: z.string().datetime().optional(),
  entityUpdatedAt: z.string().datetime().optional(),
  fetchedAt: z.string().datetime(),
  processedAt: z.string().datetime(),
  status: z.enum(["fresh", "aging", "stale", "unknown"]),
  ageSeconds: z.number().nonnegative(),
  isRealtime: z.boolean(),
});

export const vehicleSchema = z.object({
  id: z.string().min(1),
  sourceVehicleId: z.string().min(1),
  mode: z.enum([
    "bus",
    "light_rail",
    "streetcar",
    "commuter_rail",
    "ferry",
    "unknown",
  ]),
  routeId: z.string().min(1).optional(),
  tripId: z.string().min(1).optional(),
  blockId: z.string().min(1).optional(),
  directionId: z.union([z.literal(0), z.literal(1)]).optional(),
  headsign: z.string().min(1).optional(),
  coordinate: coordinateSchema,
  bearing: z.number().min(0).max(360).optional(),
  speedMetersPerSecond: z.number().nonnegative().nullable().optional(),
  currentStopId: z.string().min(1).optional(),
  nextStopId: z.string().min(1).optional(),
  scheduleDeviationSeconds: z.number().int().nullable().optional(),
  inService: z.boolean(),
  freshness: freshnessSchema,
});

export const vehicleCollectionSchema = z.object({
  snapshotId: z.string().min(1),
  vehicles: z.array(vehicleSchema),
  freshness: freshnessSchema,
});
export const vehicleSearchSchema = z.object({
  query: z.string(),
  vehicles: z.array(vehicleSchema),
  freshness: freshnessSchema,
});
export const vehicleDetailSchema = z.object({ vehicle: vehicleSchema });
/**
 * A bounded, normalized vehicle-position timeline. It intentionally contains
 * no adherence classification: vehicle positions alone cannot establish
 * whether a service was early or late (ADR-0016).
 */
export const vehicleHistoryObservationSchema = z.object({
  coordinate: coordinateSchema,
  observedAt: z.string().datetime(),
  routeId: z.string().min(1).optional(),
  tripId: z.string().min(1).optional(),
  mode: vehicleSchema.shape.mode,
  freshness: z.object({
    status: freshnessSchema.shape.status,
    fetchedAt: z.string().datetime(),
  }),
});
export const vehicleHistorySchema = z.object({
  vehicleId: z.string().min(1),
  observations: z.array(vehicleHistoryObservationSchema),
  retentionDays: z.literal(30),
  freshness: z.object({
    status: z.literal("historical"),
    source: z.literal("normalized-vehicle-observations"),
  }),
});
export const configSchema = z.object({
  apiVersion: z.string().min(1),
  minimumAppVersion: z.string().min(1),
  features: z.record(
    z.object({ enabled: z.boolean(), reason: z.string().optional() }),
  ),
  sources: z.record(
    z.object({ enabled: z.boolean(), reason: z.string().optional() }),
  ),
  pollingRecommendations: z.record(z.number().int().positive()).optional(),
  staleThresholdSeconds: z.record(z.number().int().nonnegative()),
  serviceBounds: z.object({
    bbox: z.tuple([z.number(), z.number(), z.number(), z.number()]),
  }),
  staticFeed: z.object({
    version: z.string(),
    publishedAt: z.string().datetime(),
  }),
  urls: z.record(z.string().url()).optional(),
});

// The generated contract is compile-time evidence; Zod is the runtime boundary.
export type Vehicle = z.infer<typeof vehicleSchema> &
  components["schemas"]["Vehicle"];
export type VehicleCollection = z.infer<typeof vehicleCollectionSchema>;
export type VehicleSearch = z.infer<typeof vehicleSearchSchema>;
export type VehicleHistory = z.infer<typeof vehicleHistorySchema>;
export type VehicleHistoryObservation = z.infer<
  typeof vehicleHistoryObservationSchema
>;
export type PublicConfig = z.infer<typeof configSchema>;
export type VehicleMode = Vehicle["mode"];

export type VehicleFilters = {
  modes: VehicleMode[];
  routeId?: string;
  freshness?: Vehicle["freshness"]["status"];
};

export function filterVehicles(
  vehicles: readonly Vehicle[],
  filters: VehicleFilters,
): Vehicle[] {
  return vehicles.filter((vehicle) => {
    if (filters.modes.length > 0 && !filters.modes.includes(vehicle.mode))
      return false;
    if (filters.routeId && vehicle.routeId !== filters.routeId) return false;
    return !filters.freshness || vehicle.freshness.status === filters.freshness;
  });
}

export function formatFreshness(freshness: Vehicle["freshness"]): string {
  const label = freshness.status === "fresh" ? "Live" : freshness.status;
  return `${label}; ${Math.round(freshness.ageSeconds)} seconds old; source ${freshness.source}`;
}
