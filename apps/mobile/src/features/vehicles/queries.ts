import { useQuery } from "@tanstack/react-query";
import { AppState } from "react-native";

import type { VehicleFilters } from "@/domain/vehicleModels";
import { vehicleRepository } from "@/data/api/vehicleRepository";

export const vehicleQueryKeys = {
  config: ["config"] as const,
  vehicles: (filters: VehicleFilters) => ["vehicles", filters] as const,
  search: (query: string) => ["vehicle-search", query] as const,
  detail: (id?: string) => ["vehicle", id] as const,
};

export function useVehicleConfig() {
  return useQuery({
    queryKey: vehicleQueryKeys.config,
    queryFn: () => vehicleRepository.config(),
  });
}

export function useVehicles(filters: VehicleFilters, pollSeconds = 15) {
  return useQuery({
    queryKey: vehicleQueryKeys.vehicles(filters),
    queryFn: () => vehicleRepository.vehicles(filters),
    refetchInterval: () =>
      AppState.currentState === "active" ? pollSeconds * 1_000 : false,
  });
}

export function useVehicleSearch(query: string) {
  return useQuery({
    queryKey: vehicleQueryKeys.search(query),
    queryFn: () => vehicleRepository.search(query),
    enabled: query.trim().length > 0,
  });
}

export function useVehicleDetail(id?: string) {
  return useQuery({
    queryKey: vehicleQueryKeys.detail(id),
    queryFn: () => vehicleRepository.vehicle(id ?? ""),
    enabled: Boolean(id),
  });
}
