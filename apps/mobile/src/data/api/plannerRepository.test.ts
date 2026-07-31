import { describe, expect, it, vi } from "vitest";

import { defaultPlannerConstraints } from "@/domain/tripPlanning";

import {
  PlannerFeatureDisabledError,
  PlannerRepository,
} from "./plannerRepository";

describe("PlannerRepository", () => {
  it("returns deterministic fixture results only for complete drafts", async () => {
    await expect(
      new PlannerRepository().plan({
        origin: { id: "fixture:stop:101", label: "Origin", kind: "stop" },
        destination: {
          id: "fixture:stop:102",
          label: "Destination",
          kind: "stop",
        },
        constraints: defaultPlannerConstraints,
      }),
    ).resolves.toMatchObject([{ departureAt: defaultPlannerConstraints.time }]);
  });
  it("does not synthesize a plan from an incomplete draft", async () => {
    await expect(
      new PlannerRepository().plan({ constraints: defaultPlannerConstraints }),
    ).rejects.toBeInstanceOf(PlannerFeatureDisabledError);
  });

  it("uses the normalized Tabi journey endpoint in remote mode", async () => {
    const request = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          source: "trimet-web-services",
          freshness: { isRealtime: true },
          itineraries: [
            {
              id: "plan:1",
              durationSeconds: 600,
              transfers: 0,
              walkingMeters: 0,
              legs: [
                {
                  mode: "rail",
                  routeId: "MAX",
                  fromName: "Origin",
                  toName: "Destination",
                  startAt: "2026-07-23T17:00:00Z",
                  endAt: "2026-07-23T17:10:00Z",
                },
              ],
            },
          ],
        }),
      ),
    );
    const repository = new PlannerRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      request,
    );
    const draft = {
      origin: { id: "trimet:stop:101", label: "Origin", kind: "stop" as const },
      destination: {
        id: "trimet:stop:102",
        label: "Destination",
        kind: "stop" as const,
      },
      constraints: defaultPlannerConstraints,
    };

    await expect(repository.plan(draft)).resolves.toMatchObject([
      { source: "trimet-web-services", legs: [{ mode: "light_rail" }] },
    ]);
    expect(request).toHaveBeenCalledWith(
      "https://api.example.test/v1/journeys/plan",
      expect.objectContaining({
        method: "POST",
        body: expect.stringContaining('"stopId":"trimet:stop:101"'),
      }),
    );
  });

  it("keeps unsupported local endpoints out of the remote provider request", async () => {
    const request = vi.fn();
    const repository = new PlannerRepository(
      { apiMode: "remote", apiBaseUrl: "https://api.example.test" },
      request,
    );
    await expect(
      repository.plan({
        origin: { id: "local:pin:1", label: "Map pin", kind: "map_pin" },
        destination: {
          id: "trimet:stop:102",
          label: "Destination",
          kind: "stop",
        },
        constraints: defaultPlannerConstraints,
      }),
    ).rejects.toBeInstanceOf(PlannerFeatureDisabledError);
    expect(request).not.toHaveBeenCalled();
  });
});
