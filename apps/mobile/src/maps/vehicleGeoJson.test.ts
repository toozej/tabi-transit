import { describe, expect, it } from "vitest";

import { fixtureVehicles } from "@/data/api/fixtures";

import { toVehicleFeatureCollection } from "./vehicleGeoJson";

describe("toVehicleFeatureCollection", () => {
  it("keeps stable IDs and queryable layer properties", () => {
    const collection = toVehicleFeatureCollection(fixtureVehicles.slice(0, 2));

    expect(collection.features).toHaveLength(2);
    expect(collection.features[0]).toMatchObject({
      id: "fixture:vehicle:2901",
      geometry: { coordinates: [-122.67, 45.52] },
      properties: {
        id: "fixture:vehicle:2901",
        mode: "bus",
        freshness: "fresh",
      },
    });
  });
});
