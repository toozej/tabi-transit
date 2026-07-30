import { StyleSheet, Text, View } from "react-native";

import type { VehicleHistory as VehicleHistoryData } from "@/domain/vehicleModels";

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
      <Text>Normalized observations retained for up to 30 days.</Text>
      {isLoading && (
        <Text accessibilityLiveRegion="polite">Loading vehicle history.</Text>
      )}
      {isError && <Text accessibilityRole="alert">{HISTORY_UNAVAILABLE}</Text>}
      {!isLoading && !isError && history?.observations.length === 0 && (
        <Text accessibilityLiveRegion="polite">{HISTORY_EMPTY}</Text>
      )}
      {!isLoading && !isError && history && history.observations.length > 0 && (
        <>
          <Text>{`History source: ${history.freshness.source}`}</Text>
          <View accessibilityLabel="Vehicle history timeline">
            {history.observations.map((observation, index) => (
              <Text key={`${observation.observedAt}-${index}`}>
                {formatHistoryObservation(observation)}
              </Text>
            ))}
          </View>
        </>
      )}
      <Text accessibilityRole="alert">{ADHERENCE_UNAVAILABLE}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: 4, paddingTop: 12 },
  heading: { fontSize: 20, fontWeight: "600" },
});
