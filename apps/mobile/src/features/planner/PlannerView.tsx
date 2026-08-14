import { useState } from "react";
import MaterialCommunityIcons from "@expo/vector-icons/MaterialCommunityIcons";
import {
  Linking,
  Pressable,
  ScrollView,
  Share,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import { fixturePlanEndpoints } from "@/data/api/plannerFixtures";
import {
  plannerRepository,
  PlannerFeatureDisabledError,
} from "@/data/api/plannerRepository";
import {
  createPlanDeepLink,
  rankItineraries,
  type Itinerary,
  type PlanEndpoint,
} from "@/domain/tripPlanning";
import { createExternalWalkingDirectionsLink } from "@/domain/externalDirections";
import { usePlannerStore } from "@/state/plannerStore";
import { tabi } from "@/ui/tabi";

const colors = {
  midnight: tabi.color.ink,
  platform: tabi.color.canvas,
  signal: tabi.color.accent,
  route: "#F1C75B",
  track: tabi.color.border,
  ink: tabi.color.ink,
  quiet: tabi.color.mutedInk,
  white: tabi.color.white,
} as const;

function EndpointPicker({
  label,
  value,
  onSelect,
}: {
  label: string;
  value?: PlanEndpoint;
  onSelect: (endpoint: PlanEndpoint) => void;
}) {
  const isOrigin = label === "Origin";
  return (
    <View accessibilityLabel={`${label} picker`} style={styles.endpointGroup}>
      <View style={styles.endpointMarker}>
        <View style={isOrigin ? styles.originDot : styles.destinationDot} />
      </View>
      <View style={styles.endpointBody}>
        <Text style={styles.endpointLabel}>{label}</Text>
        <Text accessibilityRole="header" style={styles.endpointValue}>
          {value?.label ??
            `Choose ${isOrigin ? "where you are" : "where you are going"}`}
        </Text>
        <View style={styles.endpointOptions}>
          {fixturePlanEndpoints.map((endpoint) => {
            const selected = value?.id === endpoint.id;
            return (
              <Pressable
                key={`${label}:${endpoint.id}`}
                accessibilityRole="button"
                accessibilityState={{ selected }}
                accessibilityLabel={`Set ${label.toLowerCase()} to ${endpoint.label}`}
                onPress={() => onSelect(endpoint)}
                style={({ pressed }) => [
                  styles.endpointOption,
                  selected && styles.endpointOptionSelected,
                  pressed && styles.pressed,
                ]}
              >
                <Text
                  style={[
                    styles.endpointOptionText,
                    selected && styles.endpointOptionTextSelected,
                  ]}
                >
                  {endpoint.label}
                </Text>
              </Pressable>
            );
          })}
        </View>
      </View>
    </View>
  );
}

function formatTime(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value));
}

