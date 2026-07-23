import { describe, expect, it } from "vitest";

import { getMapboxAccessToken } from "./config";

describe("getMapboxAccessToken", () => {
  it("does not enable maps for a missing or blank token", () => {
    expect(getMapboxAccessToken({})).toBeUndefined();
    expect(getMapboxAccessToken({ mapboxAccessToken: "  " })).toBeUndefined();
  });

  it("accepts a configured public token without logging it", () => {
    expect(getMapboxAccessToken({ mapboxAccessToken: "pk.test" })).toBe(
      "pk.test",
    );
  });
});
