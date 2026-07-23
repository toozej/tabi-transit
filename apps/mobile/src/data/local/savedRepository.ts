import {
  emptySavedSnapshot,
  sanitizeSavedSnapshot,
  type RecentItem,
  type SavedItem,
  type SavedSnapshot,
} from "@/domain/savedItems";

export interface SavedRepository {
  load(): Promise<SavedSnapshot>;
  replace(snapshot: SavedSnapshot): Promise<void>;
}

/** Safe deterministic fallback used for fixtures and when native storage fails. */
export class MemorySavedRepository implements SavedRepository {
  private snapshot: SavedSnapshot;

  constructor(initial: SavedSnapshot = emptySavedSnapshot()) {
    this.snapshot = sanitizeSavedSnapshot(initial);
  }

  async load(): Promise<SavedSnapshot> {
    return structuredClone(this.snapshot);
  }

  async replace(snapshot: SavedSnapshot): Promise<void> {
    this.snapshot = sanitizeSavedSnapshot(snapshot);
  }
}

type SavedRow = SavedItem;
type RecentRow = RecentItem & { position: number };
type SqlExecutor = {
  getAllAsync<T>(source: string): Promise<T[]>;
  runAsync(source: string, ...params: (string | number)[]): Promise<unknown>;
};
export type SavedSqliteDatabase = SqlExecutor & {
  withExclusiveTransactionAsync(
    task: (transaction: SqlExecutor) => Promise<void>,
  ): Promise<void>;
};

/**
 * Expo SQLite adapter. It is intentionally expressed in terms of a small
 * executor interface so its data semantics are unit-testable without native
 * SQLite. Device execution remains a separate evidence gate.
 */
export class SqliteSavedRepository implements SavedRepository {
  constructor(private readonly database: SavedSqliteDatabase) {}

  async load(): Promise<SavedSnapshot> {
    const [saved, recents] = await Promise.all([
      this.database.getAllAsync<SavedRow>(
        "SELECT id, label, kind, saved_at AS savedAt FROM saved_items ORDER BY saved_at DESC, id ASC",
      ),
      this.database.getAllAsync<RecentRow>(
        "SELECT id, label, kind, opened_at AS openedAt, position FROM recent_items ORDER BY position ASC",
      ),
    ]);
    return sanitizeSavedSnapshot({ saved, recents });
  }

  async replace(snapshot: SavedSnapshot): Promise<void> {
    const valid = sanitizeSavedSnapshot(snapshot);
    await this.database.withExclusiveTransactionAsync(async (transaction) => {
      await transaction.runAsync("DELETE FROM saved_items");
      await transaction.runAsync("DELETE FROM recent_items");
      for (const item of valid.saved) {
        await transaction.runAsync(
          "INSERT INTO saved_items (id, label, kind, saved_at) VALUES (?, ?, ?, ?)",
          item.id,
          item.label,
          item.kind,
          item.savedAt,
        );
      }
      for (const [position, item] of valid.recents.entries()) {
        await transaction.runAsync(
          "INSERT INTO recent_items (id, label, kind, opened_at, position) VALUES (?, ?, ?, ?, ?)",
          item.id,
          item.label,
          item.kind,
          item.openedAt,
          position,
        );
      }
    });
  }
}
