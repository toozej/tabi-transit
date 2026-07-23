import { Link } from "expo-router";
import { useEffect } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";

import { formatDistance, serviceDayTime } from "@/domain/riderInfo";
import { formatFreshness } from "@/domain/vehicleModels";
import { useSavedStore } from "@/state/savedStore";
import {
  useAlerts,
  useArrivals,
  useNearbyStops,
  useRoute,
  useRouteShape,
  useSchedule,
  useStaticManifest,
  useStop,
} from "./queries";

function State({
  loading,
  error,
  children,
}: {
  loading: boolean;
  error: boolean;
  children: React.ReactNode;
}) {
  if (loading)
    return (
      <Text accessibilityLiveRegion="polite">Loading rider information.</Text>
    );
  if (error)
    return (
      <Text accessibilityRole="alert">
        Rider information is unavailable. Cached or fixture data is not
        presented as live.
      </Text>
    );
  return <>{children}</>;
}
export function NearbyView() {
  const query = useNearbyStops(2);
  return (
    <State loading={query.isLoading} error={query.isError}>
      <ScrollView contentContainerStyle={styles.page}>
        <Text accessibilityRole="header" style={styles.heading}>
          Nearby stops
        </Text>
        <Text>
          Nearest two stops for each selected mode. Distances are straight line.
        </Text>
        {query.data?.groups.map((group) => (
          <View
            key={group.mode}
            accessibilityLabel={`${group.mode} nearby stops`}
          >
            <Text accessibilityRole="header" style={styles.subheading}>
              {group.mode}
            </Text>
            {group.stops.map((stop) => (
              <Link
                key={stop.id}
                href={{
                  pathname: "/stop/[stopId]",
                  params: { stopId: stop.id },
                }}
                asChild
              >
                <Pressable
                  accessibilityRole="link"
                  accessibilityLabel={`${stop.name}, ${formatDistance(stop.distanceMeters)}`}
                >
                  <Text>{`${stop.name} · ${formatDistance(stop.distanceMeters)}`}</Text>
                </Pressable>
              </Link>
            ))}
          </View>
        ))}
        {query.data && <Text>{formatFreshness(query.data.freshness)}</Text>}
      </ScrollView>
    </State>
  );
}
export function StopView({ id }: { id: string }) {
  const stop = useStop(id);
  const arrivals = useArrivals(id);
  const schedule = useSchedule(id);
  const { saved, toggleSaved, addRecent } = useSavedStore();
  useEffect(() => {
    if (stop.data) addRecent({ id, label: stop.data.name, kind: "stop" });
  }, [addRecent, id, stop.data]);
  const isSaved = saved.some((item) => item.id === id);
  return (
    <State loading={stop.isLoading} error={stop.isError}>
      <ScrollView contentContainerStyle={styles.page}>
        {stop.data && (
          <>
            <Text accessibilityRole="header" style={styles.heading}>
              {stop.data.name}
            </Text>
            <Text>{`Stop ID: ${stop.data.id}`}</Text>
            <Text>{`Modes: ${stop.data.modes.join(", ")}`}</Text>
            <Text>
              {stop.data.wheelchairAccessible
                ? "Wheelchair accessible"
                : "Accessibility status unavailable"}
            </Text>
            <Pressable
              accessibilityRole="button"
              onPress={() =>
                toggleSaved({ id, label: stop.data!.name, kind: "stop" })
              }
            >
              <Text>{isSaved ? "Remove saved stop" : "Save stop"}</Text>
            </Pressable>
            <Text accessibilityRole="header" style={styles.subheading}>
              Arrivals
            </Text>
            {arrivals.data?.map((arrival) => (
              <Text
                key={arrival.id}
              >{`${arrival.routeId} to ${arrival.headsign ?? "destination unavailable"}: ${arrival.status}${arrival.estimatedAt ? " (estimated)" : " (scheduled)"}`}</Text>
            ))}
            {arrivals.isError && (
              <Text accessibilityRole="alert">
                Arrivals unavailable; no realtime claim is shown.
              </Text>
            )}
            <Text accessibilityRole="header" style={styles.subheading}>
              Offline schedule
            </Text>
            {schedule.data?.schedule.map((time) => (
              <Text
                key={time.tripId}
              >{`${serviceDayTime(time.serviceDaySeconds)} to ${time.headsign ?? "destination unavailable"} (service-day time)`}</Text>
            ))}
            {schedule.isError && (
              <Text>
                Schedule is unavailable offline until a static-feed artifact is
                downloaded.
              </Text>
            )}
          </>
        )}
      </ScrollView>
    </State>
  );
}
export function RouteView({ id }: { id: string }) {
  const route = useRoute(id);
  const shape = useRouteShape(id);
  const manifest = useStaticManifest();
  return (
    <State loading={route.isLoading} error={route.isError}>
      <ScrollView contentContainerStyle={styles.page}>
        {route.data && (
          <>
            <Text
              accessibilityRole="header"
              style={styles.heading}
            >{`${route.data.route.shortName} ${route.data.route.longName}`}</Text>
            <Text>{`Mode: ${route.data.route.mode}`}</Text>
            <Text accessibilityRole="header" style={styles.subheading}>
              Directions
            </Text>
            {route.data.directions.map((direction) => (
              <Text key={direction.directionId}>
                {direction.headsign ?? `Direction ${direction.directionId}`}
              </Text>
            ))}
            <Text>{formatFreshness(route.data.freshness)}</Text>
            {shape.data && (
              <Text>{`Route shape ready for map rendering: ${shape.data.features.length} line feature(s).`}</Text>
            )}
            {manifest.data && (
              <Text>{`Offline static feed available for sync: ${manifest.data.staticFeedVersion}; ${manifest.data.artifacts.length} artifacts listed.`}</Text>
            )}
          </>
        )}
      </ScrollView>
    </State>
  );
}
export function AlertsView() {
  const query = useAlerts();
  return (
    <State loading={query.isLoading} error={query.isError}>
      <ScrollView contentContainerStyle={styles.page}>
        <Text accessibilityRole="header" style={styles.heading}>
          Alerts
        </Text>
        {query.data?.map((alert) => (
          <View
            key={alert.id}
            accessibilityLabel={`${alert.severity ?? "unknown"} alert`}
          >
            <Text accessibilityRole="header">{alert.header}</Text>
            <Text>{alert.description ?? "No additional description."}</Text>
            <Text>{`${alert.effect ?? "Effect unavailable"}; ${formatFreshness(alert.freshness)}`}</Text>
          </View>
        ))}
      </ScrollView>
    </State>
  );
}
export function SavedView() {
  const { saved, recents, clearRecents } = useSavedStore();
  return (
    <ScrollView contentContainerStyle={styles.page}>
      <Text accessibilityRole="header" style={styles.heading}>
        Saved
      </Text>
      {saved.length === 0 ? (
        <Text>No saved stops, routes, or vehicles yet.</Text>
      ) : (
        saved.map((item) => <Text key={item.id}>{item.label}</Text>)
      )}
      <Text accessibilityRole="header" style={styles.subheading}>
        Recent
      </Text>
      {recents.length === 0 ? (
        <Text>No recent selections.</Text>
      ) : (
        recents.map((item) => <Text key={item.id}>{item.label}</Text>)
      )}
      <Pressable accessibilityRole="button" onPress={clearRecents}>
        <Text>Clear recent selections</Text>
      </Pressable>
      <Link href="/settings/notifications" asChild>
        <Pressable accessibilityRole="link">
          <Text>Notification settings</Text>
        </Pressable>
      </Link>
    </ScrollView>
  );
}
export function SchedulePreview() {
  return <Text>{serviceDayTime(90_060)}</Text>;
}
const styles = StyleSheet.create({
  page: { gap: 12, padding: 20 },
  heading: { fontSize: 24, fontWeight: "600" },
  subheading: { fontSize: 19, fontWeight: "600", marginTop: 8 },
});
