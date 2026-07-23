import { useEffect, useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";

import {
  notificationRepository,
  unavailableNotificationRegistration,
} from "@/data/api/notificationRepository";
import type { NotificationSubscriptionDraft } from "@/domain/notifications";
import { useNotificationStore } from "@/state/notificationStore";

const defaultAlertDraft: NotificationSubscriptionDraft = {
  type: "service_alert",
  scope: { routeId: "trimet:route:20", source: "fixture-alerts" },
  quietHours: {
    startsAt: "22:00",
    endsAt: "07:00",
    timeZone: "America/Los_Angeles",
  },
  expiresAt: "2026-08-01T00:00:00Z",
};

const defaultDepartureDraft: NotificationSubscriptionDraft = {
  type: "departure_reminder",
  scope: { stopId: "trimet:stop:101" },
  leadMinutes: 10,
  expiresAt: "2026-07-24T00:00:00Z",
};

function describeScope(scope: NotificationSubscriptionDraft["scope"]): string {
  return (
    scope.routeId ??
    scope.stopId ??
    scope.mode ??
    scope.source ??
    "scope unavailable"
  );
}

export function NotificationSettingsView() {
  const {
    subscriptions,
    setSubscriptions,
    addSubscription,
    removeSubscription,
  } = useNotificationStore();
  const [status, setStatus] = useState(
    "Notification permission has not been requested.",
  );

  useEffect(() => {
    void notificationRepository.list().then(setSubscriptions);
    void unavailableNotificationRegistration
      .status()
      .then((registration) => setStatus(registration.message));
  }, [setSubscriptions]);

  async function add(draft: NotificationSubscriptionDraft) {
    try {
      const subscription = await notificationRepository.create(draft);
      addSubscription(subscription);
      setStatus(
        "Fixture subscription saved locally. Push delivery remains disabled.",
      );
    } catch (error) {
      setStatus(
        error instanceof Error
          ? error.message
          : "Notification subscription is unavailable.",
      );
    }
  }

  async function remove(id: string) {
    await notificationRepository.remove(id);
    removeSubscription(id);
    setStatus("Fixture subscription deleted locally.");
  }

  return (
    <ScrollView contentContainerStyle={styles.page}>
      <Text accessibilityRole="header" style={styles.heading}>
        Notifications
      </Text>
      <Text>
        Tabi asks for notification permission only after you choose a delivery
        option. This fixture build does not request permission, register a push
        token, or deliver notifications.
      </Text>
      <Text accessibilityLiveRegion="polite">{status}</Text>

      <View style={styles.group}>
        <Text accessibilityRole="header" style={styles.subheading}>
          Add a fixture subscription
        </Text>
        <Text>
          Quiet hours use America/Los_Angeles from 22:00 to 07:00. The alert
          watch expires on August 1, 2026.
        </Text>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Add fixture service alert subscription for route 20"
          onPress={() => void add(defaultAlertDraft)}
          style={styles.button}
        >
          <Text>Add service alert for Route 20</Text>
        </Pressable>
        <Text>
          The departure reminder expires on July 24, 2026 and uses a 10 minute
          lead time.
        </Text>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Add fixture departure reminder for stop 101"
          onPress={() => void add(defaultDepartureDraft)}
          style={styles.button}
        >
          <Text>Add departure reminder for Stop 101</Text>
        </Pressable>
      </View>

      <View style={styles.group}>
        <Text accessibilityRole="header" style={styles.subheading}>
          Your subscriptions
        </Text>
        {subscriptions.length === 0 ? (
          <Text>
            No notification subscriptions are saved in this fixture session.
          </Text>
        ) : (
          subscriptions.map((subscription) => (
            <View key={subscription.id} style={styles.subscription}>
              <Text accessibilityRole="header">
                {subscription.type === "service_alert"
                  ? "Service alert"
                  : "Departure reminder"}
              </Text>
              <Text>{`Scope: ${describeScope(subscription.scope)}`}</Text>
              {subscription.leadMinutes !== undefined && (
                <Text>{`Lead time: ${subscription.leadMinutes} minutes`}</Text>
              )}
              {subscription.quietHours && (
                <Text>{`Quiet hours: ${subscription.quietHours.startsAt}–${subscription.quietHours.endsAt} (${subscription.quietHours.timeZone})`}</Text>
              )}
              {subscription.expiresAt && (
                <Text>{`Expires: ${subscription.expiresAt}`}</Text>
              )}
              <Pressable
                accessibilityRole="button"
                accessibilityLabel={`Delete ${subscription.type} subscription`}
                onPress={() => void remove(subscription.id)}
                style={styles.deleteButton}
              >
                <Text>Delete subscription</Text>
              </Pressable>
            </View>
          ))
        )}
      </View>
      <Text>
        Background work is limited to cache and static-data maintenance plus
        subscription reconciliation. It does not monitor location or poll
        vehicles continuously.
      </Text>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  page: { gap: 12, padding: 20 },
  heading: { fontSize: 24, fontWeight: "600" },
  subheading: { fontSize: 19, fontWeight: "600" },
  group: { gap: 8 },
  button: { borderWidth: 1, borderColor: "#444", padding: 10 },
  deleteButton: { borderWidth: 1, borderColor: "#991b1b", padding: 10 },
  subscription: { borderWidth: 1, borderColor: "#777", gap: 5, padding: 12 },
});
