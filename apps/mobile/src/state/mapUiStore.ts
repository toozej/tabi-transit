import { create } from "zustand";

import type { VehicleFilters, VehicleMode } from "@/domain/vehicleModels";

type MapUiState = {
  selectedVehicleId?: string;
  filters: VehicleFilters;
  setSelectedVehicleId: (vehicleId?: string) => void;
  setModeEnabled: (mode: VehicleMode, enabled: boolean) => void;
  setRouteId: (routeId?: string) => void;
  setFreshness: (freshness?: VehicleFilters["freshness"]) => void;
};

/** UI-only state: server vehicle records stay in TanStack Query, not Zustand. */
export const useMapUiStore = create<MapUiState>((set) => ({
  selectedVehicleId: undefined,
  filters: { modes: [] },
  setSelectedVehicleId: (selectedVehicleId) => set({ selectedVehicleId }),
  setModeEnabled: (mode, enabled) =>
    set((state) => ({
      filters: {
        ...state.filters,
        modes: enabled
          ? [...new Set([...state.filters.modes, mode])]
          : state.filters.modes.filter((item) => item !== mode),
      },
    })),
  setRouteId: (routeId) =>
    set((state) => ({ filters: { ...state.filters, routeId } })),
  setFreshness: (freshness) =>
    set((state) => ({ filters: { ...state.filters, freshness } })),
}));
