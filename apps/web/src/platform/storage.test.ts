import { describe, expect, it } from "vitest";
import { createSessionRepository } from "./storage";

describe("saved repository", () => {
  it("does not retain malformed or sensitive persisted values", async () => {
    const values = new Map<string, string>();
    const storage = {
      getItem: (key: string) => values.get(key) ?? null,
      setItem: (key: string, value: string) => {
        values.set(key, value);
      },
    };
    storage.setItem(
      "tabi.web.saved.v1",
      '{"saved":[{"id":"45.5,-122.6","label":"location","kind":"place"}],"recents":[]}',
    );
    await expect(createSessionRepository(storage).read()).resolves.toEqual({
      saved: [],
      recents: [],
    });
  });
});
