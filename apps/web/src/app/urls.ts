import { z } from "zod";

export const opaqueIdSchema = z
  .string()
  .min(1)
  .max(160)
  .regex(/^[A-Za-z0-9:_-]+$/);
const plannerQuerySchema = z.object({
  origin: opaqueIdSchema.optional(),
  destination: opaqueIdSchema.optional(),
  maxTransfers: z.coerce.number().int().min(0).max(5).optional(),
  accessibility: z.literal("wheelchair").optional(),
});

export function validOpaqueId(value: string | undefined): string | undefined {
  return opaqueIdSchema.safeParse(value).data;
}

/** URLs may hold opaque IDs and choices, never coordinates or search strings. */
export function parsePlannerSearch(search: string) {
  const params = new URLSearchParams(search);
  return plannerQuerySchema.safeParse({
    origin: params.get("origin") ?? undefined,
    destination: params.get("destination") ?? undefined,
    maxTransfers: params.get("maxTransfers") ?? undefined,
    accessibility: params.get("accessibility") ?? undefined,
  });
}
