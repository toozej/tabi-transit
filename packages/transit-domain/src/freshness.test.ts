import { describe, expect, it } from "vitest";

import { isFreshnessState } from "./freshness.js";

describe("isFreshnessState", () => {
  it("accepts API freshness states and rejects unknown values", () => {
    expect(isFreshnessState("fresh")).toBe(true);
    expect(isFreshnessState("late")).toBe(false);
  });
});
