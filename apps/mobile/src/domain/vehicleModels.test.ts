import { describe, expect, it } from "vitest";

import { fixtureVehicles } from "@/data/api/fixtures";

import { filterVehicles, formatFreshness } from "./vehicleModels";

describe("vehicle filtering and freshness", () => {
  it("filters by mode, route, and freshness without mutating server state", () => {
    expect(
      filterVehicles(fixtureVehicles, {
        modes: ["bus"],
        routeId: "fixture:route:20",
        freshness: "fresh",
      }),
    ).toEqual([fixtureVehicles[0]]);
    expect(fixtureVehicles).toHaveLength(3);
  });

  it("never labels stale data as live", () => {
    expect(formatFreshness(fixtureVehicles[2]!.freshness)).toContain("stale");
    expect(formatFreshness(fixtureVehicles[0]!.freshness)).toContain("Live");
  });
});
