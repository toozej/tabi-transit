import { describe, expect, it, vi } from "vitest";
import { RiderInfoRepository } from "./riderInfoRepository";

const fixtureRuntime = {
  apiMode: "fixture" as const,
  apiBaseUrl: "http://unused",
};
describe("RiderInfoRepository", () => {
  it("keeps the per-mode nearby limit independent", async () => {
    const repository = new RiderInfoRepository(fixtureRuntime, vi.fn());
    const result = await repository.nearby(1);
    expect(result.groups).toHaveLength(2);
    expect(result.groups.every((group) => group.stops.length <= 1)).toBe(true);
  });
  it("offers fixture schedules after midnight and a Mapbox-consumable route shape", async () => {
    const repository = new RiderInfoRepository(fixtureRuntime, vi.fn());
    expect(
      (await repository.schedule("fixture:stop:101")).schedule[0]
        ?.serviceDaySeconds,
    ).toBe(90_060);
    expect(
      (await repository.routeShape("fixture:route:20")).features[0]?.geometry
        .type,
    ).toBe("LineString");
  });
  it("validates remote nearby payloads at the app boundary", async () => {
    const repository = new RiderInfoRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            distanceType: "wrong",
            groups: [],
            freshness: {},
          }),
          { status: 200 },
        ),
      ),
    );
    await expect(repository.nearby()).rejects.toMatchObject({
      kind: "invalid_response",
    });
  });
});
