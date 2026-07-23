export type VehicleMode = "bus" | "max" | "wes" | "streetcar" | "aerial_tram";

export type SyntheticVehicle = {
  id: string;
  mode: VehicleMode;
  routeId: string;
  coordinate: readonly [longitude: number, latitude: number];
  bearing: number;
  freshness: "fresh" | "stale";
};

const MODES: readonly VehicleMode[] = [
  "bus",
  "max",
  "wes",
  "streetcar",
  "aerial_tram",
];

/** Deterministic synthetic fleet used only by the Phase 0 compatibility spike. */
export function createSyntheticFleet(count = 1_500): SyntheticVehicle[] {
  return Array.from({ length: count }, (_, index) => {
    const mode = MODES[index % MODES.length] ?? "bus";
    const row = Math.floor(index / 50);
    const column = index % 50;

    return {
      id: `synthetic-${String(index + 1).padStart(4, "0")}`,
      mode,
      routeId: `${mode.toUpperCase()}-${(index % 12) + 1}`,
      coordinate: [-122.77 + column * 0.003, 45.45 + row * 0.0025],
      bearing: (index * 37) % 360,
      freshness: index % 11 === 0 ? "stale" : "fresh",
    };
  });
}