function ItineraryTimeline({ itinerary }: { itinerary: Itinerary }) {
  const duration = Math.round(itinerary.durationSeconds / 60);
  return (
    <View
      accessibilityLabel={`Itinerary taking ${duration} minutes`}
      style={styles.itinerary}
    >
      <View style={styles.itineraryTopline}>
        <View>
          <Text style={styles.boardLabel}>BEST AVAILABLE</Text>
          <Text accessibilityRole="header" style={styles.duration}>
            {duration} min
          </Text>
        </View>
        <View style={styles.arrivalBlock}>
          <Text style={styles.boardLabel}>ARRIVE</Text>
          <Text style={styles.arrivalTime}>
            {formatTime(itinerary.arrivalAt)}
          </Text>
        </View>
      </View>
      <Text style={styles.tripFacts}>
        {`${itinerary.transfers === 0 ? "Direct" : `${itinerary.transfers} transfer${itinerary.transfers === 1 ? "" : "s"}`} · ${itinerary.walkingMeters} m walk`}
      </Text>
      <View style={styles.legs}>
        {itinerary.legs.map((leg, index) => {
          const walkingDirections = createExternalWalkingDirectionsLink(leg);
          const isTransit = leg.mode !== "walk";
          return (
            <View key={leg.id} style={styles.legRow}>
              <View style={styles.legRail}>
                <View
                  style={isTransit ? styles.transitStop : styles.walkStop}
                />
                {index < itinerary.legs.length - 1 && (
                  <View style={styles.railLine} />
                )}
              </View>
              <View style={styles.legContent}>
                <View style={styles.legHeading}>
                  <Text
                    style={isTransit ? styles.routeBadge : styles.walkBadge}
                  >
                    {leg.routeLabel ?? "WALK"}
                  </Text>
                  <Text style={styles.legTime}>{formatTime(leg.startAt)}</Text>
                </View>
                <Text style={styles.legPlace}>{leg.startLabel}</Text>
                <Text style={styles.legDirection}>
                  {leg.headsign
                    ? `Toward ${leg.headsign}`
                    : `Walk to ${leg.endLabel}`}
                </Text>
                {walkingDirections && (
                  <Pressable
                    accessibilityRole="link"
                    accessibilityLabel={`Open walking directions from ${leg.startLabel} to ${leg.endLabel} in your maps app`}
                    onPress={() => void Linking.openURL(walkingDirections)}
                    style={({ pressed }) => [
                      styles.directionsButton,
                      pressed && styles.pressed,
                    ]}
                  >
                    <Text style={styles.directionsButtonText}>
                      OPEN WALKING MAP
                    </Text>
                  </Pressable>
                )}
              </View>
            </View>
          );
        })}
      </View>
      <Text style={styles.freshness}>{itinerary.freshness.message}</Text>
    </View>
  );
}

