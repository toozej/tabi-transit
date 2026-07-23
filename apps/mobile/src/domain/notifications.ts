import { z } from "zod";

import { transitModeSchema } from "./riderInfo";

const opaqueIdSchema = z
  .string()
  .min(1)
  .max(120)
  .regex(/^[a-zA-Z0-9:_-]+$/);
const clockTimeSchema = z.string().regex(/^([01]\d|2[0-3]):[0-5]\d$/);

export const notificationTypeSchema = z.enum([
  "service_alert",
  "departure_reminder",
]);
export const quietHoursSchema = z.object({
  startsAt: clockTimeSchema,
  endsAt: clockTimeSchema,
  timeZone: z.string().min(1).max(100),
});
export const notificationScopeSchema = z
  .object({
    routeId: opaqueIdSchema.optional(),
    stopId: opaqueIdSchema.optional(),
    mode: transitModeSchema.optional(),
    source: z.string().min(1).max(80).optional(),
  })
  .refine(
    (scope) =>
      scope.routeId !== undefined ||
      scope.stopId !== undefined ||
      scope.mode !== undefined ||
      scope.source !== undefined,
    "A notification must be scoped to a route, stop, mode, or source.",
  );
const notificationSubscriptionDraftShape = z.object({
  type: notificationTypeSchema,
  scope: notificationScopeSchema,
  quietHours: quietHoursSchema.optional(),
  expiresAt: z.string().datetime().optional(),
  leadMinutes: z.number().int().min(1).max(120).optional(),
});

function validateSubscriptionType(
  value: z.infer<typeof notificationSubscriptionDraftShape>,
  context: z.RefinementCtx,
) {
  if (value.type === "departure_reminder" && value.leadMinutes === undefined) {
    context.addIssue({
      code: z.ZodIssueCode.custom,
      message: "A departure reminder needs a lead time.",
      path: ["leadMinutes"],
    });
  }
  if (value.type === "service_alert" && value.leadMinutes !== undefined) {
    context.addIssue({
      code: z.ZodIssueCode.custom,
      message: "Service alerts do not use a departure lead time.",
      path: ["leadMinutes"],
    });
  }
}

export const notificationSubscriptionDraftSchema =
  notificationSubscriptionDraftShape.superRefine(validateSubscriptionType);
export const notificationSubscriptionSchema = notificationSubscriptionDraftShape
  .extend({
    id: opaqueIdSchema,
    createdAt: z.string().datetime(),
  })
  .superRefine(validateSubscriptionType);

export type NotificationSubscriptionDraft = z.infer<
  typeof notificationSubscriptionDraftSchema
>;
export type NotificationSubscription = z.infer<
  typeof notificationSubscriptionSchema
>;

export function validateQuietHours(value: z.infer<typeof quietHoursSchema>) {
  try {
    Intl.DateTimeFormat("en-US", { timeZone: value.timeZone });
  } catch {
    throw new Error("Choose a valid IANA time zone for quiet hours.");
  }
  if (value.startsAt === value.endsAt)
    throw new Error("Quiet hours must have a start and end time.");
  return value;
}

export function validateSubscriptionDraft(
  value: NotificationSubscriptionDraft,
  now = new Date(),
): NotificationSubscriptionDraft {
  const parsed = notificationSubscriptionDraftSchema.parse(value);
  if (parsed.quietHours) validateQuietHours(parsed.quietHours);
  if (parsed.expiresAt && new Date(parsed.expiresAt).getTime() <= now.getTime())
    throw new Error("A notification expiry must be in the future.");
  return parsed;
}

const notificationLinkSchema = z.object({
  type: notificationTypeSchema,
  entityId: opaqueIdSchema,
  subscriptionId: opaqueIdSchema,
  revision: opaqueIdSchema.optional(),
});
export type NotificationDeepLink = z.infer<typeof notificationLinkSchema>;

/** Notification links contain only opaque IDs; payloads never hold coordinates, tokens, or itineraries. */
export function parseNotificationDeepLink(value: string): NotificationDeepLink {
  const url = new URL(value);
  if (url.protocol !== "tabi:" || url.hostname !== "notification")
    throw new Error("Unsupported Tabi notification link.");
  const permitted = new Set(["type", "entityId", "subscriptionId", "revision"]);
  for (const key of url.searchParams.keys()) {
    if (!permitted.has(key))
      throw new Error("Unsupported notification link field.");
  }
  return notificationLinkSchema.parse({
    type: url.searchParams.get("type"),
    entityId: url.searchParams.get("entityId"),
    subscriptionId: url.searchParams.get("subscriptionId"),
    revision: url.searchParams.get("revision") ?? undefined,
  });
}

export function createNotificationDeepLink(
  value: NotificationDeepLink,
): string {
  const parsed = notificationLinkSchema.parse(value);
  const params = new URLSearchParams({
    type: parsed.type,
    entityId: parsed.entityId,
    subscriptionId: parsed.subscriptionId,
  });
  if (parsed.revision) params.set("revision", parsed.revision);
  return `tabi://notification?${params.toString()}`;
}
