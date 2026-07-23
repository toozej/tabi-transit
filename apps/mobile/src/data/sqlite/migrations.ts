import type { SQLiteDatabase } from "expo-sqlite";

export type SqlExecutor = Pick<SQLiteDatabase, "execAsync" | "getFirstAsync">;

const CURRENT_SCHEMA_VERSION = 1;

const CREATE_METADATA_SQL = `
  CREATE TABLE IF NOT EXISTS metadata (
    key TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL
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
    await database.execAsync(
      `BEGIN IMMEDIATE;${CREATE_METADATA_SQL}PRAGMA user_version = 1;COMMIT;`,
    );
  }
}

export { CURRENT_SCHEMA_VERSION };
