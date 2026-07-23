import { create } from "zustand";

import type { NotificationSubscription } from "@/domain/notifications";

type NotificationState = {
  subscriptions: NotificationSubscription[];
  setSubscriptions: (subscriptions: NotificationSubscription[]) => void;
  addSubscription: (subscription: NotificationSubscription) => void;
  removeSubscription: (id: string) => void;
};

/** UI mirror only; a future authenticated API owns durable subscription truth. */
export const useNotificationStore = create<NotificationState>((set) => ({
  subscriptions: [],
  setSubscriptions: (subscriptions) => set({ subscriptions }),
  addSubscription: (subscription) =>
    set((state) => ({ subscriptions: [...state.subscriptions, subscription] })),
  removeSubscription: (id) =>
    set((state) => ({
      subscriptions: state.subscriptions.filter(
        (subscription) => subscription.id !== id,
      ),
    })),
}));
