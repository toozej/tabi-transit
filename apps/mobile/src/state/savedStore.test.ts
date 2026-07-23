import { beforeEach, describe, expect, it } from "vitest";
import { useSavedStore } from "./savedStore";

beforeEach(() => useSavedStore.setState({ saved: [], recents: [] }));
describe("saved and recent rider state", () => {
  it("toggles saved items without duplicating IDs", () => {
    useSavedStore
      .getState()
      .toggleSaved({ id: "fixture:stop:101", label: "Burnside", kind: "stop" });
    expect(useSavedStore.getState().saved).toHaveLength(1);
    useSavedStore
      .getState()
      .toggleSaved({ id: "fixture:stop:101", label: "Burnside", kind: "stop" });
    expect(useSavedStore.getState().saved).toHaveLength(0);
  });
  it("bounds recents and moves repeated selections to the front", () => {
    for (let index = 0; index < 21; index += 1)
      useSavedStore
        .getState()
        .addRecent({ id: `stop:${index}`, label: String(index), kind: "stop" });
    useSavedStore
      .getState()
      .addRecent({ id: "stop:5", label: "5", kind: "stop" });
    expect(useSavedStore.getState().recents).toHaveLength(20);
    expect(useSavedStore.getState().recents[0]?.id).toBe("stop:5");
  });
});
