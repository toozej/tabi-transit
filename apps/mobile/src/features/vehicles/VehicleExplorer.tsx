import { useState } from "react";
import {
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import {
  filterVehicles,
  formatFreshness,
  type VehicleMode,
} from "@/domain/vehicleModels";
import { getMapboxAccessToken } from "@/maps/config";
import { VehicleMap } from "@/maps/VehicleMap";
import { useMapUiStore } from "@/state/mapUiStore";

import { VehicleList } from "./VehicleList";
import { useVehicleDetail, useVehicleSearch, useVehicles } from "./queries";

const MODES: readonly VehicleMode[] = [
  "bus",
  "light_rail",
  "streetcar",
  "commuter_rail",
];

export function VehicleExplorer() {
  const [search, setSearch] = useState("");
  const {
    filters,
    selectedVehicleId,
    setFreshness,
    setModeEnabled,
    setSelectedVehicleId,
  } = useMapUiStore();
  const vehiclesQuery = useVehicles(filters);
  const searchQuery = useVehicleSearch(search);
  const detailQuery = useVehicleDetail(selectedVehicleId);
  const collection = vehiclesQuery.data;
  const vehicles = filterVehicles(collection?.vehicles ?? [], filters);
  const mapAvailable = getMapboxAccessToken() !== undefined;
  const selected =
    detailQuery.data ?? vehicles.find((item) => item.id === selectedVehicleId);

  if (vehiclesQuery.isLoading || !collection)
    return (
      <Text accessibilityLiveRegion="polite">Loading vehicle positions.</Text>
    );
  if (vehiclesQuery.isError) {
    return (
      <Text accessibilityRole="alert">
        Vehicle positions are unavailable. Existing data has not been presented
        as live.
      </Text>
    );
  }

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text accessibilityRole="header" style={styles.heading}>
        Vehicles
      </Text>
      <Text accessibilityLiveRegion="polite">
        {formatFreshness(collection.freshness)}
      </Text>
      {collection.freshness.status === "stale" && (
        <Text accessibilityRole="alert">
          Vehicle positions are stale and are not live.
        </Text>
      )}
      {!mapAvailable && (
        <Text accessibilityLiveRegion="polite">
          Map rendering is unavailable without an approved public Maps SDK
          token. The vehicle list remains available.
        </Text>
      )}
      <View accessibilityLabel="Vehicle filters" style={styles.filters}>
        {MODES.map((mode) => {
          const selectedMode = filters.modes.includes(mode);
          return (
            <Pressable
              key={mode}
              accessibilityRole="button"
              accessibilityState={{ selected: selectedMode }}
              accessibilityLabel={`${selectedMode ? "Hide" : "Show"} ${mode} vehicles`}
              onPress={() => setModeEnabled(mode, !selectedMode)}
            >
              <Text>{mode}</Text>
            </Pressable>
          );
        })}
        <Pressable
          accessibilityRole="button"
          onPress={() =>
            setFreshness(filters.freshness === "fresh" ? undefined : "fresh")
          }
        >
          <Text>Fresh only</Text>
        </Pressable>
      </View>
      <TextInput
        accessibilityLabel="Search exact vehicle ID"
        value={search}
        onChangeText={setSearch}
        placeholder="Vehicle ID"
        autoCapitalize="none"
      />
      {search.trim() !== "" && (
        <View accessibilityLabel="Vehicle search results">
          <Text accessibilityRole="header">Search results</Text>
          {(searchQuery.data?.vehicles ?? []).map((vehicle) => (
            <Pressable
              key={vehicle.id}
              accessibilityRole="button"
              accessibilityLabel={`Select search result ${vehicle.sourceVehicleId}`}
              onPress={() => setSelectedVehicleId(vehicle.id)}
            >
              <Text>{`Vehicle ${vehicle.sourceVehicleId}`}</Text>
            </Pressable>
          ))}
        </View>
      )}
      {mapAvailable && (
        <View style={styles.map}>
          <VehicleMap
            vehicles={vehicles}
            selectedVehicleId={selectedVehicleId}
          />
        </View>
      )}
      <VehicleList vehicles={vehicles} onSelect={setSelectedVehicleId} />
      {selected && (
        <View
          accessibilityLabel="Selected vehicle details"
          style={styles.detail}
        >
          <Text accessibilityRole="header">{`Vehicle ${selected.sourceVehicleId}`}</Text>
          <Text>{`Route: ${selected.routeId ?? "Unavailable"}`}</Text>
          <Text>{`Source: ${selected.freshness.source}`}</Text>
          <Text>{formatFreshness(selected.freshness)}</Text>
          {selected.freshness.status !== "fresh" && (
            <Text accessibilityRole="alert">
              This vehicle is not shown as live.
            </Text>
          )}
        </View>
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: { gap: 12, padding: 20 },
  heading: { fontSize: 24, fontWeight: "600" },
  filters: { flexDirection: "row", flexWrap: "wrap", gap: 10 },
  map: { height: 340 },
  detail: { gap: 4, paddingTop: 8 },
});
