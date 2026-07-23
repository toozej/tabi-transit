import { z } from "zod";

export const savedKindSchema = z.enum([
  "stop",
  "route",
  "vehicle",
  "place",
  "trip",
]);
export type SavedKind = z.infer<typeof savedKindSchema>;

export const savedItemInputSchema = z.object({
  id: z.string().min(1).max(160),
  // Labels are device-local display text only. Coordinates, search queries,
  // provider payloads, and push credentials are never part of this record.
  label: z.string().min(1).max(200),
  kind: savedKindSchema,
});

export const savedItemSchema = savedItemInputSchema.extend({
  savedAt: z.string().datetime(),
});
export const recentItemSchema = savedItemInputSchema.extend({
  openedAt: z.string().datetime(),
});
export const savedSnapshotSchema = z.object({
  saved: z.array(savedItemSchema),
  recents: z.array(recentItemSchema),
});

export type SavedItem = z.infer<typeof savedItemSchema>;
export type RecentItem = z.infer<typeof recentItemSchema>;
export type SavedSnapshot = z.infer<typeof savedSnapshotSchema>;

export const MAX_SAVED_ITEMS = 100;
export const MAX_RECENT_ITEMS = 20;

export function emptySavedSnapshot(): SavedSnapshot {
  return { saved: [], recents: [] };
}

/**
 * Retains only bounded, schema-valid local display records. Invalid persisted
 * values are ignored rather than being surfaced as a broken saved screen.
 */
export function sanitizeSavedSnapshot(value: unknown): SavedSnapshot {
  const parsed = savedSnapshotSchema.safeParse(value);
  if (!parsed.success) return emptySavedSnapshot();

  const uniqueSaved = new Map<string, SavedItem>();
  for (const item of parsed.data.saved) {
    if (!uniqueSaved.has(item.id)) uniqueSaved.set(item.id, item);
  }
  const uniqueRecents = new Map<string, RecentItem>();
  for (const item of parsed.data.recents) {
    if (!uniqueRecents.has(item.id)) uniqueRecents.set(item.id, item);
  }
  return {
    saved: [...uniqueSaved.values()]
      .sort((left, right) => right.savedAt.localeCompare(left.savedAt))
      .slice(0, MAX_SAVED_ITEMS),
    recents: [...uniqueRecents.values()]
      .sort((left, right) => right.openedAt.localeCompare(left.openedAt))
      .slice(0, MAX_RECENT_ITEMS),
  };
}
