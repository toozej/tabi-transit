import { useQuery } from "@tanstack/react-query";

import {
  riderInfoRepository,
  type NearbyCoordinate,
} from "@/data/api/riderInfoRepository";
import { getApiRuntimeConfig } from "@/data/api/config";

const AGENCY_TIME_ZONE = "America/Los_Angeles";

export function serviceDateInAgencyTimeZone(now = new Date()): string {
  const parts = new Intl.DateTimeFormat("en-US", {
    timeZone: AGENCY_TIME_ZONE,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(now);
  const value = Object.fromEntries(
    parts
      .filter((part) => part.type !== "literal")
      .map((part) => [part.type, part.value]),
  );
  return `${value.year}-${value.month}-${value.day}`;
}

export const riderInfoKeys = {
  nearby: (coordinate: NearbyCoordinate | undefined, limitPerMode?: number) =>
    [
      "nearby-stops",
      coordinate?.latitude,
      coordinate?.longitude,
      limitPerMode,
    ] as const,
  stop: (id: string) => ["stop", id] as const,
  arrivals: (stopId: string) => ["arrivals", stopId] as const,
  route: (id: string) => ["route", id] as const,
  routeShape: (id: string) => ["route-shape", id] as const,
  routeStops: (id: string, directionId?: 0 | 1) =>
    ["route-stops", id, directionId] as const,
  schedule: (id: string, serviceDate: string) =>
    ["schedule", id, serviceDate] as const,
  alerts: ["alerts"] as const,
  manifest: ["static-manifest"] as const,
};

export function useNearbyStops(
  coordinate?: NearbyCoordinate,
  limitPerMode = 2,
) {
  return useQuery({
    queryKey: riderInfoKeys.nearby(coordinate, limitPerMode),
    queryFn: () => riderInfoRepository.nearby(coordinate, limitPerMode),
    enabled:
      coordinate !== undefined || getApiRuntimeConfig().apiMode === "fixture",
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
export function useRouteStops(id: string, directionId?: 0 | 1) {
  return useQuery({
    queryKey: riderInfoKeys.routeStops(id, directionId),
    queryFn: () => riderInfoRepository.routeStops(id, directionId),
    enabled: Boolean(id) && directionId !== undefined,
    staleTime: Infinity,
  });
}
export function useSchedule(
  id: string,
  serviceDate = serviceDateInAgencyTimeZone(),
) {
  return useQuery({
    queryKey: riderInfoKeys.schedule(id, serviceDate),
    queryFn: () => riderInfoRepository.schedule(id, serviceDate),
    enabled: Boolean(id),
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
