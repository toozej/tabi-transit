import MaterialCommunityIcons from "@expo/vector-icons/MaterialCommunityIcons";
import { useRouter } from "expo-router";
import { useEffect, useState } from "react";
import {
  Alert,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { orderRouteStops, serviceDayTime } from "@/domain/riderInfo";
import { formatFreshness } from "@/domain/vehicleModels";
import { useSavedStore, type SavedKind } from "@/state/savedStore";
import { tabi, tabiCommonStyles } from "@/ui/tabi";

import { AppleNearbyView } from "./AppleNearbyView";
import {
  useAlerts,
  useArrivals,
  useRoute,
  useRouteShape,
  useRouteStops,
  useSchedule,
  useStaticManifest,
  useStop,
} from "./queries";

const iconForKind: Record<
  SavedKind,
  keyof typeof MaterialCommunityIcons.glyphMap
> = {
  stop: "map-marker-outline",
  route: "transit-connection-variant",
  vehicle: "bus",
  place: "map-marker-star-outline",
  trip: "directions-fork",
};

function LoadingState({ error }: { error: boolean }) {
  return (
    <View style={styles.centerState}>
      <View style={styles.stateIcon}>
        <MaterialCommunityIcons
          name={error ? "alert-circle-outline" : "map-search-outline"}
          color={error ? tabi.color.danger : tabi.color.accent}
          size={28}
        />
      </View>
      <Text
        accessibilityLiveRegion={error ? undefined : "polite"}
        accessibilityRole={error ? "alert" : undefined}
        style={styles.stateTitle}
      >
        {error
          ? "Rider information is unavailable"
          : "Loading rider information"}
      </Text>
      <Text style={styles.stateCopy}>
        {error
          ? "Cached or fixture data is never presented as live. Try again when your connection returns."
          : "Checking the latest available transit data…"}
      </Text>
    </View>
  );
}

function State({
  loading,
  error,
  children,
}: {
  loading: boolean;
  error: boolean;
  children: React.ReactNode;
}) {
  if (loading || error) return <LoadingState error={error} />;
  return <>{children}</>;
}

function ScreenHeading({
  eyebrow,
  title,
  subtitle,
}: {
  eyebrow: string;
  title: string;
  subtitle?: string;
}) {
  return (
    <View style={styles.headingBlock}>
      <Text style={tabiCommonStyles.eyebrow}>{eyebrow}</Text>
      <Text accessibilityRole="header" style={tabiCommonStyles.title}>
        {title}
      </Text>
      {subtitle && <Text style={tabiCommonStyles.subtitle}>{subtitle}</Text>}
    </View>
  );
}

function SectionHeading({
  icon,
  title,
}: {
  icon: keyof typeof MaterialCommunityIcons.glyphMap;
  title: string;
}) {
  return (
    <View style={styles.sectionHeading}>
      <MaterialCommunityIcons name={icon} color={tabi.color.accent} size={20} />
      <Text accessibilityRole="header" style={tabiCommonStyles.sectionTitle}>
        {title}
      </Text>
    </View>
  );
}

function formatClock(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value));
}

export function NearbyView() {
  // One experience and one data contract are shared across both platforms.
  // Native differences live in navigation, typography, elevation, and touch.
  return <AppleNearbyView />;
}

