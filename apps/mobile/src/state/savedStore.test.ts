import { beforeEach, describe, expect, it } from "vitest";
import {
  configureSavedRepository,
  configureSavedStoreClock,
  resetSavedStoreForTest,
  useSavedStore,
} from "./savedStore";
import {
  MemorySavedRepository,
  type SavedRepository,
} from "@/data/local/savedRepository";

beforeEach(() => {
  resetSavedStoreForTest();
  configureSavedStoreClock(() => "2026-07-23T00:00:00.000Z");
});
describe("saved and recent rider state", () => {
  it("toggles saved items without duplicating IDs", async () => {
    await useSavedStore
      .getState()
      .toggleSaved({ id: "fixture:stop:101", label: "Burnside", kind: "stop" });
    expect(useSavedStore.getState().saved).toHaveLength(1);
    await useSavedStore
      .getState()
      .toggleSaved({ id: "fixture:stop:101", label: "Burnside", kind: "stop" });
    expect(useSavedStore.getState().saved).toHaveLength(0);
  });
  it("bounds recents and moves repeated selections to the front", async () => {
    for (let index = 0; index < 21; index += 1)
      await useSavedStore
        .getState()
        .addRecent({ id: `stop:${index}`, label: String(index), kind: "stop" });
    await useSavedStore
      .getState()
      .addRecent({ id: "stop:5", label: "5", kind: "stop" });
    expect(useSavedStore.getState().recents).toHaveLength(20);
    expect(useSavedStore.getState().recents[0]?.id).toBe("stop:5");
  });

  it("validates local records and evicts favorites at the privacy-safe limit", async () => {
    for (let index = 0; index < 101; index += 1) {
      await useSavedStore.getState().toggleSaved({
        id: `stop:${index}`,
        label: String(index),
        kind: "stop",
      });
    }
    await useSavedStore
      .getState()
      .addRecent({ id: "", label: "invalid", kind: "stop" } as never);

    expect(useSavedStore.getState().saved).toHaveLength(100);
    expect(
      useSavedStore.getState().saved.some((item) => item.id === "stop:0"),
    ).toBe(false);
    expect(useSavedStore.getState().recents).toEqual([]);
  });

  it("recovers durable local favorites and recents through the repository boundary", async () => {
    const repository = new MemorySavedRepository({
      saved: [
        {
          id: "fixture:route:20",
          label: "20 Burnside/Stark",
          kind: "route",
          savedAt: "2026-07-22T00:00:00.000Z",
        },
      ],
      recents: [
        {
          id: "fixture:stop:101",
          label: "Burnside & 10th",
          kind: "stop",
          openedAt: "2026-07-23T00:00:00.000Z",
        },
      ],
    });
    configureSavedRepository(repository);

    await useSavedStore.getState().hydrate();

    expect(useSavedStore.getState()).toMatchObject({
      persistence: "ready",
      saved: [{ id: "fixture:route:20" }],
      recents: [{ id: "fixture:stop:101" }],
    });
  });

  it("keeps choices in-session and marks storage unavailable when writing fails", async () => {
    const failingRepository: SavedRepository = {
      load: async () => ({ saved: [], recents: [] }),
      replace: async () => Promise.reject(new Error("storage unavailable")),
    };
    configureSavedRepository(failingRepository);

    await useSavedStore
      .getState()
      .toggleSaved({ id: "fixture:stop:101", label: "Burnside", kind: "stop" });

    expect(useSavedStore.getState()).toMatchObject({
      persistence: "unavailable",
      saved: [{ id: "fixture:stop:101" }],
    });
  });

  it("clears all local user records and persists the deletion", async () => {
    const repository = new MemorySavedRepository();
    configureSavedRepository(repository);
    await useSavedStore
      .getState()
      .toggleSaved({ id: "fixture:trip:1", label: "Trip", kind: "trip" });
    await useSavedStore.getState().addRecent({
      id: "fixture:place:1",
      label: "Saved place",
      kind: "place",
    });

    await useSavedStore.getState().clearAllLocalData();

    await expect(repository.load()).resolves.toEqual({
      saved: [],
      recents: [],
    });
  });
});
