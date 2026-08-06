import { describe, expect, it } from "vitest";

import { parsePlannerSearch, validOpaqueId } from "./urls";

describe("shareable URL validation", () => {
  it("accepts opaque planner values", () => {
    expect(
      parsePlannerSearch("?origin=fixture%3Astop%3A101&maxTransfers=2").data,
    ).toMatchObject({
      origin: "fixture:stop:101",
      maxTransfers: 2,
    });
  });

  it("rejects coordinates, raw text, and malformed IDs", () => {
    expect(validOpaqueId("45.52,-122.67")).toBeUndefined();
    expect(validOpaqueId("Burnside & 10th")).toBeUndefined();
    expect(parsePlannerSearch("?origin=45.52%2C-122.67").success).toBe(false);
  });
});