export function StopView({ id }: { id: string }) {
  const router = useRouter();
  const stop = useStop(id);
  const arrivals = useArrivals(id);
  const schedule = useSchedule(id);
  const { saved, toggleSaved, addRecent } = useSavedStore();

  useEffect(() => {
    if (stop.data) void addRecent({ id, label: stop.data.name, kind: "stop" });
  }, [addRecent, id, stop.data]);

  const isSaved = saved.some((item) => item.id === id);

  return (
    <State loading={stop.isLoading} error={stop.isError}>
      <ScrollView
        contentContainerStyle={styles.page}
        style={tabiCommonStyles.screen}
        showsVerticalScrollIndicator={false}
      >
        {stop.data && (
          <>
            <ScreenHeading
              eyebrow="STOP DETAILS"
              title={stop.data.name}
              subtitle={`${stop.data.modes.map((mode) => mode.replace("_", " ")).join(" · ")} · ${stop.data.id}`}
            />

            <View style={styles.actionRow}>
              <Pressable
                accessibilityRole="button"
                accessibilityState={{ selected: isSaved }}
                onPress={() =>
                  void toggleSaved({ id, label: stop.data!.name, kind: "stop" })
                }
                style={({ pressed }) => [
                  styles.primaryAction,
                  isSaved && styles.primaryActionSelected,
                  pressed && tabiCommonStyles.pressed,
                ]}
              >
                <MaterialCommunityIcons
                  name={isSaved ? "bookmark" : "bookmark-outline"}
                  color={isSaved ? tabi.color.accent : tabi.color.white}
                  size={20}
                />
                <Text
                  style={[
                    styles.primaryActionText,
                    isSaved && styles.primaryActionTextSelected,
                  ]}
                >
                  {isSaved ? "Saved" : "Save stop"}
                </Text>
              </Pressable>
              <View style={styles.accessibilityPill}>
                <MaterialCommunityIcons
                  name="wheelchair-accessibility"
                  color={
                    stop.data.wheelchairAccessible
                      ? tabi.color.success
                      : tabi.color.mutedInk
                  }
                  size={19}
                />
                <Text style={styles.accessibilityText}>
                  {stop.data.wheelchairAccessible
                    ? "Accessible"
                    : "Not confirmed"}
                </Text>
              </View>
            </View>

            <View style={styles.section}>
              <SectionHeading icon="clock-fast" title="Next arrivals" />
              {arrivals.isError && (
                <View style={styles.warningCard}>
                  <MaterialCommunityIcons
                    name="information-outline"
                    color={tabi.color.warning}
                    size={20}
                  />
                  <Text accessibilityRole="alert" style={styles.warningText}>
                    Arrivals are unavailable, so no realtime claim is shown.
                  </Text>
                </View>
              )}
              {!arrivals.isError && arrivals.data?.length === 0 && (
                <Text style={tabiCommonStyles.secondary}>
                  No upcoming arrivals are available for this stop.
                </Text>
              )}
              {arrivals.data?.map((arrival) => {
                const time = arrival.estimatedAt ?? arrival.scheduledAt;
                return (
                  <Pressable
                    key={arrival.id}
                    accessibilityRole="link"
                    onPress={() =>
                      router.push({
                        pathname: "/route/[routeId]",
                        params: { routeId: arrival.routeId },
                      })
                    }
                    style={({ pressed }) => [
                      styles.arrivalCard,
                      pressed && tabiCommonStyles.pressed,
                    ]}
                  >
                    <View style={styles.routeBadge}>
                      <Text style={styles.routeBadgeText}>
                        {arrival.routeId.split(":").at(-1)}
                      </Text>
                    </View>
                    <View style={styles.arrivalCopy}>
                      <Text style={styles.arrivalHeadsign}>
                        {arrival.headsign ?? "Destination unavailable"}
                      </Text>
                      <Text style={styles.arrivalStatus}>
                        {arrival.estimatedAt
                          ? "Realtime estimate"
                          : "Scheduled"}
                      </Text>
                    </View>
                    <Text style={styles.arrivalTime}>{formatClock(time)}</Text>
                    <MaterialCommunityIcons
                      name="chevron-right"
                      color={tabi.color.faintInk}
                      size={22}
                    />
                  </Pressable>
                );
              })}
            </View>

            <View style={styles.section}>
              <SectionHeading
                icon="calendar-blank-outline"
                title="Offline schedule"
              />
              <View style={tabiCommonStyles.card}>
                {schedule.data?.schedule.map((time) => (
                  <View key={time.tripId} style={styles.scheduleRow}>
                    <Text style={styles.scheduleTime}>
                      {serviceDayTime(time.serviceDaySeconds)}
                    </Text>
                    <Text style={styles.scheduleHeadsign}>
                      {time.headsign ?? "Destination unavailable"}
                    </Text>
                  </View>
                ))}
                {schedule.isError && (
                  <Text style={tabiCommonStyles.secondary}>
                    Download static transit data to use schedules offline.
                  </Text>
                )}
              </View>
            </View>
          </>
        )}
      </ScrollView>
    </State>
  );
}

