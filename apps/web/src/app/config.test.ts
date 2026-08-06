import { describe, expect, it } from "vitest";

import { readConfig } from "./config";

describe("web configuration", () => {
  it("keeps the browser map disabled when no public token is configured", () => {
    expect(
      readConfig({
        VITE_MAPBOX_ACCESS_TOKEN: "   ",
        VITE_MAPBOX_STYLE_URL: "",
      }),
    ).toMatchObject({
      mapboxAccessToken: undefined,
      mapboxStyleUrl: "mapbox://styles/mapbox/light-v11",
    });
  });

  it("accepts a public token and an approved map style URL", () => {
    expect(
      readConfig({
        VITE_MAPBOX_ACCESS_TOKEN: "pk.test",
        VITE_MAPBOX_STYLE_URL: "mapbox://styles/tabi/light",
      }),
    ).toMatchObject({
      mapboxAccessToken: "pk.test",
      mapboxStyleUrl: "mapbox://styles/tabi/light",
    });
  });
});
