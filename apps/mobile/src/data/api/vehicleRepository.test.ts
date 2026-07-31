import { describe, expect, it, vi } from "vitest";

import { VehicleRepository } from "./vehicleRepository";

describe("VehicleRepository", () => {
  it("uses deterministic normalized fixtures by default", async () => {
    const repository = new VehicleRepository({ apiMode: "fixture" });

    await expect(repository.config()).resolves.toMatchObject({
      apiVersion: "0.1.0",
    });
    await expect(repository.vehicles()).resolves.toMatchObject({
      snapshotId: "fixture-snapshot-42",
    });
    await expect(repository.search("2901")).resolves.toMatchObject({
      vehicles: [{ id: "fixture:vehicle:2901" }],
    });
    const history = await repository.history("fixture:vehicle:2901");
    expect(history.vehicleId).toBe("fixture:vehicle:2901");
    expect(history.observations[0]?.observedAt).toBe("2026-07-22T16:30:02Z");
    await expect(
      repository.history("fixture:vehicle:unknown"),
    ).resolves.toMatchObject({
      observations: [],
    });
  });

  it("validates the remote normalized history response", async () => {
    const request = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          vehicleId: "trimet:vehicle:2901",
          observations: [],
          retentionDays: 30,
          freshness: {
            source: "normalized-vehicle-observations",
            status: "historical",
          },
        }),
      ),
    );
    const repository = new VehicleRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      request,
    );

    await expect(
      repository.history("trimet:vehicle:2901"),
    ).resolves.toMatchObject({
      vehicleId: "trimet:vehicle:2901",
    });
    expect(request).toHaveBeenCalledWith(
      "https://api.example.test/v1/vehicles/trimet%3Avehicle%3A2901/history",
      expect.anything(),
    );
  });

  it("sends and reuses an ETag for remote snapshots", async () => {
    const request = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            snapshotId: "remote-1",
            vehicles: [],
            freshness: {
              source: "fixture",
              fetchedAt: "2026-07-22T16:30:02Z",
              processedAt: "2026-07-22T16:30:02Z",
              status: "fresh",
              ageSeconds: 1,
              isRealtime: true,
            },
          }),
          { headers: { ETag: '"snapshot"' } },
        ),
      )
      .mockResolvedValueOnce(new Response(null, { status: 304 }));
    const repository = new VehicleRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      request,
    );

    const first = await repository.vehicles();
    await expect(repository.vehicles()).resolves.toEqual(first);
    expect(request.mock.calls[1]?.[1]?.headers).toEqual({
      "If-None-Match": '"snapshot"',
    });
  });

  it("does not convert a missing Retry-After header to an immediate retry", async () => {
    const repository = new VehicleRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      vi.fn().mockResolvedValue(new Response(null, { status: 503 })),
    );
    await expect(repository.vehicles()).rejects.toMatchObject({
      kind: "source_unavailable",
      retryAfterSeconds: undefined,
    });
  });
});
