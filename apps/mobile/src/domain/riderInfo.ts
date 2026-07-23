import { z } from "zod";

import type { components } from "@tabi/api-client";

import { freshnessSchema } from "./vehicleModels";

export const transitModeSchema = z.enum([
  "bus",
  "light_rail",
  "streetcar",
  "commuter_rail",
  "ferry",
  "unknown",
]);
export const coordinateSchema = z.tuple([z.number(), z.number()]);
export const stopSchema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  coordinate: coordinateSchema,
  modes: z.array(transitModeSchema).min(1),
  parentStopId: z.string().optional(),
  routeIds: z.array(z.string()).optional(),
  wheelchairAccessible: z.boolean().optional(),
});
export const nearbyStopsSchema = z.object({
  distanceType: z.literal("straight_line"),
  groups: z.array(
    z.object({
      mode: transitModeSchema,
      stops: z.array(
        stopSchema.extend({ distanceMeters: z.number().nonnegative() }),
      ),
    }),
  ),
  freshness: freshnessSchema,
});
export const routeSchema = z.object({
  id: z.string().min(1),
  mode: transitModeSchema,
  shortName: z.string(),
  longName: z.string(),
  color: z.string().optional(),
  textColor: z.string().optional(),
});
export const routeDetailSchema = z.object({
  route: routeSchema,
  directions: z.array(
    z.object({
      directionId: z.union([z.literal(0), z.literal(1)]),
      headsign: z.string().optional(),
    }),
  ),
  staticFeedVersion: z.string(),
  alertIds: z.array(z.string()).optional(),
  freshness: freshnessSchema,
});
export const stopDetailSchema = z.object({
  stop: stopSchema,
  staticFeedVersion: z.string(),
  freshness: freshnessSchema,
});
export const arrivalSchema = z.object({
  id: z.string(),
  stopId: z.string(),
  routeId: z.string(),
  directionId: z.union([z.literal(0), z.literal(1)]).optional(),
  headsign: z.string().optional(),
  scheduledAt: z.string().datetime(),
  estimatedAt: z.string().datetime().optional(),
  status: z.enum(["scheduled", "estimated", "cancelled", "skipped", "unknown"]),
  freshness: freshnessSchema,
});
export const alertSchema = z.object({
  id: z.string(),
  revision: z.string(),
  header: z.string(),
  description: z.string().optional(),
  effect: z.string().optional(),
  severity: z.enum(["info", "warning", "severe", "unknown"]).optional(),
  periods: z.array(
    z.object({
      startAt: z.string().datetime().optional(),
      endAt: z.string().datetime().optional(),
    }),
  ),
  source: z.string(),
  freshness: freshnessSchema,
});
export const staticManifestSchema = z.object({
  staticFeedVersion: z.string(),
  publishedAt: z.string().datetime(),
  artifacts: z.array(
    z.object({
      name: z.enum(["stops", "routes", "schedules", "shapes"]),
      version: z.string(),
      sha256: z.string().regex(/^[a-f0-9]{64}$/),
      mediaType: z.literal("application/json"),
      sizeBytes: z.number().nonnegative(),
    }),
  ),
  freshness: freshnessSchema,
});
export const scheduleSchema = z.object({
  stopId: z.string(),
  serviceDate: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
  staticFeedVersion: z.string(),
  schedule: z.array(
    z.object({
      tripId: z.string(),
      routeId: z.string(),
      stopId: z.string(),
      serviceDate: z.string(),
      serviceDaySeconds: z.number().int().min(0).max(172799),
      headsign: z.string().optional(),
    }),
  ),
});
export const routeShapeSchema = z.object({
  type: z.literal("FeatureCollection"),
  staticFeedVersion: z.string(),
  features: z.array(
    z.object({
      type: z.literal("Feature"),
      geometry: z.object({
        type: z.literal("LineString"),
        coordinates: z.array(coordinateSchema).min(2),
      }),
      properties: z.object({
        shapeId: z.string(),
        routeId: z.string(),
        directionId: z.union([z.literal(0), z.literal(1)]).optional(),
      }),
    }),
  ),
});

export type Stop = z.infer<typeof stopSchema> & components["schemas"]["Stop"];
export type NearbyStops = z.infer<typeof nearbyStopsSchema>;
export type RouteDetail = z.infer<typeof routeDetailSchema>;
export type Arrival = z.infer<typeof arrivalSchema>;
export type Alert = z.infer<typeof alertSchema>;
export type StaticManifest = z.infer<typeof staticManifestSchema>;
export type Schedule = z.infer<typeof scheduleSchema>;
export type RouteShape = z.infer<typeof routeShapeSchema>;

/** Keeps "nearest N of each mode" semantically distinct from a total limit. */
export function applyNearbyLimit(
  nearby: NearbyStops,
  limitPerMode?: number,
): NearbyStops {
  if (!limitPerMode || limitPerMode < 1) return nearby;
  return {
    ...nearby,
    groups: nearby.groups.map((group) => ({
      ...group,
      stops: [...group.stops]
        .sort((a, b) => a.distanceMeters - b.distanceMeters)
        .slice(0, limitPerMode),
    })),
  };
}

export function formatDistance(distanceMeters: number): string {
  return distanceMeters >= 1000
    ? `${(distanceMeters / 1000).toFixed(1)} km away`
    : `${Math.round(distanceMeters)} m away`;
}

export function serviceDayTime(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
}
