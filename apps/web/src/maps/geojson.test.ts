import { describe, expect, it } from "vitest";
import { vehicles } from "../features/fixtures";
import { selectedVehicleFeature, vehicleFeatureCollection } from "./geojson";

describe("vehicle layer inputs", () => {
  it("keeps the selected vehicle in a separate source", () => {
    expect(
      vehicleFeatureCollection(vehicles, vehicles[0]!.id).features,
    ).toHaveLength(2);
    expect(selectedVehicleFeature(vehicles[0]!)?.properties.id).toBe(
      vehicles[0]!.id,
    );
  });
});
