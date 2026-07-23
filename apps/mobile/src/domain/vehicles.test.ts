import { describe, expect, it } from "vitest";

import { createSyntheticFleet } from "./vehicles";

describe("createSyntheticFleet", () => {
  it("creates a stable 1,500-point fleet with source-independent identifiers", () => {
    const fleet = createSyntheticFleet();

    expect(fleet).toHaveLength(1_500);
    expect(fleet[0]).toEqual({
      id: "synthetic-0001",
      mode: "bus",
      routeId: "BUS-1",
      coordinate: [-122.77, 45.45],
      bearing: 0,
      freshness: "stale",
    });
    expect(new Set(fleet.map((vehicle) => vehicle.id)).size).toBe(1_500);
  });
});
