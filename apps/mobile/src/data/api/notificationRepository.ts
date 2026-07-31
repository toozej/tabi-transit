import {
  type NotificationSubscription,
  type NotificationSubscriptionDraft,
  validateSubscriptionDraft,
} from "@/domain/notifications";

import { fixtureNotificationSubscriptions } from "./notificationFixtures";

export type NotificationRegistrationStatus = {
  available: false;
  reason: "device_push_disabled";
  message: string;
};

/** Platform push boundary. This fixture implementation never requests permissions or tokens. */
export interface NotificationRegistrationAdapter {
  status(): Promise<NotificationRegistrationStatus>;
}

export const unavailableNotificationRegistration: NotificationRegistrationAdapter =
  {
    async status() {
      return {
        available: false,
        reason: "device_push_disabled",
        message:
          "Push registration is unavailable in this fixture build. Tabi has not requested notification permission.",
      };
    },
  };

export interface NotificationRepository {
  list(): Promise<NotificationSubscription[]>;
  create(
    draft: NotificationSubscriptionDraft,
  ): Promise<NotificationSubscription>;
  remove(id: string): Promise<void>;
}

export function createFixtureNotificationRepository(
  initial: NotificationSubscription[] = fixtureNotificationSubscriptions,
  now: () => Date = () => new Date(),
): NotificationRepository {
  let values = initial.map((value) => ({ ...value }));
  let nextId = values.length + 1;
  return {
    async list() {
      const currentTime = now().getTime();
      return values
        .filter(
          (value) =>
            value.expiresAt === undefined ||
            new Date(value.expiresAt).getTime() > currentTime,
        )
        .map((value) => ({ ...value }));
    },
    async create(draft) {
      const validated = validateSubscriptionDraft(draft, now());
      const subscription: NotificationSubscription = {
        ...validated,
        id: `fixture:subscription:${nextId++}`,
        createdAt: now().toISOString(),
      };
      values = [...values, subscription];
      return { ...subscription };
    },
    async remove(id) {
      values = values.filter((value) => value.id !== id);
    },
  };
}

export const notificationRepository = createFixtureNotificationRepository();
