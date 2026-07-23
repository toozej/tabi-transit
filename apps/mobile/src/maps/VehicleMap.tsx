import Mapbox from "@rnmapbox/maps";
import { useMemo } from "react";
import { StyleSheet, View } from "react-native";

import { createSyntheticFleet } from "@/domain/vehicles";

import { getMapboxAccessToken } from "./config";
import { toVehicleFeatureCollection } from "./vehicleGeoJson";

const MAP_STYLE = "mapbox://styles/mapbox/light-v11";

export function VehicleMap() {
  const accessToken = getMapboxAccessToken();
  const fleet = useMemo(
    () => toVehicleFeatureCollection(createSyntheticFleet()),
    [],
  );

  if (accessToken === undefined) {
    return null;
  }

  Mapbox.setAccessToken(accessToken);

  return (
    <View
      style={styles.mapContainer}
      accessibilityLabel="Synthetic vehicle map"
    >
      <Mapbox.MapView style={styles.map} styleURL={MAP_STYLE}>
        <Mapbox.Camera centerCoordinate={[-122.6765, 45.5231]} zoomLevel={10} />
        <Mapbox.ShapeSource id="synthetic-vehicles" shape={fleet}>
          <Mapbox.CircleLayer
            id="synthetic-vehicle-stale"
            filter={["==", ["get", "freshness"], "stale"]}
            style={{
              circleColor: "#6B7280",
              circleRadius: 5,
              circleOpacity: 0.75,
            }}
          />
          <Mapbox.SymbolLayer
            id="synthetic-vehicle-symbol"
            filter={["==", ["get", "freshness"], "fresh"]}
            style={{
              iconImage: "marker-15",
              iconRotate: ["get", "bearing"],
              iconAllowOverlap: true,
              iconColor: [
                "match",
                ["get", "mode"],
                "max",
                "#DC2626",
                "streetcar",
                "#7C3AED",
                "wes",
                "#2563EB",
                "#15803D",
              ],
            }}
          />
        </Mapbox.ShapeSource>
      </Mapbox.MapView>
    </View>
  );
}

const styles = StyleSheet.create({
  mapContainer: { flex: 1 },
  map: { flex: 1 },
});
