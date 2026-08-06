import { z } from "zod";
import type { paths } from "@tabi/api-client/src/generated/openapi";

import { webConfig } from "../app/config";
import { alerts, stops, vehicles } from "./fixtures";
import {
  alertSchema,
  stopSchema,
  vehicleSchema,
  type Alert,
  type Stop,
  type Vehicle,
} from "./models";

type GeneratedApiPath = Extract<keyof paths, `/v1/${string}`>;
const generatedEndpoints = {
  alerts: "/v1/alerts",
  nearby: "/v1/stops/nearby",
  vehicles: "/v1/vehicles",
} as const satisfies Record<string, GeneratedApiPath>;

const nearbyResponseSchema = z.object({
  groups: z.array(z.object({ stops: z.array(stopSchema) })),
});
const vehicleCollectionSchema = z.object({ vehicles: z.array(vehicleSchema) });
const alertCollectionSchema = z.object({ alerts: z.array(alertSchema) });
const stopDetailSchema = z.object({ stop: stopSchema });
const arrivalSchema = z.object({
  id: z.string(),
  routeId: z.string(),
  headsign: z.string().optional(),
  status: z.enum(["scheduled", "estimated", "cancelled", "skipped", "unknown"]),
  estimatedAt: z.string().optional(),
});
const arrivalCollectionSchema = z.object({ arrivals: z.array(arrivalSchema) });
const routeDetailSchema = z.object({
  route: z.object({
    id: z.string(),
    shortName: z.string(),
    longName: z.string(),
    mode: z.string(),
  }),
  directions: z.array(
    z.object({ directionId: z.number(), headsign: z.string().optional() }),
  ),
});
const routeStopCollectionSchema = z.object({
  stops: z.array(stopSchema.extend({ sequence: z.number() })),
});
const journeyPlanSchema = z.object({
  source: z.string(),
  itineraries: z.array(
    z.object({
      id: z.string(),
      durationSeconds: z.number(),
      transfers: z.number(),
      walkingMeters: z.number(),
      legs: z.array(
        z.object({
          mode: z.string(),
          fromName: z.string(),
          toName: z.string(),
          routeId: z.string().optional(),
        }),
      ),
    }),
  ),
});

export type NearbyPosition = { latitude: number; longitude: number };
export type Arrival = z.infer<typeof arrivalSchema>;
export type RouteDetail = z.infer<typeof routeDetailSchema>;
export type RouteStop = z.infer<
  typeof routeStopCollectionSchema
>["stops"][number];
export type JourneyPlan = z.infer<typeof journeyPlanSchema>;

export function flattenNearbyGroups(value: unknown): Stop[] {
  return nearbyResponseSchema
    .parse(value)
    .groups.flatMap((group) => group.stops);
}

export function collectionVehicles(value: unknown): Vehicle[] {
  return vehicleCollectionSchema.parse(value).vehicles;
}

export function collectionAlerts(value: unknown): Alert[] {
  return alertCollectionSchema.parse(value).alerts;
}

async function get<T>(
  path: string,
  schema: z.ZodType<T>,
  fixture: T,
): Promise<T> {
  if (webConfig.mode === "fixture") return fixture;
  const response = await fetch(`${webConfig.apiBaseUrl}${path}`, {
    headers: { Accept: "application/json" },
    credentials: "same-origin",
  });
  if (!response.ok) throw new Error(`Request failed (${response.status}).`);
  return schema.parse(await response.json());
}

// The generated API client is the source for request/response TypeScript types;
// Zod remains the mandatory runtime boundary for browser-delivered data.
export const queryKeys = {
  nearby: ["nearby"] as const,
  vehicles: ["vehicles"] as const,
  alerts: ["alerts"] as const,
};
export async function getNearby(position?: NearbyPosition): Promise<Stop[]> {
  if (webConfig.mode === "fixture") return stops;
  if (!position)
    throw new Error("Choose location access before loading nearby stops.");
  const query = new URLSearchParams({
    lat: String(position.latitude),
    lon: String(position.longitude),
  });
  const response = await get(
    `${generatedEndpoints.nearby.slice(3)}?${query}`,
    nearbyResponseSchema,
    { groups: [] },
  );
  return flattenNearbyGroups(response);
}
export async function getVehicles(): Promise<Vehicle[]> {
  if (webConfig.mode === "fixture") return vehicles;
  const response = await get(
    generatedEndpoints.vehicles.slice(3),
    vehicleCollectionSchema,
    { vehicles: [] },
  );
  return collectionVehicles(response);
}
export async function getAlerts(): Promise<Alert[]> {
  if (webConfig.mode === "fixture") return alerts;
  const response = await get(
    generatedEndpoints.alerts.slice(3),
    alertCollectionSchema,
    { alerts: [] },
  );
  return collectionAlerts(response);
}
export async function getStop(id: string): Promise<Stop> {
  if (webConfig.mode === "fixture") {
    const stop = stops.find((item) => item.id === id);
    if (!stop) throw new Error("Stop was not found.");
    return stop;
  }
  return (
    await get(`/stops/${encodeURIComponent(id)}`, stopDetailSchema, {
      stop: stops[0]!,
    })
  ).stop;
}
export async function getArrivals(id: string): Promise<Arrival[]> {
  if (webConfig.mode === "fixture") return [];
  return (
    await get(
      `/stops/${encodeURIComponent(id)}/arrivals`,
      arrivalCollectionSchema,
      { arrivals: [] },
    )
  ).arrivals;
}
export async function getRoute(id: string): Promise<RouteDetail> {
  if (webConfig.mode === "fixture")
    return {
      route: { id, shortName: "20", longName: "Burnside/Stark", mode: "bus" },
      directions: [],
    };
  return get(`/routes/${encodeURIComponent(id)}`, routeDetailSchema, {
    route: { id, shortName: "", longName: "", mode: "unknown" },
    directions: [],
  });
}
export async function getRouteStops(id: string): Promise<RouteStop[]> {
  if (webConfig.mode === "fixture") return [];
  return (
    await get(
      `/routes/${encodeURIComponent(id)}/stops`,
      routeStopCollectionSchema,
      { stops: [] },
    )
  ).stops.sort((left, right) => left.sequence - right.sequence);
}

export async function planJourney(
  origin: string,
  destination: string,
): Promise<JourneyPlan> {
  if (webConfig.mode === "fixture") {
    return {
      source: "fixture-planner",
      itineraries: [
        {
          id: "fixture:journey:1",
          durationSeconds: 1800,
          transfers: 0,
          walkingMeters: 350,
          legs: [{ mode: "walk", fromName: origin, toName: destination }],
        },
      ],
    };
  }
  const response = await fetch(`${webConfig.apiBaseUrl}/journeys/plan`, {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    credentials: "same-origin",
    body: JSON.stringify({
      origin: { type: "stop", stopId: origin, label: origin },
      destination: { type: "stop", stopId: destination, label: destination },
      time: { mode: "depart_at", value: new Date().toISOString() },
      preferences: {
        modes: ["bus", "light_rail", "streetcar", "commuter_rail"],
        wheelchairAccessible: false,
        optimize: "fastest",
      },
    }),
  });
  if (!response.ok) {
    throw new Error(
      response.status === 503
        ? "Trip planning is unavailable."
        : `Request failed (${response.status}).`,
    );
  }
  return journeyPlanSchema.parse(await response.json());
}