export function RouteView({ id }: { id: string }) {
  const router = useRouter();
  const route = useRoute(id);
  const shape = useRouteShape(id);
  const manifest = useStaticManifest();
  const [directionId, setDirectionId] = useState<0 | 1 | undefined>();
  const selectedDirection =
    directionId ?? route.data?.directions[0]?.directionId;
  const routeStops = useRouteStops(id, selectedDirection);
  const displayedStops = routeStops.data
    ? orderRouteStops(routeStops.data.stops)
    : [];

  return (
    <State loading={route.isLoading} error={route.isError}>
      <ScrollView
        contentContainerStyle={styles.page}
        style={tabiCommonStyles.screen}
        showsVerticalScrollIndicator={false}
      >
        {route.data && (
          <>
            <ScreenHeading
              eyebrow={`${route.data.route.mode.replace("_", " ").toUpperCase()} ROUTE`}
              title={`${route.data.route.shortName} ${route.data.route.longName}`}
              subtitle={formatFreshness(route.data.freshness)}
            />

            <View style={styles.section}>
              <SectionHeading icon="swap-horizontal" title="Direction" />
              <View style={styles.segmentedControl}>
                {route.data.directions.map((direction) => {
                  const selected = selectedDirection === direction.directionId;
                  return (
                    <Pressable
                      key={direction.directionId}
                      accessibilityRole="button"
                      accessibilityState={{ selected }}
                      onPress={() => setDirectionId(direction.directionId)}
                      style={({ pressed }) => [
                        styles.segment,
                        selected && styles.segmentSelected,
                        pressed && tabiCommonStyles.pressed,
                      ]}
                    >
                      <Text
                        numberOfLines={1}
                        style={[
                          styles.segmentText,
                          selected && styles.segmentTextSelected,
                        ]}
                      >
                        {direction.headsign ??
                          `Direction ${direction.directionId}`}
                      </Text>
                    </Pressable>
                  );
                })}
              </View>
            </View>

            <View style={styles.section}>
              <View style={styles.sectionHeaderSplit}>
                <SectionHeading icon="map-marker-path" title="Stops" />
                <Text style={styles.stopCount}>
                  {displayedStops.length} STOPS
                </Text>
              </View>
              {routeStops.isLoading && (
                <Text
                  accessibilityLiveRegion="polite"
                  style={tabiCommonStyles.secondary}
                >
                  Loading route stops…
                </Text>
              )}
              {routeStops.isError && (
                <Text accessibilityRole="alert" style={styles.errorText}>
                  Route stops are unavailable. No incomplete sequence is shown.
                </Text>
              )}
              <View style={styles.stopList}>
                {displayedStops.map((stop, index) => (
                  <Pressable
                    key={stop.id}
                    accessibilityRole="link"
                    accessibilityLabel={`Stop ${stop.sequence}: ${stop.name}`}
                    onPress={() =>
                      router.push({
                        pathname: "/stop/[stopId]",
                        params: { stopId: stop.id },
                      })
                    }
                    style={({ pressed }) => [
                      styles.routeStopRow,
                      pressed && tabiCommonStyles.pressed,
                    ]}
                  >
                    <View style={styles.routeThread}>
                      <View style={styles.routeDot} />
                      {index < displayedStops.length - 1 && (
                        <View style={styles.routeLine} />
                      )}
                    </View>
                    <View style={styles.routeStopCopy}>
                      <Text style={styles.routeStopName}>{stop.name}</Text>
                      <Text style={styles.routeStopMeta}>
                        Stop {stop.sequence}
                      </Text>
                    </View>
                    <MaterialCommunityIcons
                      name="chevron-right"
                      color={tabi.color.faintInk}
                      size={22}
                    />
                  </Pressable>
                ))}
              </View>
            </View>

            <View style={styles.dataNote}>
              <MaterialCommunityIcons
                name="database-check-outline"
                color={tabi.color.success}
                size={20}
              />
              <Text style={styles.dataNoteText}>
                {shape.data
                  ? `Route map ready · ${shape.data.features.length} line shape${shape.data.features.length === 1 ? "" : "s"}`
                  : "Route map data is not available."}
                {manifest.data
                  ? `\nOffline data ${manifest.data.staticFeedVersion} is available to sync.`
                  : ""}
              </Text>
            </View>
          </>
        )}
      </ScrollView>
    </State>
  );
}

