import MaterialCommunityIcons from "@expo/vector-icons/MaterialCommunityIcons";
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
  type Vehicle,
  type VehicleMode,
} from "@/domain/vehicleModels";
import { getMapboxAccessToken } from "@/maps/config";
import { VehicleMap } from "@/maps/VehicleMap";
import { useMapUiStore } from "@/state/mapUiStore";
import { tabi, tabiCommonStyles } from "@/ui/tabi";

import { VehicleHistory } from "./VehicleHistory";
import { VehicleList } from "./VehicleList";
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

const modeLabels: Record<VehicleMode, string> = {
  bus: "Bus",
  light_rail: "Light rail",
  streetcar: "Streetcar",
  commuter_rail: "Commuter rail",
  ferry: "Ferry",
  unknown: "Other",
};

function displayFreshness(freshness: Vehicle["freshness"]) {
  const age = Math.round(freshness.ageSeconds);
  const prefix =
    freshness.status === "fresh" && freshness.isRealtime
      ? "Live positions"
      : freshness.status === "stale"
        ? "Last known positions"
        : "Latest available positions";
  return `${prefix} · updated ${age} second${age === 1 ? "" : "s"} ago`;
}

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
          <View style={styles.mapUnavailableIcon}>
            <MaterialCommunityIcons
              name="map-outline"
              color={tabi.color.accent}
              size={34}
            />
          </View>
          <Text style={styles.mapUnavailableTitle}>
            Map preview unavailable
          </Text>
          <Text
            accessibilityLiveRegion="polite"
            style={styles.mapUnavailableCopy}
          >
            You can still browse every vehicle in the accessible list.
          </Text>
        </View>
      )}

      <SafeAreaView
        edges={["top"]}
        pointerEvents="box-none"
        style={styles.topOverlay}
      >
        <View style={styles.statusCard}>
          <View style={styles.mapTitleRow}>
            <View>
              <Text style={styles.eyebrow}>LIVE NETWORK</Text>
              <Text accessibilityRole="header" style={styles.mapTitle}>
                Vehicles
              </Text>
            </View>
            {!loading && !unavailable && (
              <View style={styles.livePill}>
                <View style={styles.liveDot} />
                <Text style={styles.liveText}>{vehicles.length} shown</Text>
              </View>
            )}
          </View>
          {loading ? (
            <Text accessibilityLiveRegion="polite" style={styles.statusText}>
              Loading vehicle positions…
            </Text>
          ) : unavailable ? (
            <Text accessibilityRole="alert" style={styles.errorText}>
              Vehicle positions are unavailable. Existing data is not presented
              as live.
            </Text>
          ) : (
            <Text accessibilityLiveRegion="polite" style={styles.statusText}>
              {displayFreshness(collection.freshness)}
            </Text>
          )}
        </View>
      </SafeAreaView>

      {!loading && !unavailable && collection.freshness.status === "stale" && (
        <View style={styles.staleNotice}>
          <MaterialCommunityIcons
            name="clock-alert-outline"
            color={tabi.color.warning}
            size={18}
          />
          <Text accessibilityRole="alert" style={styles.staleText}>
            Positions are stale and are not live.
          </Text>
        </View>
      )}

      {!isPanelExpanded && !loading && !unavailable && (
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="Browse vehicle controls and list"
          onPress={() => setIsPanelExpanded(true)}
          style={({ pressed }) => [
            styles.openPanelButton,
            pressed && styles.openPanelButtonPressed,
          ]}
        >
          <MaterialCommunityIcons
            name="format-list-bulleted"
            color={tabi.color.white}
            size={20}
          />
          <Text
            style={styles.openPanelText}
          >{`Browse ${vehicles.length} vehicles`}</Text>
        </Pressable>
      )}

      {isPanelExpanded && !loading && !unavailable && (
        <View
          style={styles.sheet}
          accessibilityLabel="Vehicle controls and list"
        >
          <View style={styles.sheetHandle} />
          <View style={styles.sheetHeader}>
            <View>
              <Text style={styles.eyebrow}>MAP CONTROLS</Text>
              <Text accessibilityRole="header" style={styles.heading}>
                Browse vehicles
              </Text>
            </View>
            <Pressable
              accessibilityRole="button"
              accessibilityLabel="Hide vehicle controls and list"
              onPress={() => setIsPanelExpanded(false)}
              hitSlop={10}
              style={({ pressed }) => [
                styles.closeButton,
                pressed && tabiCommonStyles.pressed,
              ]}
            >
              <MaterialCommunityIcons
                name="chevron-down"
                color={tabi.color.ink}
                size={25}
              />
            </Pressable>
          </View>

          <ScrollView
            contentContainerStyle={styles.panelContent}
            keyboardShouldPersistTaps="handled"
            showsVerticalScrollIndicator={false}
          >
            <View accessibilityLabel="Vehicle filters" style={styles.filters}>
              {MODES.map((mode) => {
                const enabled = filters.modes.includes(mode);
                return (
                  <Pressable
                    key={mode}
                    accessibilityRole="button"
                    accessibilityState={{ selected: enabled }}
                    accessibilityLabel={`${enabled ? "Hide" : "Show"} ${modeLabels[mode]} vehicles`}
                    onPress={() => setModeEnabled(mode, !enabled)}
                    style={({ pressed }) => [
                      styles.filterChip,
                      enabled && styles.filterChipSelected,
                      pressed && tabiCommonStyles.pressed,
                    ]}
                  >
                    <Text
                      style={[
                        styles.filterChipText,
                        enabled && styles.filterChipTextSelected,
                      ]}
                    >
                      {modeLabels[mode]}
                    </Text>
                  </Pressable>
                );
              })}
              <Pressable
                accessibilityRole="button"
                accessibilityState={{ selected: filters.freshness === "fresh" }}
                onPress={() =>
                  setFreshness(
                    filters.freshness === "fresh" ? undefined : "fresh",
                  )
                }
                style={({ pressed }) => [
                  styles.filterChip,
                  filters.freshness === "fresh" && styles.filterChipSelected,
                  pressed && tabiCommonStyles.pressed,
                ]}
              >
                <Text
                  style={[
                    styles.filterChipText,
                    filters.freshness === "fresh" &&
                      styles.filterChipTextSelected,
                  ]}
                >
                  Live only
                </Text>
              </Pressable>
            </View>

            <View style={styles.searchBox}>
              <MaterialCommunityIcons
                name="magnify"
                color={tabi.color.mutedInk}
                size={21}
              />
              <TextInput
                accessibilityLabel="Search exact vehicle ID"
                value={search}
                onChangeText={setSearch}
                placeholder="Search vehicle ID"
                placeholderTextColor={tabi.color.faintInk}
                autoCapitalize="none"
                autoCorrect={false}
                returnKeyType="search"
                style={styles.searchInput}
              />
              {search.length > 0 && (
                <Pressable
                  accessibilityRole="button"
                  accessibilityLabel="Clear vehicle search"
                  onPress={() => setSearch("")}
                  hitSlop={8}
                >
                  <MaterialCommunityIcons
                    name="close-circle"
                    color={tabi.color.faintInk}
                    size={19}
                  />
                </Pressable>
              )}
            </View>

            {search.trim() !== "" && (
              <View
                accessibilityLabel="Vehicle search results"
                style={styles.searchResults}
              >
                <Text accessibilityRole="header" style={styles.miniHeading}>
                  Search results
                </Text>
                {(searchQuery.data?.vehicles ?? []).map((vehicle) => (
                  <Pressable
                    key={vehicle.id}
                    accessibilityRole="button"
                    accessibilityLabel={`Select search result ${vehicle.sourceVehicleId}`}
                    onPress={() => setSelectedVehicleId(vehicle.id)}
                    style={({ pressed }) => [
                      styles.searchResult,
                      pressed && tabiCommonStyles.pressed,
                    ]}
                  >
                    <MaterialCommunityIcons
                      name="bus"
                      color={tabi.color.accent}
                      size={19}
                    />
                    <Text
                      style={styles.searchResultText}
                    >{`Vehicle ${vehicle.sourceVehicleId}`}</Text>
                    <MaterialCommunityIcons
                      name="chevron-right"
                      color={tabi.color.faintInk}
                      size={20}
                    />
                  </Pressable>
                ))}
              </View>
            )}

            <VehicleList
              vehicles={vehicles}
              selectedVehicleId={selectedVehicleId}
              onSelect={setSelectedVehicleId}
            />

            {selected && (
              <View
                accessibilityLabel="Selected vehicle details"
                style={styles.detail}
              >
                <View style={styles.detailTitleRow}>
                  <View style={styles.vehicleIcon}>
                    <MaterialCommunityIcons
                      name="bus"
                      color={tabi.color.white}
                      size={21}
                    />
                  </View>
                  <View style={styles.detailTitleCopy}>
                    <Text accessibilityRole="header" style={styles.detailTitle}>
                      {`Vehicle ${selected.sourceVehicleId}`}
                    </Text>
                    <Text
                      style={styles.detailRoute}
                    >{`Route ${selected.routeId ?? "unavailable"}`}</Text>
                  </View>
                </View>
                <Text
                  style={styles.detailMeta}
                >{`Source · ${selected.freshness.source}`}</Text>
                <Text style={styles.detailMeta}>
                  {formatFreshness(selected.freshness)}
                </Text>
                {selected.freshness.status !== "fresh" && (
                  <Text accessibilityRole="alert" style={styles.errorText}>
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
  screen: { backgroundColor: tabi.color.surfaceMuted, flex: 1 },
  mapUnavailable: {
    alignItems: "center",
    flex: 1,
    justifyContent: "center",
    padding: 24,
  },
  mapUnavailableIcon: {
    alignItems: "center",
    backgroundColor: tabi.color.surface,
    borderRadius: tabi.radius.pill,
    height: 68,
    justifyContent: "center",
    width: 68,
  },
  mapUnavailableTitle: {
    color: tabi.color.ink,
    fontSize: 19,
    fontWeight: "800",
    marginTop: 14,
  },
  mapUnavailableCopy: {
    color: tabi.color.mutedInk,
    fontSize: 14,
    marginTop: 6,
    textAlign: "center",
  },
  topOverlay: { left: 0, position: "absolute", right: 0, top: 0 },
  statusCard: {
    backgroundColor: "rgba(255, 255, 255, 0.96)",
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    gap: 5,
    marginHorizontal: 16,
    marginTop: 8,
    padding: 14,
    ...tabi.shadow,
  },
  mapTitleRow: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  eyebrow: {
    color: tabi.color.accent,
    fontFamily: tabi.type.utility,
    fontSize: 9,
    fontWeight: "700",
    letterSpacing: 1.1,
  },
  mapTitle: {
    color: tabi.color.ink,
    fontSize: 23,
    fontWeight: "800",
    marginTop: 2,
  },
  livePill: {
    alignItems: "center",
    backgroundColor: "#E8F4EC",
    borderRadius: tabi.radius.pill,
    flexDirection: "row",
    gap: 6,
    paddingHorizontal: 10,
    paddingVertical: 7,
  },
  liveDot: {
    backgroundColor: tabi.color.success,
    borderRadius: 4,
    height: 8,
    width: 8,
  },
  liveText: { color: tabi.color.success, fontSize: 12, fontWeight: "700" },
  statusText: { color: tabi.color.mutedInk, fontSize: 12 },
  errorText: { color: tabi.color.danger, fontSize: 13, lineHeight: 18 },
  staleNotice: {
    alignItems: "center",
    backgroundColor: tabi.color.warningSoft,
    borderRadius: tabi.radius.small,
    flexDirection: "row",
    gap: 8,
    left: 16,
    padding: 10,
    position: "absolute",
    right: 16,
    top: 108,
  },
  staleText: { color: tabi.color.ink, flex: 1, fontSize: 13 },
  openPanelButton: {
    alignItems: "center",
    alignSelf: "center",
    backgroundColor: tabi.color.ink,
    borderRadius: tabi.radius.pill,
    bottom: 18,
    flexDirection: "row",
    gap: 8,
    minHeight: tabi.touchTarget,
    paddingHorizontal: 20,
    position: "absolute",
    ...tabi.shadow,
  },
  openPanelButtonPressed: {
    backgroundColor: "#30394A",
    transform: [{ scale: 0.98 }],
  },
  openPanelText: { color: tabi.color.white, fontSize: 15, fontWeight: "700" },
  sheet: {
    backgroundColor: tabi.color.canvas,
    borderTopLeftRadius: tabi.radius.large,
    borderTopRightRadius: tabi.radius.large,
    bottom: 0,
    elevation: 12,
    left: 0,
    maxHeight: "72%",
    paddingTop: 8,
    position: "absolute",
    right: 0,
    shadowColor: tabi.color.ink,
    shadowOffset: { width: 0, height: -4 },
    shadowOpacity: 0.16,
    shadowRadius: 12,
  },
  sheetHandle: {
    alignSelf: "center",
    backgroundColor: tabi.color.border,
    borderRadius: 2,
    height: 4,
    width: 38,
  },
  sheetHeader: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
    paddingHorizontal: 20,
    paddingTop: 12,
  },
  panelContent: { gap: 16, padding: 20, paddingBottom: 36, paddingTop: 14 },
  heading: {
    color: tabi.color.ink,
    fontSize: 24,
    fontWeight: "800",
    marginTop: 2,
  },
  closeButton: {
    alignItems: "center",
    backgroundColor: tabi.color.surfaceMuted,
    borderRadius: tabi.radius.pill,
    height: tabi.touchTarget,
    justifyContent: "center",
    width: tabi.touchTarget,
  },
  filters: { flexDirection: "row", flexWrap: "wrap", gap: 8 },
  filterChip: {
    alignItems: "center",
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.pill,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 38,
    paddingHorizontal: 12,
  },
  filterChipSelected: {
    backgroundColor: tabi.color.accentSoft,
    borderColor: tabi.color.accent,
  },
  filterChipText: {
    color: tabi.color.mutedInk,
    fontSize: 13,
    fontWeight: "600",
  },
  filterChipTextSelected: { color: tabi.color.accent, fontWeight: "700" },
  searchBox: {
    alignItems: "center",
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    flexDirection: "row",
    gap: 9,
    minHeight: tabi.touchTarget,
    paddingHorizontal: 13,
  },
  searchInput: {
    color: tabi.color.ink,
    flex: 1,
    fontSize: 16,
    paddingVertical: 9,
  },
  searchResults: { gap: 5 },
  miniHeading: { color: tabi.color.ink, fontSize: 14, fontWeight: "700" },
  searchResult: {
    alignItems: "center",
    backgroundColor: tabi.color.surface,
    borderRadius: tabi.radius.small,
    flexDirection: "row",
    gap: 9,
    minHeight: tabi.touchTarget,
    paddingHorizontal: 12,
  },
  searchResultText: {
    color: tabi.color.ink,
    flex: 1,
    fontSize: 14,
    fontWeight: "600",
  },
  detail: {
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    gap: 5,
    padding: 14,
  },
  detailTitleRow: {
    alignItems: "center",
    flexDirection: "row",
    gap: 11,
    marginBottom: 5,
  },
  vehicleIcon: {
    alignItems: "center",
    backgroundColor: tabi.color.bus,
    borderRadius: tabi.radius.small,
    height: 40,
    justifyContent: "center",
    width: 40,
  },
  detailTitleCopy: { flex: 1, gap: 2 },
  detailTitle: { color: tabi.color.ink, fontSize: 17, fontWeight: "800" },
  detailRoute: { color: tabi.color.mutedInk, fontSize: 13 },
  detailMeta: { color: tabi.color.mutedInk, fontSize: 12, lineHeight: 17 },
});
