import type { Vehicle } from "@/domain/vehicleModels";

export type VehicleFeatureCollection = GeoJSON.FeatureCollection<
  GeoJSON.Point,
  {
    id: string;
    mode: Vehicle["mode"];
    routeId?: string;
    freshness: Vehicle["freshness"]["status"];
    bearing?: number;
  }
>;

export function toVehicleFeatureCollection(
  vehicles: readonly Vehicle[],
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
        freshness: vehicle.freshness.status,
        bearing: vehicle.bearing,
      },
    })),
  };
}
