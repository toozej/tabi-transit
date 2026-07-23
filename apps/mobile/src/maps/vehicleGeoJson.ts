import type { SyntheticVehicle } from "@/domain/vehicles";

export type VehicleFeatureCollection = GeoJSON.FeatureCollection<
  GeoJSON.Point,
  {
    id: string;
    mode: SyntheticVehicle["mode"];
    routeId: string;
    freshness: SyntheticVehicle["freshness"];
    bearing: number;
  }
>;

export function toVehicleFeatureCollection(
  vehicles: readonly SyntheticVehicle[],
): VehicleFeatureCollection {
  return {
    type: "FeatureCollection",
    features: vehicles.map((vehicle) => ({
      type: "Feature",
      id: vehicle.id,
      geometry: {
        type: "Point",
        coordinates: [...vehicle.coordinate],
      },
      properties: {
        id: vehicle.id,
        mode: vehicle.mode,
        routeId: vehicle.routeId,
        freshness: vehicle.freshness,
        bearing: vehicle.bearing,
      },
    })),
  };
}
