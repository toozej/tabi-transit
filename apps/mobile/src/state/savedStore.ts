import { create } from "zustand";

export type SavedKind = "stop" | "route" | "vehicle";
export type SavedItem = {
  id: string;
  label: string;
  kind: SavedKind;
  savedAt: string;
};
export type RecentItem = {
  id: string;
  label: string;
  kind: SavedKind;
  openedAt: string;
};

const MAX_RECENTS = 20;
type SavedState = {
  saved: SavedItem[];
  recents: RecentItem[];
  toggleSaved: (item: Omit<SavedItem, "savedAt">) => void;
  addRecent: (item: Omit<RecentItem, "openedAt">) => void;
  clearRecents: () => void;
};
/** Device-local UI state only; SQLite persistence is a later migration boundary. */
export const useSavedStore = create<SavedState>((set) => ({
  saved: [],
  recents: [],
  toggleSaved: (item) =>
    set((state) => ({
      saved: state.saved.some((saved) => saved.id === item.id)
        ? state.saved.filter((saved) => saved.id !== item.id)
        : [...state.saved, { ...item, savedAt: new Date().toISOString() }],
    })),
  addRecent: (item) =>
    set((state) => ({
      recents: [
        { ...item, openedAt: new Date().toISOString() },
        ...state.recents.filter((recent) => recent.id !== item.id),
      ].slice(0, MAX_RECENTS),
    })),
  clearRecents: () => set({ recents: [] }),
}));
