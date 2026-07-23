import { create } from "zustand";

type MapUiState = {
  selectedVehicleId?: string;
  setSelectedVehicleId: (vehicleId?: string) => void;
};

/** UI-only state: server vehicle records stay in TanStack Query, not Zustand. */
export const useMapUiStore = create<MapUiState>((set) => ({
  selectedVehicleId: undefined,
  setSelectedVehicleId: (selectedVehicleId) => set({ selectedVehicleId }),
}));
