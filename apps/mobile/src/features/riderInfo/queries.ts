import { useQuery } from "@tanstack/react-query";

import { riderInfoRepository } from "@/data/api/riderInfoRepository";

export const riderInfoKeys = {
  nearby: (limitPerMode?: number) => ["nearby-stops", limitPerMode] as const,
  stop: (id: string) => ["stop", id] as const,
  arrivals: (stopId: string) => ["arrivals", stopId] as const,
  route: (id: string) => ["route", id] as const,
  routeShape: (id: string) => ["route-shape", id] as const,
  schedule: (id: string) => ["schedule", id] as const,
  alerts: ["alerts"] as const,
  manifest: ["static-manifest"] as const,
};

export function useNearbyStops(limitPerMode = 2) {
  return useQuery({
    queryKey: riderInfoKeys.nearby(limitPerMode),
    queryFn: () => riderInfoRepository.nearby(limitPerMode),
  });
}
export function useStop(id: string) {
  return useQuery({
    queryKey: riderInfoKeys.stop(id),
    queryFn: () => riderInfoRepository.stop(id),
    enabled: Boolean(id),
  });
}
export function useArrivals(stopId: string) {
  return useQuery({
    queryKey: riderInfoKeys.arrivals(stopId),
    queryFn: () => riderInfoRepository.arrivals(stopId),
    enabled: Boolean(stopId),
  });
}
export function useRoute(id: string) {
  return useQuery({
    queryKey: riderInfoKeys.route(id),
    queryFn: () => riderInfoRepository.route(id),
    enabled: Boolean(id),
  });
}
export function useRouteShape(id: string) {
  return useQuery({
    queryKey: riderInfoKeys.routeShape(id),
    queryFn: () => riderInfoRepository.routeShape(id),
    enabled: Boolean(id),
    staleTime: Infinity,
  });
}
export function useSchedule(id: string) {
  return useQuery({
    queryKey: riderInfoKeys.schedule(id),
    queryFn: () => riderInfoRepository.schedule(id),
    enabled: Boolean(id),
    staleTime: Infinity,
  });
}
export function useAlerts() {
  return useQuery({
    queryKey: riderInfoKeys.alerts,
    queryFn: () => riderInfoRepository.alerts(),
  });
}
export function useStaticManifest() {
  return useQuery({
    queryKey: riderInfoKeys.manifest,
    queryFn: () => riderInfoRepository.manifest(),
    staleTime: Infinity,
  });
}
