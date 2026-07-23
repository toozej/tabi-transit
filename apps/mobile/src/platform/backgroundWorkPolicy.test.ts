import { describe, expect, it } from "vitest";

import { mayScheduleBackgroundWork } from "./backgroundWorkPolicy";

describe("background work policy", () => {
  it("allows only deferrable maintenance", () => {
    expect(mayScheduleBackgroundWork("refresh_static_manifest")).toBe(true);
    expect(mayScheduleBackgroundWork("reconcile_subscriptions")).toBe(true);
    expect(mayScheduleBackgroundWork("vehicle_polling")).toBe(false);
    expect(mayScheduleBackgroundWork("location_monitoring")).toBe(false);
  });
});