export function PlannerView() {
  const {
    draft,
    locationPermission,
    setOrigin,
    setDestination,
    setConstraints,
    setLocationPermission,
    swap,
  } = usePlannerStore();
  const [itineraries, setItineraries] = useState<Itinerary[]>([]);
  const [message, setMessage] = useState(
    "Choose an origin and destination to view a synthetic planning result.",
  );

  async function plan() {
    try {
      const ranked = rankItineraries([], draft.constraints);
      const result = await plannerRepository.plan(draft);
      setItineraries(result);
      setMessage(
        result.length > 0
          ? result[0]?.source === "fixture-planner"
            ? "Fixture result. Provider-backed trip planning is currently disabled."
            : "Trip plan from the configured Tabi service."
          : (ranked.disclosure ??
              "No fixture itinerary meets these constraints. Relax a filter."),
      );
    } catch (error) {
      setItineraries([]);
      setMessage(
        error instanceof PlannerFeatureDisabledError
          ? "Choose both endpoints. No current location permission is required."
          : "Trip planning is unavailable.",
      );
    }
  }

  async function sharePlan() {
    if (!draft.origin || !draft.destination) return;
    await Share.share({
      message: `Plan a trip from ${draft.origin.label} to ${draft.destination.label} with Tabi: ${createPlanDeepLink(draft)}`,
      title: "Share Tabi trip",
    });
  }

  return (
    <SafeAreaView edges={["top"]} style={styles.safe}>
      <ScrollView
        contentContainerStyle={styles.page}
        showsVerticalScrollIndicator={false}
      >
        <View style={styles.masthead}>
          <Text style={styles.mastheadKicker}>TABi / TRIP DESK</Text>
          <Text accessibilityRole="header" style={styles.heading}>
            Plan your next move
          </Text>
          <Text style={styles.mastheadCopy}>
            A clear way across the city, with the details that matter at the
            stop.
          </Text>
        </View>

        <View style={styles.journeyCard}>
          <View style={styles.journeyHeading}>
            <Text style={styles.boardLabel}>JOURNEY</Text>
            <Text style={styles.boardTime}>NOW</Text>
          </View>
          <EndpointPicker
            label="Origin"
            value={draft.origin}
            onSelect={setOrigin}
          />
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="Swap origin and destination"
            onPress={swap}
            style={({ pressed }) => [
              styles.swapButton,
              pressed && styles.pressed,
            ]}
          >
            <Text style={styles.swapButtonText}>⇅</Text>
            <Text style={styles.swapButtonLabel}>SWAP STOPS</Text>
          </Pressable>
          <EndpointPicker
            label="Destination"
            value={draft.destination}
            onSelect={setDestination}
          />
        </View>

        <View style={styles.controls}>
          <Text style={styles.boardLabel}>TRIP SETTINGS</Text>
          <View style={styles.controlRow}>
            <Pressable
              accessibilityRole="checkbox"
              accessibilityState={{
                checked: draft.constraints.wheelchairAccessible,
              }}
              onPress={() =>
                setConstraints({
                  wheelchairAccessible: !draft.constraints.wheelchairAccessible,
                })
              }
              style={({ pressed }) => [
                styles.choice,
                draft.constraints.wheelchairAccessible && styles.choiceActive,
                pressed && styles.pressed,
              ]}
            >
              <Text style={styles.choiceMark}>
                {draft.constraints.wheelchairAccessible ? "✓" : "○"}
              </Text>
              <Text style={styles.choiceText}>Step-free</Text>
            </Pressable>
            <Pressable
              accessibilityRole="button"
              onPress={() =>
                setConstraints({
                  optimization:
                    draft.constraints.optimization === "fastest"
                      ? "fewer_transfers"
                      : "fastest",
                })
              }
              style={({ pressed }) => [
                styles.choice,
                pressed && styles.pressed,
              ]}
            >
              <Text style={styles.choiceMark}>↯</Text>
              <Text style={styles.choiceText}>
                {draft.constraints.optimization === "fastest"
                  ? "Fastest"
                  : "Fewer changes"}
              </Text>
            </Pressable>
          </View>
        </View>

        {locationPermission === "denied" ? (
          <View accessibilityRole="alert" style={styles.permissionNote}>
            <Text style={styles.permissionTitle}>Location stays off</Text>
            <Text style={styles.permissionCopy}>
              Choose stops, saved places, or a map pin instead.
            </Text>
          </View>
        ) : (
          <Pressable
            accessibilityRole="button"
            onPress={() => setLocationPermission("denied")}
            style={({ pressed }) => [
              styles.locationButton,
              pressed && styles.pressed,
            ]}
          >
            <Text style={styles.locationButtonText}>
              CONTINUE WITHOUT LOCATION
            </Text>
          </Pressable>
        )}

        <Pressable
          accessibilityRole="button"
          onPress={plan}
          style={({ pressed }) => [
            styles.primaryButton,
            pressed && styles.primaryPressed,
          ]}
        >
          <Text style={styles.primaryButtonText}>FIND TRIPS</Text>
          <Text style={styles.primaryButtonArrow}>→</Text>
        </Pressable>
        <Text accessibilityLiveRegion="polite" style={styles.statusMessage}>
          {message}
        </Text>
        {draft.origin && draft.destination && (
          <Pressable
            accessibilityRole="button"
            onPress={() => void sharePlan()}
            style={({ pressed }) => [
              styles.shareButton,
              pressed && styles.pressed,
            ]}
          >
            <MaterialCommunityIcons
              name="share-variant-outline"
              color={colors.signal}
              size={18}
            />
            <Text style={styles.shareButtonText}>Share this trip</Text>
          </Pressable>
        )}
        {itineraries.map((itinerary) => (
          <ItineraryTimeline key={itinerary.id} itinerary={itinerary} />
        ))}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { backgroundColor: colors.platform, flex: 1 },
  page: {
    alignSelf: "center",
    backgroundColor: colors.platform,
    gap: 18,
    maxWidth: 680,
    padding: 20,
    paddingBottom: 40,
    width: "100%",
  },
  masthead: { gap: 6, paddingTop: 8 },
  mastheadKicker: {
    color: colors.signal,
    fontFamily: "monospace",
    fontSize: 11,
    fontWeight: "700",
    letterSpacing: 1.4,
  },
  heading: {
    color: colors.midnight,
    fontFamily: tabi.type.display,
    fontSize: 32,
    fontWeight: "800",
    letterSpacing: -0.8,
  },
  mastheadCopy: {
    color: colors.quiet,
    fontSize: 15,
    lineHeight: 21,
    maxWidth: 340,
  },
  journeyCard: {
    backgroundColor: colors.white,
    borderColor: colors.track,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    gap: 12,
    padding: 16,
    ...tabi.shadow,
  },
  journeyHeading: {
    alignItems: "center",
    borderBottomColor: colors.track,
    borderBottomWidth: 1,
    flexDirection: "row",
    justifyContent: "space-between",
    paddingBottom: 10,
  },
  boardLabel: {
    color: colors.quiet,
    fontFamily: "monospace",
    fontSize: 10,
    fontWeight: "700",
    letterSpacing: 1.1,
  },
  boardTime: {
    color: colors.signal,
    fontFamily: "monospace",
    fontSize: 11,
    fontWeight: "700",
    letterSpacing: 0.8,
  },
  endpointGroup: { flexDirection: "row", gap: 12 },
  endpointMarker: { alignItems: "center", paddingTop: 5, width: 18 },
  originDot: {
    backgroundColor: colors.midnight,
    borderRadius: 6,
    height: 12,
    width: 12,
  },
  destinationDot: {
    backgroundColor: colors.route,
    borderColor: colors.midnight,
    borderRadius: 6,
    borderWidth: 2,
    height: 12,
    width: 12,
  },
  endpointBody: { flex: 1, gap: 4 },
  endpointLabel: {
    color: colors.quiet,
    fontFamily: "monospace",
    fontSize: 10,
    fontWeight: "700",
    letterSpacing: 1,
  },
  endpointValue: {
    color: colors.ink,
    fontFamily: tabi.type.display,
    fontSize: 21,
    fontWeight: "700",
    letterSpacing: -0.25,
  },
  endpointOptions: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: 6,
    marginTop: 3,
  },
  endpointOption: {
    borderColor: colors.track,
    borderRadius: tabi.radius.pill,
    borderWidth: 1,
    paddingHorizontal: 8,
    paddingVertical: 6,
  },
  endpointOptionSelected: {
    backgroundColor: colors.midnight,
    borderColor: colors.midnight,
  },
  endpointOptionText: { color: colors.ink, fontSize: 12, fontWeight: "600" },
  endpointOptionTextSelected: { color: colors.white },
  swapButton: {
    alignItems: "center",
    alignSelf: "flex-start",
    flexDirection: "row",
    gap: 7,
    marginLeft: 25,
    paddingHorizontal: 4,
    paddingVertical: 2,
  },
  swapButtonText: { color: colors.signal, fontSize: 21, fontWeight: "700" },
  swapButtonLabel: {
    color: colors.signal,
    fontFamily: "monospace",
    fontSize: 10,
    fontWeight: "700",
    letterSpacing: 0.9,
  },
  controls: { gap: 10 },
  controlRow: { flexDirection: "row", flexWrap: "wrap", gap: 8 },
  choice: {
    alignItems: "center",
    borderColor: colors.track,
    borderWidth: 1,
    borderRadius: tabi.radius.pill,
    flexDirection: "row",
    gap: 7,
    paddingHorizontal: 11,
    paddingVertical: 9,
  },
  choiceActive: { backgroundColor: "#FFF0E9", borderColor: colors.signal },
  choiceMark: { color: colors.signal, fontSize: 16, fontWeight: "800" },
  choiceText: { color: colors.ink, fontSize: 14, fontWeight: "600" },
  permissionNote: {
    borderLeftColor: colors.signal,
    borderLeftWidth: 3,
    gap: 3,
    paddingLeft: 10,
  },
  permissionTitle: { color: colors.ink, fontSize: 14, fontWeight: "700" },
  permissionCopy: { color: colors.quiet, fontSize: 13, lineHeight: 18 },
  locationButton: { alignSelf: "flex-start", paddingVertical: 5 },
  locationButtonText: {
    color: colors.quiet,
    fontFamily: "monospace",
    fontSize: 10,
    fontWeight: "700",
    letterSpacing: 0.8,
    textDecorationLine: "underline",
  },
  primaryButton: {
    alignItems: "center",
    backgroundColor: colors.signal,
    borderRadius: tabi.radius.medium,
    flexDirection: "row",
    justifyContent: "space-between",
    minHeight: 56,
    paddingHorizontal: 18,
  },
  primaryPressed: { backgroundColor: "#C94216" },
  primaryButtonText: {
    color: colors.white,
    fontFamily: tabi.type.display,
    fontSize: 19,
    fontWeight: "800",
    letterSpacing: 0.3,
  },
  primaryButtonArrow: { color: colors.white, fontSize: 26, fontWeight: "500" },
  statusMessage: { color: colors.quiet, fontSize: 13, lineHeight: 18 },
  shareButton: {
    alignItems: "center",
    alignSelf: "flex-start",
    flexDirection: "row",
    gap: 7,
    minHeight: tabi.touchTarget,
  },
  shareButtonText: { color: colors.signal, fontSize: 14, fontWeight: "700" },
  itinerary: {
    backgroundColor: colors.midnight,
    borderRadius: tabi.radius.medium,
    gap: 10,
    padding: 16,
  },
  itineraryTopline: {
    alignItems: "flex-start",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  duration: {
    color: colors.white,
    fontFamily: tabi.type.display,
    fontSize: 36,
    fontWeight: "800",
    letterSpacing: -1,
  },
  arrivalBlock: { alignItems: "flex-end", gap: 3 },
  arrivalTime: {
    color: colors.route,
    fontFamily: "monospace",
    fontSize: 16,
    fontWeight: "700",
  },
  tripFacts: { color: "#DCE3E9", fontSize: 14, fontWeight: "600" },
  legs: { gap: 2, marginTop: 3 },
  legRow: { flexDirection: "row", minHeight: 76 },
  legRail: { alignItems: "center", paddingTop: 5, width: 20 },
  transitStop: {
    backgroundColor: colors.route,
    borderRadius: 5,
    height: 10,
    width: 10,
  },
  walkStop: {
    backgroundColor: colors.midnight,
    borderColor: colors.white,
    borderRadius: 5,
    borderWidth: 2,
    height: 10,
    width: 10,
  },
  railLine: {
    backgroundColor: "#718094",
    flex: 1,
    marginBottom: -4,
    marginTop: 4,
    width: 2,
  },
  legContent: { flex: 1, gap: 2, paddingBottom: 13 },
  legHeading: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  routeBadge: {
    backgroundColor: colors.route,
    color: colors.midnight,
    fontFamily: tabi.type.display,
    fontSize: 13,
    fontWeight: "800",
    maxWidth: "82%",
    paddingHorizontal: 6,
    paddingVertical: 2,
  },
  walkBadge: {
    color: colors.white,
    fontFamily: "monospace",
    fontSize: 11,
    fontWeight: "700",
    letterSpacing: 0.6,
  },
  legTime: { color: "#DCE3E9", fontFamily: "monospace", fontSize: 11 },
  legPlace: { color: colors.white, fontSize: 15, fontWeight: "700" },
  legDirection: { color: "#DCE3E9", fontSize: 13 },
  directionsButton: {
    alignSelf: "flex-start",
    borderBottomColor: colors.route,
    borderBottomWidth: 1,
    marginTop: 5,
    paddingBottom: 2,
  },
  directionsButtonText: {
    color: colors.route,
    fontFamily: "monospace",
    fontSize: 10,
    fontWeight: "700",
    letterSpacing: 0.6,
  },
  freshness: {
    borderTopColor: "#526273",
    borderTopWidth: 1,
    color: "#DCE3E9",
    fontSize: 11,
    lineHeight: 15,
    paddingTop: 9,
  },
  pressed: { opacity: 0.72 },
});
