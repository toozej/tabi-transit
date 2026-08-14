import MaterialCommunityIcons from "@expo/vector-icons/MaterialCommunityIcons";
import { useRouter } from "expo-router";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { formatDistance } from "@/domain/riderInfo";
import type { Vehicle } from "@/domain/vehicleModels";
import { tabi } from "@/ui/tabi";

import { useNearbyStops } from "./queries";

const color = {
  paper: tabi.color.canvas,
  card: tabi.color.surface,
  ink: tabi.color.ink,
  mutedInk: tabi.color.mutedInk,
  rule: tabi.color.border,
  line: tabi.color.accent,
  bus: tabi.color.bus,
  rail: tabi.color.rail,
};
const { type } = tabi;

function modeLabel(mode: string) {
  return mode === "light_rail" ? "RAIL" : mode.toUpperCase();
}

function displayFreshness(freshness: Vehicle["freshness"]) {
  const age = Math.round(freshness.ageSeconds);
  return `${freshness.isRealtime ? "Live data" : "Latest available data"} · updated ${age} second${age === 1 ? "" : "s"} ago`;
}

/**
 * The red transit thread is the signature: it only appears beside a place the
 * rider can open, echoing the line a finger follows on a printed map.
 */
export function AppleNearbyView() {
  const router = useRouter();
  const query = useNearbyStops(undefined, 2);

  if (query.isLoading) {
    return (
      <SafeAreaView style={styles.safe}>
        <Text accessibilityLiveRegion="polite" style={styles.state}>
          Finding the stops around you.
        </Text>
      </SafeAreaView>
    );
  }

  if (query.isError) {
    return (
      <SafeAreaView style={styles.safe}>
        <Text accessibilityRole="alert" style={styles.state}>
          Rider information is unavailable. Cached or fixture data is not
          presented as live.
        </Text>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView edges={["top"]} style={styles.safe}>
      <ScrollView
        contentContainerStyle={styles.page}
        showsVerticalScrollIndicator={false}
      >
        <View style={styles.masthead}>
          <Text style={styles.wordmark}>TABI</Text>
          <Text accessibilityRole="header" style={styles.title}>
            Departures,
            {"\n"}within reach.
          </Text>
          <Text style={styles.intro}>
            Choose a stop to see its next arrivals and saved details.
          </Text>
        </View>

        <View style={styles.locationNote}>
          <MaterialCommunityIcons
            name="crosshairs-gps"
            color={color.line}
            size={18}
          />
          <Text style={styles.locationText}>
            Location is off — showing available sample stops.
          </Text>
        </View>

        <View style={styles.sectionHeader}>
          <Text accessibilityRole="header" style={styles.sectionTitle}>
            Your lines of sight
          </Text>
          <Text style={styles.count}>
            {query.data?.groups.length ?? 0} MODES
          </Text>
        </View>

        {query.data?.groups.map((group) => (
          <View key={group.mode} style={styles.modeGroup}>
            <View style={styles.modeHeader}>
              <View
                style={[
                  styles.modeTag,
                  group.mode === "light_rail" ? styles.railTag : styles.busTag,
                ]}
              >
                <Text style={styles.modeTagText}>{modeLabel(group.mode)}</Text>
              </View>
              <Text style={styles.modeName}>
                {group.mode.replace("_", " ")}
              </Text>
            </View>

            {group.stops.map((stop, index) => (
              <Pressable
                key={stop.id}
                accessibilityRole="link"
                accessibilityLabel={`${stop.name}, ${formatDistance(stop.distanceMeters)}`}
                onPress={() =>
                  router.push({
                    pathname: "/stop/[stopId]",
                    params: { stopId: stop.id },
                  })
                }
                style={({ pressed }) => [
                  styles.stopRow,
                  index === group.stops.length - 1 && styles.stopRowLast,
                  pressed && styles.stopRowPressed,
                ]}
              >
                <View style={styles.thread}>
                  <View style={styles.threadDot} />
                  {index !== group.stops.length - 1 && (
                    <View style={styles.threadLine} />
                  )}
                </View>
                <View style={styles.stopCopy}>
                  <Text style={styles.stopName}>{stop.name}</Text>
                  <Text style={styles.stopDistance}>
                    {formatDistance(stop.distanceMeters)}
                  </Text>
                </View>
                <MaterialCommunityIcons
                  name="arrow-top-right"
                  color={color.line}
                  size={20}
                />
              </Pressable>
            ))}
          </View>
        ))}

        {query.data && (
          <Text accessibilityLiveRegion="polite" style={styles.freshness}>
            {displayFreshness(query.data.freshness)}
          </Text>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: color.paper, flex: 1 },
  page: {
    alignSelf: "center",
    maxWidth: 680,
    paddingBottom: 34,
    width: "100%",
  },
  state: {
    color: color.ink,
    fontFamily: type.body,
    fontSize: 16,
    lineHeight: 24,
    padding: 24,
  },
  masthead: { paddingHorizontal: 24, paddingTop: 18 },
  wordmark: {
    color: color.line,
    fontFamily: type.utility,
    fontSize: 11,
    fontWeight: "700",
    letterSpacing: 2.2,
  },
  title: {
    color: color.ink,
    fontFamily: type.display,
    fontSize: 40,
    letterSpacing: -1.3,
    lineHeight: 43,
    marginTop: 10,
  },
  intro: {
    color: color.mutedInk,
    fontFamily: type.body,
    fontSize: 16,
    lineHeight: 23,
    marginTop: 12,
    maxWidth: 310,
  },
  locationNote: {
    alignItems: "center",
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderColor: color.rule,
    borderTopWidth: StyleSheet.hairlineWidth,
    flexDirection: "row",
    gap: 9,
    marginHorizontal: 24,
    marginTop: 25,
    paddingVertical: 13,
  },
  locationText: {
    color: color.mutedInk,
    flex: 1,
    fontFamily: type.body,
    fontSize: 13,
    lineHeight: 18,
  },
  sectionHeader: {
    alignItems: "baseline",
    flexDirection: "row",
    justifyContent: "space-between",
    paddingHorizontal: 24,
    paddingTop: 26,
  },
  sectionTitle: {
    color: color.ink,
    fontFamily: type.display,
    fontSize: 24,
    letterSpacing: -0.3,
  },
  count: {
    color: color.mutedInk,
    fontFamily: type.utility,
    fontSize: 10,
    letterSpacing: 0.8,
  },
  modeGroup: { marginHorizontal: 16, marginTop: 16 },
  modeHeader: {
    alignItems: "center",
    flexDirection: "row",
    gap: 9,
    paddingHorizontal: 8,
    paddingBottom: 8,
  },
  modeTag: { borderRadius: 3, paddingHorizontal: 6, paddingVertical: 3 },
  busTag: { backgroundColor: color.bus },
  railTag: { backgroundColor: color.rail },
  modeTagText: {
    color: "#FFFFFF",
    fontFamily: type.utility,
    fontSize: 10,
    fontWeight: "700",
    letterSpacing: 0.6,
  },
  modeName: {
    color: color.mutedInk,
    fontFamily: type.body,
    fontSize: 14,
    textTransform: "capitalize",
  },
  stopRow: {
    alignItems: "center",
    backgroundColor: color.card,
    borderRadius: tabi.radius.medium,
    borderColor: color.rule,
    borderTopWidth: StyleSheet.hairlineWidth,
    flexDirection: "row",
    minHeight: 82,
    paddingHorizontal: 16,
  },
  stopRowLast: { borderBottomWidth: StyleSheet.hairlineWidth },
  stopRowPressed: { backgroundColor: "#EEEAE2" },
  thread: {
    alignItems: "center",
    alignSelf: "stretch",
    marginRight: 13,
    paddingTop: 27,
    width: 12,
  },
  threadDot: {
    backgroundColor: color.line,
    borderColor: color.card,
    borderRadius: 7,
    borderWidth: 3,
    height: 14,
    width: 14,
    zIndex: 1,
  },
  threadLine: {
    backgroundColor: color.line,
    flex: 1,
    marginTop: -2,
    opacity: 0.62,
    width: 1,
  },
  stopCopy: { flex: 1, gap: 4 },
  stopName: {
    color: color.ink,
    fontFamily: type.body,
    fontSize: 17,
    fontWeight: "600",
    letterSpacing: -0.2,
  },
  stopDistance: {
    color: color.mutedInk,
    fontFamily: type.utility,
    fontSize: 11,
    letterSpacing: 0.1,
  },
  freshness: {
    color: color.mutedInk,
    fontFamily: type.utility,
    fontSize: 10,
    lineHeight: 16,
    marginHorizontal: 24,
    marginTop: 24,
  },
});
