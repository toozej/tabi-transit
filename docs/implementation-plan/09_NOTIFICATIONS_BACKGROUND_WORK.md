---
doc_id: TAB-PLAN-009
title: "Notifications and Background Work"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["notifications-agent", "backend-agent", "mobile-agent"]
depends_on: ["TAB-PLAN-004", "TAB-PLAN-006", "TAB-PLAN-008"]
---


# Notifications and Background Work

## Principle

Mobile background execution cannot guarantee continuous vehicle/arrival monitoring. Backend workers evaluate conditions; mobile registers preferences and displays results.

## Registration

User selects notification → explain data/purpose → request OS permission → create/read SecureStore installation ID → obtain Expo push token → register platform/locale/time zone/app version → create subscription → mirror locally → handle token rotation. Never request permission on first launch.

## Subscription types

### Service alert

Scope route/stop/mode/source. Trigger on new, materially changed, and optionally ended alert. Dedup key includes subscription, alert, revision, and trigger.

### Departure reminder

One-shot itinerary or stop/route departure with lead time. Backend may recheck cancellation/major delay before sending.

### Arrival/vehicle threshold

Post-MVP and capped/short-lived. Prefer authoritative arrival estimates over GPS distance alone.

## Quiet hours/expiry

Store IANA time zone and validated local window. Default: no bypass. Drop stale arrival notices instead of queueing until morning. All watches expire and have per-installation limits.

## Push provider

Initial Expo Push Service through `expo-notifications`, behind an abstraction for direct APNs/FCM later. Batch sends, persist ticket IDs, process receipts, classify errors, disable invalid tokens, bounded retry, never retry expired time-sensitive messages.

Minimal payload contains notification type, entity ID, safe route/deep link, subscription ID, revision. Never exact location, credentials, or full itinerary.

## Worker

After normalized source changes: compare revision → indexed subscription selection → quiet/expiry → insert unique delivery → send → receipts/metrics. Transactions and unique keys prevent duplicates under concurrent workers.

## Mobile background tasks

Only deferrable work: refresh static manifest, prune cache, download approved update, reconcile subscriptions. Never continuous polling, guaranteed exact alerts, or location tracking. UI must not imply a fixed execution cadence.

Local notifications are acceptable for deterministic user reminders, with backend preferred for source-aware updates.

## Controls/abuse

Global/per-subscription settings, quiet hours, diagnostic status, delete all. Rate limits, max subscriptions, bounded windows, validated entity IDs, server-owned templates, encrypted/redacted tokens, and no public push relay.

## Acceptance

Contextual permission; token rotation and invalid cleanup; no duplicates; quiet hours/time-zone/expiry tests; no late arrival messages; installation deletion purges server data; safe validated deep links; push outage isolated; no dependence on continuous mobile background execution.
