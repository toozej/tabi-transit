import { describe, expect, it } from "vitest";

import { createSyntheticFleet } from "@/domain/vehicles";

import { toVehicleFeatureCollection } from "./vehicleGeoJson";

describe("toVehicleFeatureCollection", () => {
  it("keeps stable IDs and queryable layer properties", () => {
    const collection = toVehicleFeatureCollection(createSyntheticFleet(2));

    expect(collection.features).toHaveLength(2);
    expect(collection.features[0]).toMatchObject({
      id: "synthetic-0001",
      geometry: { coordinates: [-122.77, 45.45] },
      properties: { id: "synthetic-0001", mode: "bus", freshness: "stale" },
    });
  });
});
