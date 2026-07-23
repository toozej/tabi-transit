export type FreshnessState = "fresh" | "stale" | "unknown";

export interface Freshness {
  fetchedAt: string;
  state: FreshnessState;
}

export function isFreshnessState(value: string): value is FreshnessState {
  return value === "fresh" || value === "stale" || value === "unknown";
}
