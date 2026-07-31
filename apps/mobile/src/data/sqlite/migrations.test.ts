import { describe, expect, it } from "vitest";

import { migrateDatabase, type SqlExecutor } from "./migrations";

function executorAt(version: number): SqlExecutor & { statements: string[] } {
  const statements: string[] = [];
  return {
    statements,
    execAsync: async (statement: string) => {
      statements.push(statement);
    },
    getFirstAsync: async <T>() => ({ user_version: version }) as T,
  };
}

describe("migrateDatabase", () => {
  it("creates the initial schema in one transaction", async () => {
    const database = executorAt(0);

    await migrateDatabase(database);

    expect(database.statements).toEqual([
      "BEGIN IMMEDIATE",
      expect.any(String),
      "COMMIT",
    ]);
    expect(database.statements[1]).toContain(
      "CREATE TABLE IF NOT EXISTS metadata",
    );
    expect(database.statements[1]).toContain(
      "CREATE TABLE IF NOT EXISTS saved_items",
    );
    expect(database.statements[1]).toContain("PRAGMA user_version = 2");
  });

  it("refuses a database created by a newer application", async () => {
    await expect(migrateDatabase(executorAt(3))).rejects.toThrow(
      "newer than this application",
    );
  });

  it("adds bounded saved and recent records to a version-one database", async () => {
    const database = executorAt(1);
    await migrateDatabase(database);
    expect(database.statements[1]).toContain(
      "CREATE TABLE IF NOT EXISTS recent_items",
    );
    expect(database.statements[1]).toContain("PRAGMA user_version = 2");
  });

  it("rolls back a failed migration before rethrowing the failure", async () => {
    const database = executorAt(0);
    database.execAsync = async (statement) => {
      database.statements.push(statement);
      if (statement.includes("CREATE TABLE")) throw new Error("disk full");
    };

    await expect(migrateDatabase(database)).rejects.toThrow("disk full");
    expect(database.statements).toEqual([
      "BEGIN IMMEDIATE",
      expect.any(String),
      "ROLLBACK",
    ]);
  });
});
