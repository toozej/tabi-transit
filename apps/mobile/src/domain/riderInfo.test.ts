import { describe, expect, it } from "vitest";
import {
  applyNearbyLimit,
  formatDistance,
  serviceDayTime,
  type NearbyStops,
} from "./riderInfo";

const freshness = {
  source: "fixture",
  fetchedAt: "2026-07-22T16:30:02Z",
  processedAt: "2026-07-22T16:30:02Z",
  status: "fresh" as const,
  ageSeconds: 1,
  isRealtime: false,
};
const nearby: NearbyStops = {
  distanceType: "straight_line",
  freshness,
  groups: [
    {
      mode: "bus",
      stops: [
        {
          id: "a",
          name: "A",
          coordinate: [-122, 45],
          modes: ["bus"],
          distanceMeters: 400,
        },
        {
          id: "b",
          name: "B",
          coordinate: [-122, 45],
          modes: ["bus"],
          distanceMeters: 100,
        },
      ],
    },
    {
      mode: "light_rail",
      stops: [
        {
          id: "c",
          name: "C",
          coordinate: [-122, 45],
          modes: ["light_rail"],
          distanceMeters: 200,
        },
      ],
    },
  ],
};

describe("rider information domain", () => {
  it("limits independently for every mode group", () => {
    expect(
      applyNearbyLimit(nearby, 1).groups.map((group) =>
        group.stops.map((stop) => stop.id),
      ),
    ).toEqual([["b"], ["c"]]);
  });
  it("formats distance and service times beyond midnight", () => {
    expect(formatDistance(1250)).toBe("1.3 km away");
    expect(serviceDayTime(90_060)).toBe("25:01");
  });
});
