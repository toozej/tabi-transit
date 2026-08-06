import type { Vehicle } from "../features/models";

export function vehicleFeatureCollection(
  vehicles: readonly Vehicle[],
  selectedId?: string,
) {
  return {
    type: "FeatureCollection" as const,
    features: vehicles
      .filter((vehicle) => vehicle.id !== selectedId)
      .map((vehicle) => ({
        type: "Feature" as const,
        geometry: { type: "Point" as const, coordinates: vehicle.coordinate },
        properties: {
          id: vehicle.id,
          mode: vehicle.mode,
          freshness: vehicle.freshness.status,
          bearing: vehicle.bearing,
        },
      })),
  };
}
export function selectedVehicleFeature(vehicle: Vehicle | undefined) {
  return vehicle
    ? {
        type: "Feature" as const,
        geometry: { type: "Point" as const, coordinates: vehicle.coordinate },
        properties: { id: vehicle.id },
      }
    : undefined;
}
