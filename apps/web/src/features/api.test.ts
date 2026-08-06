import { describe, expect, it } from "vitest";

import {
  collectionAlerts,
  collectionVehicles,
  flattenNearbyGroups,
  planJourney,
} from "./api";

const freshness = {
  source: "fixture",
  fetchedAt: "2026-07-22T16:30:02Z",
  status: "fresh",
  ageSeconds: 1,
  isRealtime: true,
} as const;

describe("remote API response adapters", () => {
  it("flattens the normalized nearby group envelope", () => {
    expect(
      flattenNearbyGroups({
        groups: [
          {
            mode: "bus",
            stops: [
              {
                id: "fixture:stop:1",
                name: "Fixture",
                modes: ["bus"],
                distanceMeters: 10,
              },
            ],
          },
        ],
        freshness,
      }),
    ).toMatchObject([{ id: "fixture:stop:1", distanceMeters: 10 }]);
  });

  it("reads vehicle and alert collections rather than assuming arrays", () => {
    expect(
      collectionVehicles({
        vehicles: [
          {
            id: "fixture:vehicle:1",
            sourceVehicleId: "1",
            mode: "bus",
            coordinate: [-122.67, 45.52],
            inService: true,
            freshness,
          },
        ],
        freshness,
      }),
    ).toHaveLength(1);
    expect(
      collectionAlerts({
        alerts: [{ id: "fixture:alert:1", header: "Delay", freshness }],
        freshness,
      }),
    ).toHaveLength(1);
  });

  it("uses opaque stop IDs for fixture planning", async () => {
    await expect(
      planJourney("fixture:stop:101", "fixture:stop:102"),
    ).resolves.toMatchObject({ source: "fixture-planner" });
  });
});
