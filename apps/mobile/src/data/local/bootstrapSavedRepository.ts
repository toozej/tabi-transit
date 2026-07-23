import * as SQLite from "expo-sqlite";

import { migrateDatabase } from "@/data/sqlite/migrations";
import { configureSavedRepository, useSavedStore } from "@/state/savedStore";

import {
  MemorySavedRepository,
  SqliteSavedRepository,
} from "./savedRepository";

/**
 * Starts local saved-data storage without making the app depend on it. A
 * failure leaves an explicit session-only fallback and deliberately logs no
 * saved labels or user data.
 */
export async function bootstrapSavedRepository(): Promise<void> {
  try {
    const database = await SQLite.openDatabaseAsync("tabi-local.db");
    await migrateDatabase(database);
    const repository = new SqliteSavedRepository(database);
    configureSavedRepository(repository);
    await useSavedStore.getState().hydrate();
  } catch {
    configureSavedRepository(new MemorySavedRepository());
    useSavedStore.getState().markPersistenceUnavailable();
  }
}
