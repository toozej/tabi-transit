import type { SQLiteDatabase } from "expo-sqlite";

export type SqlExecutor = Pick<SQLiteDatabase, "execAsync" | "getFirstAsync">;

const CURRENT_SCHEMA_VERSION = 2;

const CREATE_METADATA_SQL = `
  CREATE TABLE IF NOT EXISTS metadata (
    key TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL
  );
`;

const CREATE_SAVED_ITEMS_SQL = `
  CREATE TABLE IF NOT EXISTS saved_items (
    id TEXT PRIMARY KEY NOT NULL,
    label TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('stop', 'route', 'vehicle', 'place', 'trip')),
    saved_at TEXT NOT NULL
  );
  CREATE TABLE IF NOT EXISTS recent_items (
    id TEXT PRIMARY KEY NOT NULL,
    label TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('stop', 'route', 'vehicle', 'place', 'trip')),
    opened_at TEXT NOT NULL,
    position INTEGER NOT NULL UNIQUE CHECK (position >= 0)
  );
`;

export async function migrateDatabase(database: SqlExecutor): Promise<void> {
  const versionRow = await database.getFirstAsync<{ user_version: number }>(
    "PRAGMA user_version",
  );
  const version = versionRow?.user_version ?? 0;

  if (version > CURRENT_SCHEMA_VERSION) {
    throw new Error(
      `Database schema version ${version} is newer than this application.`,
    );
  }

  if (version === 0) {
    await runMigration(
      database,
      `${CREATE_METADATA_SQL}${CREATE_SAVED_ITEMS_SQL}PRAGMA user_version = 2;`,
    );
    return;
  }

  if (version === 1) {
    await runMigration(
      database,
      `${CREATE_SAVED_ITEMS_SQL}PRAGMA user_version = 2;`,
    );
  }
}

async function runMigration(
  database: SqlExecutor,
  statements: string,
): Promise<void> {
  await database.execAsync("BEGIN IMMEDIATE");
  try {
    await database.execAsync(statements);
    await database.execAsync("COMMIT");
  } catch (error) {
    try {
      await database.execAsync("ROLLBACK");
    } catch {
      // Preserve the original migration failure. A subsequent bootstrap can
      // still surface an unavailable local store rather than partial schema.
    }
    throw error;
  }
}

export { CURRENT_SCHEMA_VERSION };
