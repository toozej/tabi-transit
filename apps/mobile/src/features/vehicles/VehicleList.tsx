import MaterialCommunityIcons from "@expo/vector-icons/MaterialCommunityIcons";
import { Pressable, StyleSheet, Text, View } from "react-native";

import { formatFreshness, type Vehicle } from "@/domain/vehicleModels";
import { tabi, tabiCommonStyles } from "@/ui/tabi";

type Props = {
  vehicles: readonly Vehicle[];
  selectedVehicleId?: string;
  onSelect: (id: string) => void;
};

export function VehicleList({ vehicles, selectedVehicleId, onSelect }: Props) {
  if (vehicles.length === 0) {
    return (
      <Text accessibilityLiveRegion="polite" style={styles.empty}>
        No vehicles match these filters.
      </Text>
    );
  }

  return (
    <View accessibilityLabel="Vehicle list" style={styles.list}>
      <Text accessibilityRole="header" style={styles.heading}>
        Nearby fleet
      </Text>
      <View style={styles.card}>
        {vehicles.map((vehicle) => {
          const selected = selectedVehicleId === vehicle.id;
          return (
            <Pressable
              key={vehicle.id}
              accessibilityRole="button"
              accessibilityState={{ selected }}
              accessibilityLabel={`Select vehicle ${vehicle.sourceVehicleId}, ${vehicle.mode}, ${formatFreshness(vehicle.freshness)}`}
              onPress={() => onSelect(vehicle.id)}
              style={({ pressed }) => [
                styles.row,
                selected && styles.rowSelected,
                pressed && tabiCommonStyles.pressed,
              ]}
            >
              <View
                style={[styles.modeIcon, selected && styles.modeIconSelected]}
              >
                <MaterialCommunityIcons
                  name={vehicle.mode === "light_rail" ? "train" : "bus"}
                  color={selected ? tabi.color.white : tabi.color.bus}
                  size={20}
                />
              </View>
              <View style={styles.copy}>
                <Text
                  style={styles.title}
                >{`Vehicle ${vehicle.sourceVehicleId}`}</Text>
                <Text
                  style={styles.meta}
                >{`Route ${vehicle.routeId ?? "unavailable"} · ${formatFreshness(vehicle.freshness)}`}</Text>
              </View>
              <MaterialCommunityIcons
                name="chevron-right"
                color={tabi.color.faintInk}
                size={21}
              />
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  list: { gap: 9 },
  heading: { color: tabi.color.ink, fontSize: 18, fontWeight: "800" },
  empty: { color: tabi.color.mutedInk, fontSize: 14, lineHeight: 20 },
  card: {
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    overflow: "hidden",
  },
  row: {
    alignItems: "center",
    flexDirection: "row",
    gap: 11,
    minHeight: 64,
    paddingHorizontal: 12,
  },
  rowSelected: { backgroundColor: tabi.color.accentSoft },
  modeIcon: {
    alignItems: "center",
    backgroundColor: "#E8F0F7",
    borderRadius: tabi.radius.pill,
    height: 38,
    justifyContent: "center",
    width: 38,
  },
  modeIconSelected: { backgroundColor: tabi.color.accent },
  copy: { flex: 1, gap: 3 },
  title: { color: tabi.color.ink, fontSize: 15, fontWeight: "700" },
  meta: { color: tabi.color.mutedInk, fontSize: 11 },
});
