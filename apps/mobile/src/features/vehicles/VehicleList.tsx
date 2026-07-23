import { Pressable, StyleSheet, Text, View } from "react-native";

import { formatFreshness, type Vehicle } from "@/domain/vehicleModels";

type Props = { vehicles: readonly Vehicle[]; onSelect: (id: string) => void };

export function VehicleList({ vehicles, onSelect }: Props) {
  if (vehicles.length === 0) {
    return (
      <Text accessibilityLiveRegion="polite">
        No vehicles match the selected filters.
      </Text>
    );
  }
  return (
    <View accessibilityLabel="Vehicle list">
      <Text accessibilityRole="header" style={styles.heading}>
        Vehicle list
      </Text>
      {vehicles.map((vehicle) => (
        <Pressable
          key={vehicle.id}
          accessibilityRole="button"
          accessibilityLabel={`Select vehicle ${vehicle.sourceVehicleId}, ${vehicle.mode}, ${formatFreshness(vehicle.freshness)}`}
          onPress={() => onSelect(vehicle.id)}
          style={styles.row}
        >
          <Text>{`Vehicle ${vehicle.sourceVehicleId} · ${vehicle.routeId ?? "No route"}`}</Text>
          <Text>{formatFreshness(vehicle.freshness)}</Text>
        </Pressable>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  heading: { fontSize: 20, fontWeight: "600" },
  row: { gap: 2, paddingVertical: 10 },
});
