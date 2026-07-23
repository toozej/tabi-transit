import { describe, expect, it } from "vitest";

import {
  SqliteSavedRepository,
  type SavedSqliteDatabase,
} from "./savedRepository";

function database(): SavedSqliteDatabase & { statements: string[] } {
  const statements: string[] = [];
  const executor: SavedSqliteDatabase & { statements: string[] } = {
    statements,
    getAllAsync: async <T>() => [] as T[],
    runAsync: async (statement: string) => {
      statements.push(statement);
    },
    withExclusiveTransactionAsync: async (task) => task(executor),
  };
  return executor;
}

describe("SqliteSavedRepository", () => {
  it("replaces saved and recent data in an exclusive transaction", async () => {
    const sqlite = database();
    const repository = new SqliteSavedRepository(sqlite);

    await repository.replace({
      saved: [
        {
          id: "fixture:stop:101",
          label: "Burnside",
          kind: "stop",
          savedAt: "2026-07-23T00:00:00.000Z",
        },
      ],
      recents: [
        {
          id: "fixture:route:20",
          label: "20",
          kind: "route",
          openedAt: "2026-07-23T00:01:00.000Z",
        },
      ],
    });

    expect(sqlite.statements).toEqual([
      "DELETE FROM saved_items",
      "DELETE FROM recent_items",
      "INSERT INTO saved_items (id, label, kind, saved_at) VALUES (?, ?, ?, ?)",
      "INSERT INTO recent_items (id, label, kind, opened_at, position) VALUES (?, ?, ?, ?, ?)",
    ]);
  });
});
