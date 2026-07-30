import type { VehicleHistoryObservation } from "@/domain/vehicleModels";

export const HISTORY_UNAVAILABLE =
  "Vehicle history is unavailable. Try again later.";
export const HISTORY_EMPTY =
  "No retained observations are available for this vehicle.";
export const ADHERENCE_UNAVAILABLE =
  "Adherence is unavailable while the historical schedule evidence contract is reviewed.";

export function formatHistoryObservation(
  observation: VehicleHistoryObservation,
): string {
  const observedAt = new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "UTC",
  }).format(new Date(observation.observedAt));
  return `Observed ${observedAt} at ${observation.coordinate[1].toFixed(4)}, ${observation.coordinate[0].toFixed(4)}`;
}
