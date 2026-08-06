import { z } from "zod";

const itemSchema = z.object({
  id: z
    .string()
    .min(1)
    .max(160)
    .regex(/^[A-Za-z0-9:_-]+$/),
  label: z.string().min(1).max(160),
  kind: z.enum(["stop", "route", "vehicle", "place"]),
});
export type SavedItem = z.infer<typeof itemSchema>;
const stateSchema = z.object({
  saved: z.array(itemSchema).max(100),
  recents: z.array(itemSchema).max(30),
});
const key = "tabi.web.saved.v1";
type SessionStorage = Pick<Storage, "getItem" | "setItem">;

export interface SavedRepository {
  read(): Promise<{ saved: SavedItem[]; recents: SavedItem[] }>;
  write(value: { saved: SavedItem[]; recents: SavedItem[] }): Promise<void>;
  clear(kind: "recents" | "all"): Promise<void>;
}

/** Browser storage is untrusted and contains only non-sensitive rider choices. */
export function createSessionRepository(
  storage: SessionStorage | undefined = typeof sessionStorage === "undefined"
    ? undefined
    : sessionStorage,
): SavedRepository {
  const read = async () => {
    try {
      const serialized = storage?.getItem(key);
      return (
        stateSchema.safeParse(serialized ? JSON.parse(serialized) : {})
          .data ?? { saved: [], recents: [] }
      );
    } catch {
      return { saved: [], recents: [] };
    }
  };
  return {
    read,
    async write(value) {
      if (storage)
        storage.setItem(key, JSON.stringify(stateSchema.parse(value)));
    },
    async clear(kind) {
      const value = await read();
      if (storage)
        storage.setItem(
          key,
          JSON.stringify(
            kind === "recents"
              ? { ...value, recents: [] }
              : { saved: [], recents: [] },
          ),
        );
    },
  };
}

function requestResult<T>(request: IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    request.onsuccess = () => resolve(request.result);
    request.onerror = () =>
      reject(request.error ?? new Error("IndexedDB request failed."));
  });
}

function openDatabase(factory: IDBFactory): Promise<IDBDatabase> {
  const request = factory.open("tabi-web", 1);
  request.onupgradeneeded = () => {
    if (!request.result.objectStoreNames.contains("rider-state")) {
      request.result.createObjectStore("rider-state");
    }
  };
  return requestResult(request);
}

/** IndexedDB persists only validated saved/recents data, never location or search text. */
export function createIndexedDbRepository(
  factory: IDBFactory,
): SavedRepository {
  const database = openDatabase(factory);
  async function transaction<T>(
    mode: IDBTransactionMode,
    action: (store: IDBObjectStore) => IDBRequest<T>,
  ): Promise<T> {
    const db = await database;
    return requestResult(
      action(db.transaction("rider-state", mode).objectStore("rider-state")),
    );
  }
  return {
    async read() {
      const value = await transaction("readonly", (store) => store.get(key));
      return (
        stateSchema.safeParse(value ?? {}).data ?? { saved: [], recents: [] }
      );
    },
    async write(value) {
      await transaction("readwrite", (store) =>
        store.put(stateSchema.parse(value), key),
      );
    },
    async clear(kind) {
      const current = await this.read();
      await this.write(
        kind === "recents"
          ? { ...current, recents: [] }
          : { saved: [], recents: [] },
      );
    },
  };
}

/** Falls back explicitly to the session repository when browser durable storage is denied. */
export function createBrowserRepository(): SavedRepository {
  const session = createSessionRepository();
  const durable =
    typeof indexedDB === "undefined"
      ? undefined
      : createIndexedDbRepository(indexedDB);
  if (!durable) return session;
  return {
    async read() {
      try {
        return await durable.read();
      } catch {
        return session.read();
      }
    },
    async write(value) {
      try {
        await durable.write(value);
      } catch {
        await session.write(value);
      }
    },
    async clear(kind) {
      try {
        await durable.clear(kind);
      } catch {
        await session.clear(kind);
      }
    },
  };
}
