import type { Alert, Stop, Vehicle } from "./models";

const fresh = {
  source: "fixture-normalized",
  fetchedAt: "2026-07-22T16:30:02Z",
  status: "fresh" as const,
  ageSeconds: 4,
  isRealtime: true,
};
export const stops: Stop[] = [
  {
    id: "fixture:stop:101",
    name: "Burnside & 10th",
    modes: ["bus"],
    distanceMeters: 180,
    wheelchairAccessible: true,
  },
  {
    id: "fixture:stop:102",
    name: "Pioneer Square North",
    modes: ["light rail"],
    distanceMeters: 320,
    wheelchairAccessible: true,
  },
];
export const vehicles: Vehicle[] = [
  {
    id: "fixture:vehicle:2901",
    sourceVehicleId: "2901",
    mode: "bus",
    routeId: "fixture:route:20",
    headsign: "Gresham",
    coordinate: [-122.67, 45.52],
    bearing: 90,
    inService: true,
    freshness: fresh,
  },
  {
    id: "fixture:vehicle:77",
    sourceVehicleId: "77",
    mode: "light rail",
    routeId: "fixture:route:blue",
    headsign: "Hillsboro",
    coordinate: [-122.68, 45.53],
    bearing: 270,
    inService: true,
    freshness: fresh,
  },
  {
    id: "fixture:vehicle:900",
    sourceVehicleId: "900",
    mode: "streetcar",
    coordinate: [-122.66, 45.51],
    inService: false,
    freshness: { ...fresh, status: "stale", ageSeconds: 240 },
  },
];
export const alerts: Alert[] = [
  {
    id: "fixture:alert:1",
    header: "Minor delay",
    description: "Allow extra travel time.",
    severity: "warning",
    freshness: fresh,
  },
];
