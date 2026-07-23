import { beforeEach, describe, expect, it } from "vitest";

import { useNotificationStore } from "./notificationStore";

beforeEach(() => useNotificationStore.setState({ subscriptions: [] }));

describe("notification UI mirror", () => {
  it("adds and removes a safe subscription by opaque ID", () => {
    useNotificationStore.getState().addSubscription({
      id: "fixture:subscription:1",
      type: "service_alert",
      scope: { routeId: "trimet:route:20" },
      createdAt: "2026-07-23T00:00:00Z",
    });
    useNotificationStore
      .getState()
      .removeSubscription("fixture:subscription:1");
    expect(useNotificationStore.getState().subscriptions).toEqual([]);
  });
});
