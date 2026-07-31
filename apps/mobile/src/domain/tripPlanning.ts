import { z } from "zod";

import { transitModeSchema } from "./riderInfo";

const placeIdSchema = z
  .string()
  .min(1)
  .max(120)
  .regex(/^[a-zA-Z0-9:_-]+$/);

export const planEndpointSchema = z.object({
  id: placeIdSchema,
  label: z.string().min(1).max(160),
  kind: z.enum(["current_location", "map_pin", "saved_place", "stop", "typed"]),
});
export const plannerConstraintsSchema = z.object({
  modes: z.array(transitModeSchema).min(1),
  maxTransfers: z.number().int().min(0).max(5).optional(),
  maxWalkingMeters: z.number().int().min(0).max(20_000).optional(),
  wheelchairAccessible: z.boolean(),
  optimization: z.enum(["fastest", "fewer_transfers", "least_walking"]),
  timeMode: z.enum(["depart_at", "arrive_by"]),
  time: z.string().datetime(),
});
export const plannerDraftSchema = z.object({
  origin: planEndpointSchema.optional(),
  destination: planEndpointSchema.optional(),
  constraints: plannerConstraintsSchema,
});

export const itineraryLegSchema = z.object({
  id: z.string().min(1),
  mode: z.enum(["walk", ...transitModeSchema.options]),
  startLabel: z.string().min(1),
  endLabel: z.string().min(1),
  startAt: z.string().datetime(),
  endAt: z.string().datetime(),
  durationSeconds: z.number().int().nonnegative(),
  walkingMeters: z.number().int().nonnegative().optional(),
  routeLabel: z.string().optional(),
  headsign: z.string().optional(),
  realtime: z.enum(["scheduled", "estimated", "unknown"]),
  geometry: z
    .array(z.tuple([z.number(), z.number()]))
    .min(2)
    .optional(),
});
export const itinerarySchema = z.object({
  id: z.string().min(1),
  departureAt: z.string().datetime(),
  arrivalAt: z.string().datetime(),
  durationSeconds: z.number().int().nonnegative(),
  transfers: z.number().int().nonnegative(),
  walkingMeters: z.number().int().nonnegative(),
  wheelchairAccessible: z.boolean().optional(),
  // A fixture is the default, but a composed backend planner returns a
  // normalized provider identifier. Provider payloads are never exposed here.
  source: z.enum(["fixture-planner", "trimet-web-services"]),
  freshness: z.object({
    status: z.enum(["fixture", "fresh", "unknown"]),
    message: z.string(),
  }),
  legs: z.array(itineraryLegSchema).min(1),
});

export type PlanEndpoint = z.infer<typeof planEndpointSchema>;
export type PlannerConstraints = z.infer<typeof plannerConstraintsSchema>;
export type PlannerDraft = z.infer<typeof plannerDraftSchema>;
export type ItineraryLeg = z.infer<typeof itineraryLegSchema>;
export type Itinerary = z.infer<typeof itinerarySchema>;

export const defaultPlannerConstraints: PlannerConstraints = {
  modes: ["bus", "light_rail", "commuter_rail", "streetcar"],
  wheelchairAccessible: false,
  optimization: "fastest",
  timeMode: "depart_at",
  time: new Date().toISOString(),
};

export function swapEndpoints(draft: PlannerDraft): PlannerDraft {
  return { ...draft, origin: draft.destination, destination: draft.origin };
}

export type RankedItineraries = {
  itineraries: Itinerary[];
  disclosure?: string;
};

/** Deterministic client policy for fixture display; the backend remains authoritative. */
export function rankItineraries(
  itineraries: Itinerary[],
  constraints: PlannerConstraints,
): RankedItineraries {
  const matches = itineraries.filter((itinerary) => {
    const requestedAt = Date.parse(constraints.time);
    if (
      (constraints.timeMode === "depart_at" &&
        Date.parse(itinerary.departureAt) < requestedAt) ||
      (constraints.timeMode === "arrive_by" &&
        Date.parse(itinerary.arrivalAt) > requestedAt)
    )
      return false;
    if (
      constraints.maxTransfers !== undefined &&
      itinerary.transfers > constraints.maxTransfers
    )
      return false;
    if (
      constraints.maxWalkingMeters !== undefined &&
      itinerary.walkingMeters > constraints.maxWalkingMeters
    )
      return false;
    if (constraints.wheelchairAccessible && !itinerary.wheelchairAccessible)
      return false;
    return itinerary.legs.every(
      (leg) => leg.mode === "walk" || constraints.modes.includes(leg.mode),
    );
  });
  const ordered = [...matches].sort((left, right) => {
    if (constraints.optimization === "fewer_transfers")
      return (
        left.transfers - right.transfers ||
        left.durationSeconds - right.durationSeconds
      );
    if (constraints.optimization === "least_walking")
      return (
        left.walkingMeters - right.walkingMeters ||
        left.durationSeconds - right.durationSeconds
      );
    return left.durationSeconds - right.durationSeconds;
  });
  return ordered.length > 0
    ? { itineraries: ordered }
    : {
        itineraries: [],
        disclosure:
          "No fixture itinerary meets these constraints. Relax a filter to view available alternatives.",
      };
}

const deepLinkSchema = z.object({
  origin: placeIdSchema.optional(),
  destination: placeIdSchema.optional(),
  maxTransfers: z.coerce.number().int().min(0).max(5).optional(),
  accessibility: z.enum(["wheelchair"]).optional(),
});

/** Deep links carry opaque place IDs only, never precise coordinates or search text. */
export function parsePlanDeepLink(
  value: string,
): z.infer<typeof deepLinkSchema> {
  const url = new URL(value);
  if (url.protocol !== "tabi:" || url.hostname !== "plan")
    throw new Error("Unsupported Tabi planning link.");
  return deepLinkSchema.parse({
    origin: url.searchParams.get("origin") ?? undefined,
    destination: url.searchParams.get("destination") ?? undefined,
    maxTransfers: url.searchParams.get("maxTransfers") ?? undefined,
    accessibility: url.searchParams.get("accessibility") ?? undefined,
  });
}

export function createPlanDeepLink(draft: PlannerDraft): string {
  const params = new URLSearchParams();
  if (draft.origin) params.set("origin", draft.origin.id);
  if (draft.destination) params.set("destination", draft.destination.id);
  if (draft.constraints.maxTransfers !== undefined)
    params.set("maxTransfers", String(draft.constraints.maxTransfers));
  if (draft.constraints.wheelchairAccessible)
    params.set("accessibility", "wheelchair");
  return `tabi://plan?${params.toString()}`;
}
