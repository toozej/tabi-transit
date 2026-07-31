import { describe, expect, it } from "vitest";

import {
  createFixtureNotificationRepository,
  unavailableNotificationRegistration,
} from "./notificationRepository";

describe("fixture notification repository", () => {
  it("creates and deletes locally validated subscriptions", async () => {
    const repository = createFixtureNotificationRepository(
      [],
      () => new Date("2026-07-23T00:00:00Z"),
    );
    const created = await repository.create({
      type: "departure_reminder",
      scope: { stopId: "trimet:stop:101" },
      leadMinutes: 10,
      expiresAt: "2026-07-24T00:00:00Z",
    });
    expect((await repository.list()).map((value) => value.id)).toEqual([
      created.id,
    ]);
    await repository.remove(created.id);
    expect(await repository.list()).toEqual([]);
  });

  it("is explicitly unavailable for device push registration", async () => {
    await expect(
      unavailableNotificationRegistration.status(),
    ).resolves.toMatchObject({
      available: false,
      reason: "device_push_disabled",
    });
  });

  it("uses its current clock when omitting expired fixture subscriptions", async () => {
    const repository = createFixtureNotificationRepository(
      [
        {
          id: "fixture:subscription:expired",
          type: "service_alert",
          scope: { routeId: "trimet:route:20" },
          createdAt: "2026-07-23T00:00:00Z",
          expiresAt: "2026-07-24T00:00:00Z",
        },
      ],
      () => new Date("2026-07-25T00:00:00Z"),
    );

    await expect(repository.list()).resolves.toEqual([]);
  });
});
