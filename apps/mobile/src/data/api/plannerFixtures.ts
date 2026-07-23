import type { Itinerary, PlanEndpoint } from "@/domain/tripPlanning";

export const fixturePlanEndpoints: PlanEndpoint[] = [
  { id: "fixture:stop:101", label: "Burnside & 10th", kind: "stop" },
  { id: "fixture:stop:102", label: "Pioneer Square North", kind: "stop" },
];

export const fixtureItineraries: Itinerary[] = [
  {
    id: "fixture:itinerary:burnside-pioneer",
    departureAt: "2026-07-23T17:00:00Z",
    arrivalAt: "2026-07-23T17:23:00Z",
    durationSeconds: 1380,
    transfers: 0,
    walkingMeters: 220,
    wheelchairAccessible: true,
    source: "fixture-planner",
    freshness: {
      status: "fixture",
      message:
        "Fixture itinerary only. Trip planning is not connected to a provider.",
    },
    legs: [
      {
        id: "fixture:walk:1",
        mode: "walk",
        startLabel: "Burnside & 10th",
        endLabel: "West Burnside & 6th",
        startAt: "2026-07-23T17:00:00Z",
        endAt: "2026-07-23T17:04:00Z",
        durationSeconds: 240,
        walkingMeters: 220,
        realtime: "scheduled",
        geometry: [
          [-122.681, 45.522],
          [-122.679, 45.521],
        ],
      },
      {
        id: "fixture:transit:1",
        mode: "bus",
        startLabel: "West Burnside & 6th",
        endLabel: "Pioneer Square North",
        startAt: "2026-07-23T17:06:00Z",
        endAt: "2026-07-23T17:23:00Z",
        durationSeconds: 1020,
        routeLabel: "20 Burnside/Stark",
        headsign: "Gresham",
        realtime: "scheduled",
        geometry: [
          [-122.679, 45.521],
          [-122.677, 45.518],
        ],
      },
    ],
  },
];
