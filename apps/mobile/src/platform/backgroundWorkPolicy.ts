export type BackgroundWorkKind =
  | "refresh_static_manifest"
  | "prune_public_cache"
  | "download_approved_static_update"
  | "reconcile_subscriptions"
  | "vehicle_polling"
  | "location_monitoring";

const permittedBackgroundWork = new Set<BackgroundWorkKind>([
  "refresh_static_manifest",
  "prune_public_cache",
  "download_approved_static_update",
  "reconcile_subscriptions",
]);

/** A policy boundary: device scheduling is intentionally not implemented here. */
export function mayScheduleBackgroundWork(kind: BackgroundWorkKind): boolean {
  return permittedBackgroundWork.has(kind);
}
