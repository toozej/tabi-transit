import { StyleSheet, Text, View } from "react-native";
import MaterialCommunityIcons from "@expo/vector-icons/MaterialCommunityIcons";

import type { VehicleHistory as VehicleHistoryData } from "@/domain/vehicleModels";
import { tabi } from "@/ui/tabi";

import {
  ADHERENCE_UNAVAILABLE,
  HISTORY_EMPTY,
  HISTORY_UNAVAILABLE,
  formatHistoryObservation,
} from "./vehicleHistoryPresentation";

type Props = {
  history?: VehicleHistoryData;
  isError: boolean;
  isLoading: boolean;
};

/**
 * Text comes first so a rider can use the retained timeline without a map.
 * Adherence is explicitly unavailable until historical trip-update evidence
 * exists; do not infer it from these coordinates (ADR-0016).
 */
export function VehicleHistory({ history, isError, isLoading }: Props) {
  return (
    <View accessibilityLabel="Vehicle history" style={styles.container}>
      <Text accessibilityRole="header" style={styles.heading}>
        Vehicle history
      </Text>
      <Text style={styles.copy}>
        Normalized observations retained for up to 30 days.
      </Text>
      {isLoading && (
        <Text accessibilityLiveRegion="polite">Loading vehicle history.</Text>
      )}
      {isError && <Text accessibilityRole="alert">{HISTORY_UNAVAILABLE}</Text>}
      {!isLoading && !isError && history?.observations.length === 0 && (
        <Text accessibilityLiveRegion="polite">{HISTORY_EMPTY}</Text>
      )}
      {!isLoading && !isError && history && history.observations.length > 0 && (
        <>
          <Text
            style={styles.meta}
          >{`History source · ${history.freshness.source}`}</Text>
          <View
            accessibilityLabel="Vehicle history timeline"
            style={styles.timeline}
          >
            {history.observations.map((observation, index) => (
              <View
                key={`${observation.observedAt}-${index}`}
                style={styles.observation}
              >
                <MaterialCommunityIcons
                  name="map-marker-outline"
                  color={tabi.color.accent}
                  size={16}
                />
                <Text style={styles.observationText}>
                  {formatHistoryObservation(observation)}
                </Text>
              </View>
            ))}
          </View>
        </>
      )}
      <Text accessibilityRole="alert" style={styles.meta}>
        {ADHERENCE_UNAVAILABLE}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    borderTopColor: tabi.color.border,
    borderTopWidth: StyleSheet.hairlineWidth,
    gap: 6,
    marginTop: 8,
    paddingTop: 12,
  },
  heading: { color: tabi.color.ink, fontSize: 17, fontWeight: "800" },
  copy: { color: tabi.color.mutedInk, fontSize: 13, lineHeight: 18 },
  meta: { color: tabi.color.mutedInk, fontSize: 11, lineHeight: 16 },
  timeline: { gap: 5, paddingVertical: 4 },
  observation: { alignItems: "flex-start", flexDirection: "row", gap: 6 },
  observationText: {
    color: tabi.color.ink,
    flex: 1,
    fontSize: 12,
    lineHeight: 17,
  },
});
