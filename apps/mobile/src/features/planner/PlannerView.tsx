import { useState } from "react";
import {
  Linking,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from "react-native";

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

function EndpointPicker({
  label,
  value,
  onSelect,
}: {
  label: string;
  value?: PlanEndpoint;
  onSelect: (endpoint: PlanEndpoint) => void;
}) {
  return (
    <View accessibilityLabel={`${label} picker`} style={styles.group}>
      <Text accessibilityRole="header" style={styles.subheading}>
        {label}
      </Text>
      <Text>{value?.label ?? `Choose a ${label.toLowerCase()}.`}</Text>
      {fixturePlanEndpoints.map((endpoint) => (
        <Pressable
          key={`${label}:${endpoint.id}`}
          accessibilityRole="button"
          accessibilityLabel={`Set ${label.toLowerCase()} to ${endpoint.label}`}
          onPress={() => onSelect(endpoint)}
          style={styles.button}
        >
          <Text>{endpoint.label}</Text>
        </Pressable>
      ))}
    </View>
  );
}

function ItineraryTimeline({ itinerary }: { itinerary: Itinerary }) {
  return (
    <View
      accessibilityLabel={`Itinerary taking ${Math.round(itinerary.durationSeconds / 60)} minutes`}
      style={styles.itinerary}
    >
      <Text accessibilityRole="header" style={styles.subheading}>
        {`${Math.round(itinerary.durationSeconds / 60)} min · ${itinerary.transfers} transfer${itinerary.transfers === 1 ? "" : "s"} · ${itinerary.walkingMeters} m walking`}
      </Text>
      {itinerary.legs.map((leg) => {
        const walkingDirections = createExternalWalkingDirectionsLink(leg);
        return (
          <View key={leg.id} style={styles.leg}>
            <Text accessibilityRole="header">
              {leg.routeLabel ?? (leg.mode === "walk" ? "Walk" : leg.mode)}
            </Text>
            <Text>{`${leg.startLabel} to ${leg.endLabel}`}</Text>
            {leg.headsign && <Text>{`Toward ${leg.headsign}`}</Text>}
            <Text>{`Status: ${leg.realtime}`}</Text>
            {leg.geometry && (
              <Text>
                Map geometry is prepared for this leg; this text timeline is an
                equivalent way to follow it.
              </Text>
            )}
            {walkingDirections && (
              <Pressable
                accessibilityRole="link"
                accessibilityLabel={`Open walking directions from ${leg.startLabel} to ${leg.endLabel} in your maps app`}
                onPress={() => void Linking.openURL(walkingDirections)}
                style={styles.button}
              >
                <Text>Open walking directions</Text>
              </Pressable>
            )}
          </View>
        );
      })}
      <Text>{itinerary.freshness.message}</Text>
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

  return (
    <ScrollView contentContainerStyle={styles.page}>
      <Text accessibilityRole="header" style={styles.heading}>
        Plan a trip
      </Text>
      <Text>
        This planner never calls Mapbox or a transit provider directly.
      </Text>
      {locationPermission === "denied" ? (
        <Text accessibilityRole="alert">
          Location access is off. Choose stops, saved places, or a map pin
          instead; Tabi will not ask again here.
        </Text>
      ) : (
        <Pressable
          accessibilityRole="button"
          onPress={() => setLocationPermission("denied")}
          style={styles.button}
        >
          <Text>Continue without current location</Text>
        </Pressable>
      )}
      <EndpointPicker
        label="Origin"
        value={draft.origin}
        onSelect={setOrigin}
      />
      <Pressable
        accessibilityRole="button"
        onPress={swap}
        style={styles.button}
      >
        <Text>Swap origin and destination</Text>
      </Pressable>
      <EndpointPicker
        label="Destination"
        value={draft.destination}
        onSelect={setDestination}
      />
      <View style={styles.group}>
        <Text accessibilityRole="header" style={styles.subheading}>
          Constraints
        </Text>
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
          style={styles.button}
        >
          <Text>Wheelchair-accessible itineraries</Text>
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
          style={styles.button}
        >
          <Text>{`Optimize for ${draft.constraints.optimization === "fastest" ? "fastest trip" : "fewer transfers"}`}</Text>
        </Pressable>
      </View>
      <Pressable
        accessibilityRole="button"
        onPress={plan}
        style={styles.primaryButton}
      >
        <Text>Plan with fixture data</Text>
      </Pressable>
      <Text accessibilityLiveRegion="polite">{message}</Text>
      {draft.origin && draft.destination && (
        <Text
          selectable
        >{`Safe planning link: ${createPlanDeepLink(draft)}`}</Text>
      )}
      {itineraries.map((itinerary) => (
        <ItineraryTimeline key={itinerary.id} itinerary={itinerary} />
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  page: { gap: 12, padding: 20 },
  heading: { fontSize: 24, fontWeight: "600" },
  subheading: { fontSize: 19, fontWeight: "600" },
  group: { gap: 8 },
  button: { borderWidth: 1, borderColor: "#444", padding: 10 },
  primaryButton: { backgroundColor: "#1d4ed8", padding: 12 },
  itinerary: { borderWidth: 1, borderColor: "#777", gap: 8, padding: 12 },
  leg: {
    borderLeftWidth: 3,
    borderLeftColor: "#1d4ed8",
    gap: 3,
    paddingLeft: 9,
  },
});
