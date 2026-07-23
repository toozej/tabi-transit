import { describe, expect, it } from "vitest";

import {
  createPlanDeepLink,
  defaultPlannerConstraints,
  parsePlanDeepLink,
  rankItineraries,
  type Itinerary,
} from "./tripPlanning";

const itinerary: Itinerary = {
  id: "fixture:itinerary:1",
  departureAt: "2026-07-23T17:00:00Z",
  arrivalAt: "2026-07-23T17:30:00Z",
  durationSeconds: 1800,
  transfers: 0,
  walkingMeters: 200,
  wheelchairAccessible: true,
  source: "fixture-planner",
  freshness: { status: "fixture", message: "Synthetic planning result." },
  legs: [
    {
      id: "walk",
      mode: "walk",
      startLabel: "Origin",
      endLabel: "Stop",
      startAt: "2026-07-23T17:00:00Z",
      endAt: "2026-07-23T17:05:00Z",
      durationSeconds: 300,
      walkingMeters: 200,
      realtime: "scheduled",
    },
  ],
};

describe("trip planning domain", () => {
  it("round-trips opaque endpoint IDs without a coordinate", () => {
    const link = createPlanDeepLink({
      origin: { id: "fixture:stop:101", label: "Origin", kind: "stop" },
      destination: {
        id: "fixture:stop:102",
        label: "Destination",
        kind: "stop",
      },
      constraints: { ...defaultPlannerConstraints, maxTransfers: 1 },
    });
    expect(link).not.toContain("45.");
    expect(parsePlanDeepLink(link)).toEqual({
      origin: "fixture:stop:101",
      destination: "fixture:stop:102",
      maxTransfers: 1,
      accessibility: undefined,
    });
  });

  it("rejects non-plan and malformed links", () => {
    expect(() => parsePlanDeepLink("https://tabi.example/plan")).toThrow();
    expect(() => parsePlanDeepLink("tabi://plan?origin=not allowed")).toThrow();
  });

  it("enforces hard constraints and explains an empty result", () => {
    expect(
      rankItineraries([itinerary], {
        ...defaultPlannerConstraints,
        maxWalkingMeters: 100,
      }),
    ).toMatchObject({ itineraries: [], disclosure: expect.any(String) });
  });
});
