import { describe, expect, it } from "vitest";

import { fixtureVehicleHistory } from "@/data/api/fixtures";

import {
  ADHERENCE_UNAVAILABLE,
  HISTORY_EMPTY,
  HISTORY_UNAVAILABLE,
  formatHistoryObservation,
} from "./vehicleHistoryPresentation";

describe("vehicle history presentation", () => {
  it("formats a textual timeline entry from normalized history only", () => {
    const observation =
      fixtureVehicleHistory["fixture:vehicle:2901"]!.observations[0]!;

    expect(formatHistoryObservation(observation)).toBe(
      "Observed Jul 22, 2026, 4:30 PM at 45.5200, -122.6700",
    );
  });

  it("has distinct unavailable, empty, and adherence-unavailable copy", () => {
    expect(HISTORY_UNAVAILABLE).toMatch(/unavailable/i);
    expect(HISTORY_EMPTY).toMatch(/No retained observations/i);
    expect(ADHERENCE_UNAVAILABLE).toMatch(/contract is reviewed/i);
  });
});
