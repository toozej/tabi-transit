import { useMemo } from "react";
import { FlatList, StyleSheet, Text, View } from "react-native";

import { createSyntheticFleet } from "@/domain/vehicles";
import { getMapboxAccessToken } from "@/maps/config";
import { VehicleMap } from "@/maps/VehicleMap";

export function VehicleMapSpike() {
  const vehicles = useMemo(() => createSyntheticFleet(), []);
  const mapAvailable = getMapboxAccessToken() !== undefined;

  if (mapAvailable) {
    return <VehicleMap />;
  }

  return (
    <View style={styles.container}>
      <Text accessibilityRole="header" style={styles.heading}>
        Synthetic vehicle compatibility spike
      </Text>
      <Text style={styles.message}>
        Map rendering is disabled until a restricted public Maps SDK token is
        supplied through local configuration.
      </Text>
      <FlatList
        accessibilityLabel="Synthetic vehicles"
        data={vehicles.slice(0, 25)}
        keyExtractor={(vehicle) => vehicle.id}
        renderItem={({ item }) => (
          <Text>{`${item.id}: ${item.routeId}, ${item.freshness}`}</Text>
        )}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, gap: 12, padding: 20 },
  heading: { fontSize: 22, fontWeight: "600" },
  message: { lineHeight: 20 },
});
