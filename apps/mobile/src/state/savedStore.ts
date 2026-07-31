import { create } from "zustand";

import {
  emptySavedSnapshot,
  MAX_RECENT_ITEMS,
  MAX_SAVED_ITEMS,
  sanitizeSavedSnapshot,
  savedItemInputSchema,
  type RecentItem,
  type SavedItem,
  type SavedSnapshot,
} from "@/domain/savedItems";
import {
  MemorySavedRepository,
  type SavedRepository,
} from "@/data/local/savedRepository";

export type { RecentItem, SavedItem, SavedKind } from "@/domain/savedItems";

type PersistenceState = "loading" | "ready" | "unavailable";
type NewSavedItem = Omit<SavedItem, "savedAt">;
type NewRecentItem = Omit<RecentItem, "openedAt">;

let repository: SavedRepository = new MemorySavedRepository();
let now = () => new Date().toISOString();
let mutationGeneration = 0;

function currentSnapshot(): SavedSnapshot {
  const { saved, recents } = useSavedStore.getState();
  return { saved, recents };
}

async function persistCurrent(): Promise<void> {
  try {
    await repository.replace(currentSnapshot());
    useSavedStore.setState({ persistence: "ready" });
  } catch {
    // Do not discard the in-memory choice. The UI accurately exposes that it
    // is session-only, and sensitive record contents never enter a log.
    useSavedStore.setState({ persistence: "unavailable" });
  }
}

type SavedState = {
  saved: SavedItem[];
  recents: RecentItem[];
  persistence: PersistenceState;
  hydrate: () => Promise<void>;
  markPersistenceUnavailable: () => void;
  toggleSaved: (item: NewSavedItem) => Promise<void>;
  addRecent: (item: NewRecentItem) => Promise<void>;
  clearRecents: () => Promise<void>;
  clearAllLocalData: () => Promise<void>;
};

export const useSavedStore = create<SavedState>((set) => ({
  ...emptySavedSnapshot(),
  persistence: "loading",
  hydrate: async () => {
    const generationAtStart = mutationGeneration;
    try {
      const snapshot = sanitizeSavedSnapshot(await repository.load());
      // Hydration is asynchronous. If the rider made a choice while SQLite
      // was opening, the in-memory state is newer than this snapshot and must
      // remain authoritative.
      if (mutationGeneration !== generationAtStart) return;
      set({ ...snapshot, persistence: "ready" });
    } catch {
      if (mutationGeneration !== generationAtStart) return;
      set({ persistence: "unavailable" });
    }
  },
  markPersistenceUnavailable: () => set({ persistence: "unavailable" }),
  toggleSaved: async (item) => {
    const parsed = savedItemInputSchema.safeParse(item);
    if (!parsed.success) return;
    mutationGeneration += 1;
    const timestamp = now();
    set((state) => {
      const existing = state.saved.some((saved) => saved.id === parsed.data.id);
      return {
        saved: existing
          ? state.saved.filter((saved) => saved.id !== parsed.data.id)
          : [{ ...parsed.data, savedAt: timestamp }, ...state.saved].slice(
              0,
              MAX_SAVED_ITEMS,
            ),
      };
    });
    await persistCurrent();
  },
  addRecent: async (item) => {
    const parsed = savedItemInputSchema.safeParse(item);
    if (!parsed.success) return;
    mutationGeneration += 1;
    const timestamp = now();
    set((state) => ({
      recents: [
        { ...parsed.data, openedAt: timestamp },
        ...state.recents.filter((recent) => recent.id !== parsed.data.id),
      ].slice(0, MAX_RECENT_ITEMS),
    }));
    await persistCurrent();
  },
  clearRecents: async () => {
    mutationGeneration += 1;
    set({ recents: [] });
    await persistCurrent();
  },
  clearAllLocalData: async () => {
    mutationGeneration += 1;
    set(emptySavedSnapshot());
    await persistCurrent();
  },
}));

/** Injectable only for application bootstrap and deterministic Vitest checks. */
export function configureSavedRepository(
  nextRepository: SavedRepository,
): void {
  repository = nextRepository;
}

export function configureSavedStoreClock(clock: () => string): void {
  now = clock;
}

export function resetSavedStoreForTest(): void {
  repository = new MemorySavedRepository();
  now = () => new Date().toISOString();
  mutationGeneration = 0;
  useSavedStore.setState({ ...emptySavedSnapshot(), persistence: "loading" });
}
