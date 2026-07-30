import { describe, expect, it } from "vitest";

import { createExternalWalkingDirectionsLink } from "./externalDirections";

describe("external walking directions", () => {
  it("converts GeoJSON coordinates to an explicit walking maps link", () => {
    expect(
      createExternalWalkingDirectionsLink({
        mode: "walk",
        geometry: [
          [-122.681, 45.522],
          [-122.679, 45.521],
        ],
      }),
    ).toBe(
      "https://www.google.com/maps/dir/?api=1&origin=45.522%2C-122.681&destination=45.521%2C-122.679&travelmode=walking",
    );
  });

  it("never builds an external URL for transit or incomplete geometry", () => {
    expect(
      createExternalWalkingDirectionsLink({
        mode: "bus",
        geometry: [
          [-122.681, 45.522],
          [-122.679, 45.521],
        ],
      }),
    ).toBeUndefined();
    expect(
      createExternalWalkingDirectionsLink({
        mode: "walk",
        geometry: undefined,
      }),
    ).toBeUndefined();
  });

  it("rejects malformed coordinates rather than leaking an invalid link", () => {
    expect(() =>
      createExternalWalkingDirectionsLink({
        mode: "walk",
        geometry: [
          [-122.681, 45.522],
          [181, 45.521],
        ],
      }),
    ).toThrow("valid route coordinates");
  });
});
