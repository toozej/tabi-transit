import { describe, expect, it } from "vitest";

import {
  createNotificationDeepLink,
  parseNotificationDeepLink,
  validateQuietHours,
  validateSubscriptionDraft,
} from "./notifications";

const draft = {
  type: "service_alert" as const,
  scope: { routeId: "trimet:route:20" },
  quietHours: {
    startsAt: "22:00",
    endsAt: "07:00",
    timeZone: "America/Los_Angeles",
  },
  expiresAt: "2026-08-01T00:00:00Z",
};

describe("notification domain", () => {
  it("validates a bounded, future fixture subscription", () => {
    expect(
      validateSubscriptionDraft(draft, new Date("2026-07-23T00:00:00Z")),
    ).toEqual(draft);
  });

  it("rejects invalid IANA zones, equal quiet hours, and expired watches", () => {
    expect(() =>
      validateQuietHours({ ...draft.quietHours, timeZone: "Not/AZone" }),
    ).toThrow("IANA");
    expect(() =>
      validateQuietHours({ ...draft.quietHours, endsAt: "22:00" }),
    ).toThrow("start and end");
    expect(() =>
      validateSubscriptionDraft(
        { ...draft, expiresAt: "2026-07-01T00:00:00Z" },
        new Date("2026-07-23T00:00:00Z"),
      ),
    ).toThrow("future");
  });

  it("uses only opaque IDs in a validated notification deep link", () => {
    const link = createNotificationDeepLink({
      type: "service_alert",
      entityId: "trimet:alert:42",
      subscriptionId: "fixture:subscription:1",
      revision: "rev-2",
    });
    expect(link).not.toContain("45.");
    expect(parseNotificationDeepLink(link)).toMatchObject({
      entityId: "trimet:alert:42",
    });
    expect(() =>
      parseNotificationDeepLink(`${link}&coordinate=-122.6,45.5`),
    ).toThrow("field");
  });
});
