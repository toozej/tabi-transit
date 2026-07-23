import type { NotificationSubscription } from "@/domain/notifications";

export const fixtureNotificationSubscriptions: NotificationSubscription[] = [
  {
    id: "fixture:subscription:service-alert-20",
    type: "service_alert",
    scope: { routeId: "trimet:route:20", source: "fixture-alerts" },
    quietHours: {
      startsAt: "22:00",
      endsAt: "07:00",
      timeZone: "America/Los_Angeles",
    },
    expiresAt: "2026-08-01T00:00:00Z",
    createdAt: "2026-07-23T00:00:00Z",
  },
];
