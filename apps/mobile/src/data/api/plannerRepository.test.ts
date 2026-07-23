import { describe, expect, it } from "vitest";

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
    ).resolves.toHaveLength(1);
  });
  it("does not synthesize a plan from an incomplete draft", async () => {
    await expect(
      new PlannerRepository().plan({ constraints: defaultPlannerConstraints }),
    ).rejects.toBeInstanceOf(PlannerFeatureDisabledError);
  });
});
