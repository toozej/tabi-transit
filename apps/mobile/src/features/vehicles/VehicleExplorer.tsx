import { useState } from "react";
import {
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";

import {
  filterVehicles,
  formatFreshness,
  type VehicleMode,
} from "@/domain/vehicleModels";
import { getMapboxAccessToken } from "@/maps/config";
import { VehicleMap } from "@/maps/VehicleMap";
import { useMapUiStore } from "@/state/mapUiStore";

import { VehicleList } from "./VehicleList";
import { VehicleHistory } from "./VehicleHistory";
import {
  useVehicleDetail,
  useVehicleHistory,
  useVehicleSearch,
  useVehicles,
} from "./queries";

const MODES: readonly VehicleMode[] = [
  "bus",
  "light_rail",
  "streetcar",
  "commuter_rail",
];

export function VehicleExplorer() {
  const [search, setSearch] = useState("");
  const [isPanelExpanded, setIsPanelExpanded] = useState(false);
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
  const historyQuery = useVehicleHistory(selectedVehicleId);
  const collection = vehiclesQuery.data;
  const vehicles = filterVehicles(collection?.vehicles ?? [], filters);
  const mapAvailable = getMapboxAccessToken() !== undefined;
  const selected =
    detailQuery.data ?? vehicles.find((item) => item.id === selectedVehicleId);

  const loading = vehiclesQuery.isLoading || !collection;
  const unavailable = vehiclesQuery.isError;

  return (
    <View style={styles.screen}>
      {mapAvailable ? (
        <VehicleMap vehicles={vehicles} selectedVehicleId={selectedVehicleId} />
      ) : (
        <View style={styles.mapUnavailable}>
          <Text accessibilityLiveRegion="polite">
            Map rendering is unavailable without an approved public Maps SDK
            token. The vehicle list remains available.
          </Text>
        </View>
      )}

      <SafeAreaView edges={["top"]} style={styles.topOverlay}>
        <View style={styles.statusCard}>
          <Text accessibilityRole="header" style={styles.mapTitle}>
            Vehicles
          </Text>
          {loading ? (
            <Text accessibilityLiveRegion="polite">
              Loading vehicle positions.
            </Text>
          ) : unavailable ? (
            <Text accessibilityRole="alert">
              Vehicle positions are unavailable. Existing data has not been
              presented as live.
            </Text>
          ) : (
            <Text accessibilityLiveRegion="polite">
              {formatFreshness(collection.freshness)}
            </Text>
          )}
        </View>
      </SafeAreaView>

      {!loading && !unavailable && collection.freshness.status === "stale" && (
        <View style={styles.staleNotice}>
          <Text accessibilityRole="alert">
            Vehicle positions are stale and are not live.
          </Text>
        </View>
      )}

      {!isPanelExpanded && !loading && !unavailable && (
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Browse vehicle controls and list"
          onPress={() => setIsPanelExpanded(true)}
          style={styles.openPanelButton}
        >
          <Text
            style={styles.openPanelText}
          >{`Vehicles (${vehicles.length})`}</Text>
        </Pressable>
      )}

      {isPanelExpanded && !loading && !unavailable && (
        <View
          style={styles.sheet}
          accessibilityLabel="Vehicle controls and list"
        >
          <View style={styles.sheetHandle} />
          <View style={styles.sheetHeader}>
            <Text accessibilityRole="header" style={styles.heading}>
              Vehicles
            </Text>
            <Pressable
              accessibilityRole="button"
              accessibilityLabel="Hide vehicle controls and list"
              onPress={() => setIsPanelExpanded(false)}
              hitSlop={10}
            >
              <Text style={styles.hidePanelText}>Hide</Text>
            </Pressable>
          </View>
          <ScrollView
            contentContainerStyle={styles.panelContent}
            keyboardShouldPersistTaps="handled"
          >
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
                  setFreshness(
                    filters.freshness === "fresh" ? undefined : "fresh",
                  )
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
              style={styles.searchInput}
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
                <VehicleHistory
                  history={historyQuery.data}
                  isError={historyQuery.isError}
                  isLoading={historyQuery.isLoading}
                />
              </View>
            )}
          </ScrollView>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: "#E5E7EB" },
  mapUnavailable: {
    alignItems: "center",
    flex: 1,
    justifyContent: "center",
    padding: 24,
  },
  topOverlay: { left: 0, position: "absolute", right: 0, top: 0 },
  statusCard: {
    backgroundColor: "rgba(255, 255, 255, 0.94)",
    borderRadius: 14,
    gap: 2,
    marginHorizontal: 16,
    marginTop: 8,
    padding: 12,
  },
  mapTitle: { fontSize: 20, fontWeight: "600" },
  staleNotice: {
    backgroundColor: "rgba(254, 242, 242, 0.96)",
    borderRadius: 12,
    left: 16,
    padding: 10,
    position: "absolute",
    right: 16,
    top: 108,
  },
  openPanelButton: {
    alignSelf: "center",
    backgroundColor: "#111827",
    borderRadius: 24,
    bottom: 18,
    paddingHorizontal: 20,
    paddingVertical: 13,
    position: "absolute",
  },
  openPanelText: { color: "#FFFFFF", fontSize: 16, fontWeight: "600" },
  sheet: {
    backgroundColor: "#FFFFFF",
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    bottom: 0,
    elevation: 12,
    left: 0,
    maxHeight: "68%",
    paddingTop: 8,
    position: "absolute",
    right: 0,
    shadowColor: "#000000",
    shadowOffset: { width: 0, height: -4 },
    shadowOpacity: 0.16,
    shadowRadius: 12,
  },
  sheetHandle: {
    alignSelf: "center",
    backgroundColor: "#9CA3AF",
    borderRadius: 2,
    height: 4,
    width: 42,
  },
  sheetHeader: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
    paddingHorizontal: 20,
    paddingTop: 10,
  },
  panelContent: { gap: 12, padding: 20, paddingTop: 12 },
  heading: { fontSize: 24, fontWeight: "600" },
  hidePanelText: { color: "#2563EB", fontSize: 16, fontWeight: "600" },
  filters: { flexDirection: "row", flexWrap: "wrap", gap: 10 },
  searchInput: {
    borderBottomWidth: 1,
    borderColor: "#9CA3AF",
    paddingVertical: 8,
  },
  detail: { gap: 4, paddingTop: 8 },
});