export function AlertsView() {
  const query = useAlerts();
  return (
    <SafeAreaView edges={["top"]} style={tabiCommonStyles.screen}>
      <State loading={query.isLoading} error={query.isError}>
        <ScrollView
          contentContainerStyle={styles.page}
          showsVerticalScrollIndicator={false}
        >
          <ScreenHeading
            eyebrow="SERVICE STATUS"
            title="Alerts"
            subtitle="What could affect your trip right now."
          />
          {query.data?.length === 0 && (
            <View style={styles.emptyCard}>
              <MaterialCommunityIcons
                name="check-circle-outline"
                color={tabi.color.success}
                size={30}
              />
              <Text style={styles.emptyTitle}>No active alerts</Text>
              <Text style={styles.emptyCopy}>
                Everything looks clear for now.
              </Text>
            </View>
          )}
          {query.data?.map((alert) => {
            const severe = alert.severity === "severe";
            return (
              <View
                key={alert.id}
                accessibilityLabel={`${alert.severity ?? "unknown"} alert`}
                style={[styles.alertCard, severe && styles.alertCardSevere]}
              >
                <View style={styles.alertTopline}>
                  <View
                    style={[styles.alertIcon, severe && styles.alertIconSevere]}
                  >
                    <MaterialCommunityIcons
                      name={severe ? "alert-octagon-outline" : "alert-outline"}
                      color={severe ? tabi.color.danger : tabi.color.warning}
                      size={22}
                    />
                  </View>
                  <Text style={styles.alertSeverity}>
                    {(alert.severity ?? "service").toUpperCase()}
                  </Text>
                </View>
                <Text accessibilityRole="header" style={styles.alertTitle}>
                  {alert.header}
                </Text>
                <Text style={styles.alertDescription}>
                  {alert.description ?? "No additional description."}
                </Text>
                <View style={styles.alertFooter}>
                  <Text style={styles.alertEffect}>
                    {(alert.effect ?? "Effect unavailable").replaceAll(
                      "_",
                      " ",
                    )}
                  </Text>
                  <Text style={styles.alertFreshness}>
                    {formatFreshness(alert.freshness)}
                  </Text>
                </View>
              </View>
            );
          })}
        </ScrollView>
      </State>
    </SafeAreaView>
  );
}

