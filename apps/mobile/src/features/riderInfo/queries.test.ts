import { describe, expect, it } from "vitest";

import { serviceDateInAgencyTimeZone } from "./queries";

describe("rider information query dates", () => {
  it("uses the agency service date rather than the device UTC date", () => {
    expect(serviceDateInAgencyTimeZone(new Date("2026-07-24T06:30:00Z"))).toBe(
      "2026-07-23",
    );
  });
});
