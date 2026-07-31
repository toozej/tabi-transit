import { describe, expect, it, vi } from "vitest";
import { RiderInfoRepository } from "./riderInfoRepository";

const fixtureRuntime = {
  apiMode: "fixture" as const,
  apiBaseUrl: "http://unused",
};
describe("RiderInfoRepository", () => {
  it("keeps the per-mode nearby limit independent", async () => {
    const repository = new RiderInfoRepository(fixtureRuntime, vi.fn());
    const result = await repository.nearby(undefined, 1);
    expect(result.groups).toHaveLength(2);
    expect(result.groups.every((group) => group.stops.length <= 1)).toBe(true);
  });
  it("offers fixture schedules after midnight and a Mapbox-consumable route shape", async () => {
    const repository = new RiderInfoRepository(fixtureRuntime, vi.fn());
    expect(
      (await repository.schedule("fixture:stop:101", "2026-07-23")).schedule[0]
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
    await expect(
      repository.nearby({ latitude: 45.52, longitude: -122.68 }),
    ).rejects.toMatchObject({
      kind: "invalid_response",
    });
  });

  it("sends the caller's coordinate and contract parameter names for nearby stops", async () => {
    const request = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          distanceType: "straight_line",
          groups: [],
          freshness: {
            source: "static",
            fetchedAt: "2026-07-23T00:00:00Z",
            processedAt: "2026-07-23T00:00:00Z",
            status: "fresh",
            ageSeconds: 1,
            isRealtime: false,
          },
        }),
      ),
    );
    const repository = new RiderInfoRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      request,
    );

    await repository.nearby({ latitude: 47.6062, longitude: -122.3321 }, 2);
    expect(request).toHaveBeenCalledWith(
      "https://api.example.test/v1/stops/nearby?lat=47.6062&lon=-122.3321&limitPerMode=2",
    );
  });

  it("does not send invalid nearby coordinates to the API", async () => {
    const request = vi.fn();
    const repository = new RiderInfoRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      request,
    );
    await expect(
      repository.nearby({ latitude: Number.NaN, longitude: -122.68 }),
    ).rejects.toMatchObject({ kind: "http" });
    expect(request).not.toHaveBeenCalled();
  });

  it("includes the required service date and rejects unrelated fixture resources", async () => {
    const request = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          stopId: "x",
          serviceDate: "2026-07-23",
          staticFeedVersion: "v1",
          schedule: [],
        }),
      ),
    );
    const remote = new RiderInfoRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      request,
    );
    await remote.schedule("trimet:stop:1", "2026-07-23");
    expect(request).toHaveBeenCalledWith(
      "https://api.example.test/v1/stops/trimet%3Astop%3A1/schedule?serviceDate=2026-07-23",
    );

    const fixture = new RiderInfoRepository(fixtureRuntime, vi.fn());
    await expect(fixture.route("fixture:route:other")).rejects.toMatchObject({
      kind: "http",
    });
    await expect(
      fixture.schedule("fixture:stop:other", "2026-07-23"),
    ).rejects.toMatchObject({ kind: "http" });
  });

  it("reports malformed successful JSON as an invalid response, not offline", async () => {
    const repository = new RiderInfoRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      vi.fn().mockResolvedValue(
        new Response("{", {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );
    await expect(repository.stop("trimet:stop:1")).rejects.toMatchObject({
      kind: "invalid_response",
    });
  });
});