export function SavedView() {
  const router = useRouter();
  const { saved, recents, clearRecents, clearAllLocalData, persistence } =
    useSavedStore();

  function confirmClearAll() {
    Alert.alert(
      "Clear saved data?",
      "All saved items and recent selections will be removed from this device.",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Clear all",
          style: "destructive",
          onPress: () => void clearAllLocalData(),
        },
      ],
    );
  }

  return (
    <SafeAreaView edges={["top"]} style={tabiCommonStyles.screen}>
      <ScrollView
        contentContainerStyle={styles.page}
        showsVerticalScrollIndicator={false}
      >
        <ScreenHeading
          eyebrow="YOUR TABI"
          title="Saved"
          subtitle="Stops, routes, and recent places kept close at hand."
        />

        {persistence === "loading" && (
          <Text
            accessibilityLiveRegion="polite"
            style={tabiCommonStyles.secondary}
          >
            Loading local saved data…
          </Text>
        )}
        {persistence === "unavailable" && (
          <View style={styles.warningCard}>
            <MaterialCommunityIcons
              name="database-alert-outline"
              color={tabi.color.warning}
              size={20}
            />
            <Text accessibilityRole="alert" style={styles.warningText}>
              Device storage is unavailable. Changes last for this session only.
            </Text>
          </View>
        )}

        <View style={styles.section}>
          <View style={styles.sectionHeaderSplit}>
            <SectionHeading icon="bookmark-outline" title="Favorites" />
            <Text style={styles.stopCount}>{saved.length} SAVED</Text>
          </View>
          {saved.length === 0 ? (
            <View style={styles.emptyCard}>
              <MaterialCommunityIcons
                name="bookmark-plus-outline"
                color={tabi.color.accent}
                size={30}
              />
              <Text style={styles.emptyTitle}>Build your shortcut list</Text>
              <Text style={styles.emptyCopy}>
                Save a stop from Nearby to find it here.
              </Text>
            </View>
          ) : (
            <View style={styles.savedList}>
              {saved.map((item) => (
                <View key={item.id} style={styles.savedRow}>
                  <View style={styles.savedIcon}>
                    <MaterialCommunityIcons
                      name={iconForKind[item.kind]}
                      color={tabi.color.accent}
                      size={21}
                    />
                  </View>
                  <View style={styles.savedCopy}>
                    <Text style={styles.savedLabel}>{item.label}</Text>
                    <Text style={styles.savedKind}>{item.kind}</Text>
                  </View>
                </View>
              ))}
            </View>
          )}
        </View>

        <View style={styles.section}>
          <View style={styles.sectionHeaderSplit}>
            <SectionHeading icon="history" title="Recent" />
            {recents.length > 0 && (
              <Pressable
                accessibilityRole="button"
                onPress={() => void clearRecents()}
                hitSlop={8}
                style={({ pressed }) => pressed && tabiCommonStyles.pressed}
              >
                <Text style={styles.inlineAction}>Clear</Text>
              </Pressable>
            )}
          </View>
          {recents.length === 0 ? (
            <Text style={tabiCommonStyles.secondary}>
              No recent selections yet.
            </Text>
          ) : (
            <View style={styles.savedList}>
              {recents.map((item) => (
                <View key={item.id} style={styles.savedRow}>
                  <View style={styles.savedIconMuted}>
                    <MaterialCommunityIcons
                      name={iconForKind[item.kind]}
                      color={tabi.color.mutedInk}
                      size={20}
                    />
                  </View>
                  <Text style={styles.savedLabel}>{item.label}</Text>
                </View>
              ))}
            </View>
          )}
        </View>

        <View style={styles.settingsCard}>
          <Pressable
            accessibilityRole="link"
            onPress={() => router.push("/settings/notifications")}
            style={({ pressed }) => [
              styles.settingsRow,
              pressed && tabiCommonStyles.pressed,
            ]}
          >
            <View style={styles.settingsIcon}>
              <MaterialCommunityIcons
                name="bell-outline"
                color={tabi.color.accent}
                size={21}
              />
            </View>
            <View style={styles.settingsCopy}>
              <Text style={styles.settingsTitle}>Notifications</Text>
              <Text style={styles.settingsSubtitle}>
                Service alerts and departure reminders
              </Text>
            </View>
            <MaterialCommunityIcons
              name="chevron-right"
              color={tabi.color.faintInk}
              size={23}
            />
          </Pressable>
        </View>

        {(saved.length > 0 || recents.length > 0) && (
          <Pressable
            accessibilityRole="button"
            onPress={confirmClearAll}
            style={({ pressed }) => [
              styles.destructiveButton,
              pressed && tabiCommonStyles.pressed,
            ]}
          >
            <Text style={styles.destructiveText}>Clear all local data</Text>
          </Pressable>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

export function SchedulePreview() {
  return <Text>{serviceDayTime(90_060)}</Text>;
}

const styles = StyleSheet.create({
  page: {
    ...tabiCommonStyles.page,
    alignSelf: "center",
    maxWidth: 680,
    width: "100%",
  },
  headingBlock: { gap: 7, paddingTop: 6 },
  centerState: {
    alignItems: "center",
    backgroundColor: tabi.color.canvas,
    flex: 1,
    justifyContent: "center",
    padding: 32,
  },
  stateIcon: {
    alignItems: "center",
    backgroundColor: tabi.color.surface,
    borderRadius: tabi.radius.pill,
    height: 58,
    justifyContent: "center",
    marginBottom: 16,
    width: 58,
  },
  stateTitle: {
    color: tabi.color.ink,
    fontSize: 18,
    fontWeight: "700",
    textAlign: "center",
  },
  stateCopy: {
    color: tabi.color.mutedInk,
    fontSize: 14,
    lineHeight: 20,
    marginTop: 7,
    maxWidth: 320,
    textAlign: "center",
  },
  actionRow: { flexDirection: "row", flexWrap: "wrap", gap: 10 },
  primaryAction: {
    alignItems: "center",
    backgroundColor: tabi.color.accent,
    borderColor: tabi.color.accent,
    borderRadius: tabi.radius.pill,
    borderWidth: 1,
    flexDirection: "row",
    gap: 7,
    minHeight: tabi.touchTarget,
    paddingHorizontal: 17,
  },
  primaryActionSelected: { backgroundColor: tabi.color.accentSoft },
  primaryActionText: {
    color: tabi.color.white,
    fontSize: 15,
    fontWeight: "700",
  },
  primaryActionTextSelected: { color: tabi.color.accent },
  accessibilityPill: {
    alignItems: "center",
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.pill,
    borderWidth: StyleSheet.hairlineWidth,
    flexDirection: "row",
    gap: 7,
    minHeight: tabi.touchTarget,
    paddingHorizontal: 14,
  },
  accessibilityText: { color: tabi.color.ink, fontSize: 14, fontWeight: "600" },
  section: { gap: 12 },
  sectionHeading: { alignItems: "center", flexDirection: "row", gap: 8 },
  sectionHeaderSplit: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  arrivalCard: {
    alignItems: "center",
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    flexDirection: "row",
    gap: 12,
    minHeight: 74,
    paddingHorizontal: 14,
    ...tabi.shadow,
  },
  routeBadge: {
    alignItems: "center",
    backgroundColor: tabi.color.bus,
    borderRadius: tabi.radius.small,
    justifyContent: "center",
    minHeight: 34,
    minWidth: 38,
    paddingHorizontal: 7,
  },
  routeBadgeText: { color: tabi.color.white, fontSize: 15, fontWeight: "800" },
  arrivalCopy: { flex: 1, gap: 3 },
  arrivalHeadsign: { color: tabi.color.ink, fontSize: 16, fontWeight: "700" },
  arrivalStatus: { color: tabi.color.success, fontSize: 12, fontWeight: "600" },
  arrivalTime: { color: tabi.color.ink, fontSize: 18, fontWeight: "800" },
  warningCard: {
    alignItems: "flex-start",
    backgroundColor: tabi.color.warningSoft,
    borderRadius: tabi.radius.medium,
    flexDirection: "row",
    gap: 10,
    padding: 14,
  },
  warningText: { color: tabi.color.ink, flex: 1, fontSize: 14, lineHeight: 20 },
  scheduleRow: {
    alignItems: "center",
    flexDirection: "row",
    gap: 16,
    minHeight: 44,
  },
  scheduleTime: { color: tabi.color.ink, fontSize: 16, fontWeight: "800" },
  scheduleHeadsign: { color: tabi.color.mutedInk, flex: 1, fontSize: 15 },
  segmentedControl: {
    backgroundColor: tabi.color.surfaceMuted,
    borderRadius: tabi.radius.small,
    flexDirection: "row",
    padding: 3,
  },
  segment: {
    alignItems: "center",
    borderRadius: tabi.radius.small,
    flex: 1,
    justifyContent: "center",
    minHeight: tabi.touchTarget,
    paddingHorizontal: 8,
  },
  segmentSelected: { backgroundColor: tabi.color.surface, ...tabi.shadow },
  segmentText: { color: tabi.color.mutedInk, fontSize: 14, fontWeight: "600" },
  segmentTextSelected: { color: tabi.color.ink, fontWeight: "800" },
  stopCount: {
    color: tabi.color.mutedInk,
    fontFamily: tabi.type.utility,
    fontSize: 10,
    fontWeight: "700",
    letterSpacing: 0.8,
  },
  errorText: { color: tabi.color.danger, fontSize: 14, lineHeight: 20 },
  stopList: {
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    overflow: "hidden",
  },
  routeStopRow: {
    alignItems: "center",
    flexDirection: "row",
    minHeight: 72,
    paddingRight: 12,
  },
  routeThread: {
    alignItems: "center",
    alignSelf: "stretch",
    marginLeft: 17,
    width: 24,
  },
  routeDot: {
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.accent,
    borderRadius: 7,
    borderWidth: 3,
    height: 14,
    marginTop: 28,
    width: 14,
    zIndex: 1,
  },
  routeLine: {
    backgroundColor: tabi.color.accent,
    flex: 1,
    marginTop: -1,
    width: 2,
  },
  routeStopCopy: { flex: 1, gap: 4, paddingHorizontal: 8 },
  routeStopName: { color: tabi.color.ink, fontSize: 16, fontWeight: "700" },
  routeStopMeta: { color: tabi.color.mutedInk, fontSize: 12 },
  dataNote: {
    alignItems: "flex-start",
    backgroundColor: tabi.color.surface,
    borderRadius: tabi.radius.medium,
    flexDirection: "row",
    gap: 10,
    padding: 14,
  },
  dataNoteText: {
    color: tabi.color.mutedInk,
    flex: 1,
    fontSize: 13,
    lineHeight: 19,
  },
  emptyCard: {
    alignItems: "center",
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderStyle: "dashed",
    borderWidth: 1,
    padding: 24,
  },
  emptyTitle: {
    color: tabi.color.ink,
    fontSize: 17,
    fontWeight: "700",
    marginTop: 10,
  },
  emptyCopy: {
    color: tabi.color.mutedInk,
    fontSize: 14,
    marginTop: 5,
    textAlign: "center",
  },
  alertCard: {
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderLeftColor: tabi.color.warning,
    borderLeftWidth: 4,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    gap: 9,
    padding: 16,
    ...tabi.shadow,
  },
  alertCardSevere: { borderLeftColor: tabi.color.danger },
  alertTopline: { alignItems: "center", flexDirection: "row", gap: 8 },
  alertIcon: {
    alignItems: "center",
    backgroundColor: tabi.color.warningSoft,
    borderRadius: tabi.radius.pill,
    height: 34,
    justifyContent: "center",
    width: 34,
  },
  alertIconSevere: { backgroundColor: tabi.color.dangerSoft },
  alertSeverity: {
    color: tabi.color.mutedInk,
    fontFamily: tabi.type.utility,
    fontSize: 10,
    fontWeight: "700",
    letterSpacing: 1,
  },
  alertTitle: { color: tabi.color.ink, fontSize: 19, fontWeight: "800" },
  alertDescription: { color: tabi.color.ink, fontSize: 15, lineHeight: 21 },
  alertFooter: {
    borderTopColor: tabi.color.border,
    borderTopWidth: StyleSheet.hairlineWidth,
    gap: 3,
    paddingTop: 9,
  },
  alertEffect: {
    color: tabi.color.mutedInk,
    fontSize: 13,
    textTransform: "capitalize",
  },
  alertFreshness: { color: tabi.color.faintInk, fontSize: 11 },
  savedList: {
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    overflow: "hidden",
  },
  savedRow: {
    alignItems: "center",
    flexDirection: "row",
    gap: 12,
    minHeight: 62,
    paddingHorizontal: 14,
  },
  savedIcon: {
    alignItems: "center",
    backgroundColor: tabi.color.accentSoft,
    borderRadius: tabi.radius.pill,
    height: 38,
    justifyContent: "center",
    width: 38,
  },
  savedIconMuted: {
    alignItems: "center",
    backgroundColor: tabi.color.surfaceMuted,
    borderRadius: tabi.radius.pill,
    height: 38,
    justifyContent: "center",
    width: 38,
  },
  savedCopy: { flex: 1, gap: 2 },
  savedLabel: {
    color: tabi.color.ink,
    flex: 1,
    fontSize: 15,
    fontWeight: "600",
  },
  savedKind: {
    color: tabi.color.mutedInk,
    fontSize: 12,
    textTransform: "capitalize",
  },
  inlineAction: { color: tabi.color.accent, fontSize: 14, fontWeight: "700" },
  settingsCard: {
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    overflow: "hidden",
  },
  settingsRow: {
    alignItems: "center",
    flexDirection: "row",
    gap: 12,
    minHeight: 72,
    paddingHorizontal: 14,
  },
  settingsIcon: {
    alignItems: "center",
    backgroundColor: tabi.color.accentSoft,
    borderRadius: tabi.radius.small,
    height: 40,
    justifyContent: "center",
    width: 40,
  },
  settingsCopy: { flex: 1, gap: 3 },
  settingsTitle: { color: tabi.color.ink, fontSize: 16, fontWeight: "700" },
  settingsSubtitle: { color: tabi.color.mutedInk, fontSize: 12 },
  destructiveButton: {
    alignItems: "center",
    minHeight: tabi.touchTarget,
    justifyContent: "center",
  },
  destructiveText: {
    color: tabi.color.danger,
    fontSize: 14,
    fontWeight: "700",
  },
});
