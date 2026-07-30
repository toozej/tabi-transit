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
  it("returns only the selected direction's ordered route-stop collection in fixture mode", async () => {
    const repository = new RiderInfoRepository(fixtureRuntime, vi.fn());
    const routeStops = await repository.routeStops("fixture:route:20", 1);
    expect(routeStops.directionId).toBe(1);
    expect(routeStops.stops.map((stop) => stop.id)).toEqual([
      "fixture:stop:103",
      "fixture:stop:101",
    ]);
    await expect(
      repository.routeStops("fixture:route:unknown", 0),
    ).rejects.toMatchObject({ kind: "http" });
  });
  it("validates the remote route-stop response and encodes opaque route IDs", async () => {
    const request = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          routeId: "trimet:route:20",
          directionId: 1,
          stops: [
            {
              id: "trimet:stop:101",
              name: "Burnside & 10th",
              coordinate: [-122.681, 45.522],
              modes: ["bus"],
              sequence: 1,
            },
          ],
          staticFeedVersion: "v1",
        }),
        { status: 200 },
      ),
    );
    const repository = new RiderInfoRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      request,
    );
    await expect(
      repository.routeStops("trimet:route:20", 1),
    ).resolves.toMatchObject({
      routeId: "trimet:route:20",
    });
    expect(request).toHaveBeenCalledWith(
      "https://api.example.test/v1/routes/trimet%3Aroute%3A20/stops?directionId=1",
    );
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
