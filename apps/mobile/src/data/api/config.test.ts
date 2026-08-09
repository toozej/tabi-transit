import { describe, expect, it } from "vitest";

import { getApiRuntimeConfig } from "./config";

describe("getApiRuntimeConfig", () => {
  it("accepts the empty API URL emitted by fixture mode", () => {
    expect(getApiRuntimeConfig({ apiMode: "fixture", apiBaseUrl: "" })).toEqual(
      { apiMode: "fixture", apiBaseUrl: undefined },
    );
  });

  it("requires a URL when remote mode is enabled", () => {
    expect(() =>
      getApiRuntimeConfig({ apiMode: "remote", apiBaseUrl: "" }),
    ).toThrow("A Tabi API base URL is required when remote mode is enabled.");
  });

  it("accepts a valid remote API URL", () => {
    expect(
      getApiRuntimeConfig({
        apiMode: "remote",
        apiBaseUrl: "https://api.example.test",
      }),
    ).toEqual({
      apiMode: "remote",
      apiBaseUrl: "https://api.example.test",
    });
  });
});
