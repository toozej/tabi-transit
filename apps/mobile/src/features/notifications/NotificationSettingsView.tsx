import MaterialCommunityIcons from "@expo/vector-icons/MaterialCommunityIcons";
import { useEffect, useState } from "react";
import {
  Alert,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";

import {
  notificationRepository,
  unavailableNotificationRegistration,
} from "@/data/api/notificationRepository";
import type { NotificationSubscriptionDraft } from "@/domain/notifications";
import { useNotificationStore } from "@/state/notificationStore";
import { tabi, tabiCommonStyles } from "@/ui/tabi";

function futureIso(days: number) {
  const value = new Date();
  value.setDate(value.getDate() + days);
  return value.toISOString();
}

function serviceAlertDraft(): NotificationSubscriptionDraft {
  return {
    type: "service_alert",
    scope: { routeId: "trimet:route:20", source: "fixture-alerts" },
    quietHours: {
      startsAt: "22:00",
      endsAt: "07:00",
      timeZone: "America/Los_Angeles",
    },
    expiresAt: futureIso(30),
  };
}

function departureDraft(): NotificationSubscriptionDraft {
  return {
    type: "departure_reminder",
    scope: { stopId: "trimet:stop:101" },
    leadMinutes: 10,
    expiresAt: futureIso(1),
  };
}

function describeScope(scope: NotificationSubscriptionDraft["scope"]): string {
  return (
    scope.routeId ??
    scope.stopId ??
    scope.mode ??
    scope.source ??
    "Scope unavailable"
  );
}

function formatExpiry(value?: string) {
  if (!value) return "No expiry";
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
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
        "Subscription saved on this device. Push delivery remains disabled in this fixture build.",
      );
    } catch (error) {
      setStatus(
        error instanceof Error
          ? error.message
          : "Notification subscription is unavailable.",
      );
    }
  }

  function confirmRemove(id: string, label: string) {
    Alert.alert(
      "Delete notification?",
      `${label} will be removed from this device.`,
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Delete",
          style: "destructive",
          onPress: () => {
            void notificationRepository.remove(id).then(() => {
              removeSubscription(id);
              setStatus("Subscription deleted from this device.");
            });
          },
        },
      ],
    );
  }

  return (
    <ScrollView
      contentContainerStyle={styles.page}
      style={tabiCommonStyles.screen}
      showsVerticalScrollIndicator={false}
    >
      <View style={styles.headingBlock}>
        <Text style={tabiCommonStyles.eyebrow}>STAY AHEAD</Text>
        <Text accessibilityRole="header" style={tabiCommonStyles.title}>
          Notifications
        </Text>
        <Text style={tabiCommonStyles.subtitle}>
          Choose useful moments first. Tabi only asks for system permission when
          delivery is available and you turn something on.
        </Text>
      </View>

      <View style={styles.statusCard}>
        <View style={styles.statusIcon}>
          <MaterialCommunityIcons
            name="bell-off-outline"
            color={tabi.color.warning}
            size={23}
          />
        </View>
        <View style={styles.statusCopy}>
          <Text style={styles.statusTitle}>Push delivery is off</Text>
          <Text accessibilityLiveRegion="polite" style={styles.statusText}>
            {status}
          </Text>
        </View>
      </View>

      <View style={styles.section}>
        <Text accessibilityRole="header" style={tabiCommonStyles.sectionTitle}>
          Try notification choices
        </Text>
        <Text style={tabiCommonStyles.secondary}>
          These fixture choices exercise the same local controls on iOS and
          Android.
        </Text>

        <View style={styles.optionCard}>
          <View style={styles.optionTopline}>
            <View style={styles.optionIcon}>
              <MaterialCommunityIcons
                name="alert-circle-outline"
                color={tabi.color.accent}
                size={22}
              />
            </View>
            <View style={styles.optionCopy}>
              <Text style={styles.optionTitle}>Route 20 service alerts</Text>
              <Text style={styles.optionDescription}>
                Quiet from 10 PM to 7 AM · Expires in 30 days
              </Text>
            </View>
          </View>
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Add fixture service alert subscription for route 20"
            onPress={() => void add(serviceAlertDraft())}
            style={({ pressed }) => [
              styles.addButton,
              pressed && styles.addButtonPressed,
            ]}
          >
            <MaterialCommunityIcons
              name="plus"
              color={tabi.color.white}
              size={19}
            />
            <Text style={styles.addButtonText}>Add service alert</Text>
          </Pressable>
        </View>

        <View style={styles.optionCard}>
          <View style={styles.optionTopline}>
            <View style={styles.optionIcon}>
              <MaterialCommunityIcons
                name="clock-alert-outline"
                color={tabi.color.accent}
                size={22}
              />
            </View>
            <View style={styles.optionCopy}>
              <Text style={styles.optionTitle}>Stop 101 departure</Text>
              <Text style={styles.optionDescription}>
                10-minute reminder · Expires tomorrow
              </Text>
            </View>
          </View>
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Add fixture departure reminder for stop 101"
            onPress={() => void add(departureDraft())}
            style={({ pressed }) => [
              styles.addButton,
              pressed && styles.addButtonPressed,
            ]}
          >
            <MaterialCommunityIcons
              name="plus"
              color={tabi.color.white}
              size={19}
            />
            <Text style={styles.addButtonText}>Add reminder</Text>
          </Pressable>
        </View>
      </View>

      <View style={styles.section}>
        <View style={styles.sectionTopline}>
          <Text
            accessibilityRole="header"
            style={tabiCommonStyles.sectionTitle}
          >
            Your notifications
          </Text>
          <Text style={styles.count}>{subscriptions.length} ACTIVE</Text>
        </View>

        {subscriptions.length === 0 ? (
          <View style={styles.emptyCard}>
            <MaterialCommunityIcons
              name="bell-plus-outline"
              color={tabi.color.accent}
              size={29}
            />
            <Text style={styles.emptyTitle}>Nothing configured yet</Text>
            <Text style={styles.emptyText}>
              Add one of the choices above to preview notification controls.
            </Text>
          </View>
        ) : (
          subscriptions.map((subscription) => {
            const label =
              subscription.type === "service_alert"
                ? "Service alert"
                : "Departure reminder";
            return (
              <View key={subscription.id} style={styles.subscription}>
                <View style={styles.subscriptionHeading}>
                  <View style={styles.subscriptionIcon}>
                    <MaterialCommunityIcons
                      name={
                        subscription.type === "service_alert"
                          ? "alert-circle-outline"
                          : "clock-outline"
                      }
                      color={tabi.color.accent}
                      size={21}
                    />
                  </View>
                  <View style={styles.subscriptionCopy}>
                    <Text
                      accessibilityRole="header"
                      style={styles.subscriptionTitle}
                    >
                      {label}
                    </Text>
                    <Text style={styles.subscriptionScope}>
                      {describeScope(subscription.scope)}
                    </Text>
                  </View>
                </View>
                <View style={styles.metadata}>
                  {subscription.leadMinutes !== undefined && (
                    <Text
                      style={styles.metaText}
                    >{`${subscription.leadMinutes} minute lead time`}</Text>
                  )}
                  {subscription.quietHours && (
                    <Text
                      style={styles.metaText}
                    >{`Quiet ${subscription.quietHours.startsAt}–${subscription.quietHours.endsAt}`}</Text>
                  )}
                  <Text
                    style={styles.metaText}
                  >{`Expires ${formatExpiry(subscription.expiresAt)}`}</Text>
                </View>
                <Pressable
                  accessibilityRole="button"
                  accessibilityLabel={`Delete ${subscription.type} subscription`}
                  onPress={() => confirmRemove(subscription.id, label)}
                  style={({ pressed }) => [
                    styles.deleteButton,
                    pressed && tabiCommonStyles.pressed,
                  ]}
                >
                  <MaterialCommunityIcons
                    name="trash-can-outline"
                    color={tabi.color.danger}
                    size={18}
                  />
                  <Text style={styles.deleteText}>Delete</Text>
                </Pressable>
              </View>
            );
          })
        )}
      </View>

      <View style={styles.privacyNote}>
        <MaterialCommunityIcons
          name="shield-check-outline"
          color={tabi.color.success}
          size={20}
        />
        <Text style={styles.privacyText}>
          Background work only maintains cached transit data and reconciles your
          subscriptions. Tabi does not continuously monitor location or
          vehicles.
        </Text>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  page: {
    ...tabiCommonStyles.page,
    alignSelf: "center",
    maxWidth: 680,
    width: "100%",
  },
  headingBlock: { gap: 7, paddingTop: 6 },
  statusCard: {
    alignItems: "flex-start",
    backgroundColor: tabi.color.warningSoft,
    borderRadius: tabi.radius.medium,
    flexDirection: "row",
    gap: 12,
    padding: 15,
  },
  statusIcon: {
    alignItems: "center",
    backgroundColor: tabi.color.surface,
    borderRadius: tabi.radius.pill,
    height: 42,
    justifyContent: "center",
    width: 42,
  },
  statusCopy: { flex: 1, gap: 4 },
  statusTitle: { color: tabi.color.ink, fontSize: 15, fontWeight: "800" },
  statusText: { color: tabi.color.mutedInk, fontSize: 13, lineHeight: 18 },
  section: { gap: 11 },
  optionCard: {
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    gap: 14,
    padding: 15,
  },
  optionTopline: { alignItems: "center", flexDirection: "row", gap: 11 },
  optionIcon: {
    alignItems: "center",
    backgroundColor: tabi.color.accentSoft,
    borderRadius: tabi.radius.small,
    height: 42,
    justifyContent: "center",
    width: 42,
  },
  optionCopy: { flex: 1, gap: 3 },
  optionTitle: { color: tabi.color.ink, fontSize: 16, fontWeight: "700" },
  optionDescription: {
    color: tabi.color.mutedInk,
    fontSize: 12,
    lineHeight: 17,
  },
  addButton: {
    alignItems: "center",
    backgroundColor: tabi.color.accent,
    borderRadius: tabi.radius.small,
    flexDirection: "row",
    gap: 7,
    justifyContent: "center",
    minHeight: tabi.touchTarget,
  },
  addButtonPressed: { backgroundColor: tabi.color.accentPressed },
  addButtonText: { color: tabi.color.white, fontSize: 15, fontWeight: "700" },
  sectionTopline: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  count: {
    color: tabi.color.mutedInk,
    fontFamily: tabi.type.utility,
    fontSize: 10,
    fontWeight: "700",
    letterSpacing: 0.8,
  },
  emptyCard: {
    alignItems: "center",
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderStyle: "dashed",
    borderWidth: 1,
    padding: 22,
  },
  emptyTitle: {
    color: tabi.color.ink,
    fontSize: 16,
    fontWeight: "700",
    marginTop: 9,
  },
  emptyText: {
    color: tabi.color.mutedInk,
    fontSize: 13,
    lineHeight: 18,
    marginTop: 4,
    textAlign: "center",
  },
  subscription: {
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    gap: 11,
    padding: 15,
  },
  subscriptionHeading: { alignItems: "center", flexDirection: "row", gap: 11 },
  subscriptionIcon: {
    alignItems: "center",
    backgroundColor: tabi.color.accentSoft,
    borderRadius: tabi.radius.pill,
    height: 40,
    justifyContent: "center",
    width: 40,
  },
  subscriptionCopy: { flex: 1, gap: 2 },
  subscriptionTitle: { color: tabi.color.ink, fontSize: 16, fontWeight: "700" },
  subscriptionScope: { color: tabi.color.mutedInk, fontSize: 12 },
  metadata: {
    backgroundColor: tabi.color.surfaceMuted,
    borderRadius: tabi.radius.small,
    gap: 3,
    padding: 10,
  },
  metaText: { color: tabi.color.mutedInk, fontSize: 12, lineHeight: 17 },
  deleteButton: {
    alignItems: "center",
    alignSelf: "flex-start",
    flexDirection: "row",
    gap: 6,
    minHeight: tabi.touchTarget,
    paddingRight: 14,
  },
  deleteText: { color: tabi.color.danger, fontSize: 14, fontWeight: "700" },
  privacyNote: {
    alignItems: "flex-start",
    borderTopColor: tabi.color.border,
    borderTopWidth: StyleSheet.hairlineWidth,
    flexDirection: "row",
    gap: 10,
    paddingTop: 16,
  },
  privacyText: {
    color: tabi.color.mutedInk,
    flex: 1,
    fontSize: 12,
    lineHeight: 18,
  },
});
