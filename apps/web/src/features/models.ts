import { z } from "zod";

export const freshnessSchema = z.object({
  source: z.string(),
  fetchedAt: z.string(),
  status: z.enum(["fresh", "aging", "stale", "unknown"]),
  ageSeconds: z.number().nonnegative(),
  isRealtime: z.boolean(),
});
export type Freshness = z.infer<typeof freshnessSchema>;
export const stopSchema = z.object({
  id: z.string(),
  name: z.string(),
  modes: z.array(z.string()),
  distanceMeters: z.number().optional(),
  wheelchairAccessible: z.boolean().optional(),
});
export type Stop = z.infer<typeof stopSchema>;
export const vehicleSchema = z.object({
  id: z.string(),
  sourceVehicleId: z.string(),
  mode: z.string(),
  routeId: z.string().optional(),
  headsign: z.string().optional(),
  coordinate: z.tuple([z.number(), z.number()]),
  bearing: z.number().min(0).max(360).optional(),
  inService: z.boolean(),
  freshness: freshnessSchema,
});
export type Vehicle = z.infer<typeof vehicleSchema>;
export const alertSchema = z.object({
  id: z.string(),
  header: z.string(),
  description: z.string().optional(),
  severity: z.string().optional(),
  freshness: freshnessSchema,
});
export type Alert = z.infer<typeof alertSchema>;

export function freshnessLabel(freshness: Freshness): string {
  const state =
    freshness.status === "fresh" && freshness.isRealtime
      ? "Live"
      : freshness.status;
  return `${state}; ${Math.round(freshness.ageSeconds)} seconds old; source ${freshness.source}`;
}
