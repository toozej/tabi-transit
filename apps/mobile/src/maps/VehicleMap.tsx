import Mapbox from "@rnmapbox/maps";
import { useMemo } from "react";
import { StyleSheet, View } from "react-native";

import type { Vehicle } from "@/domain/vehicleModels";

import { getMapboxAccessToken } from "./config";
import { toVehicleFeatureCollection } from "./vehicleGeoJson";

const MAP_STYLE = "mapbox://styles/mapbox/light-v11";

type Props = {
  vehicles: readonly Vehicle[];
  selectedVehicleId?: string;
};

export function VehicleMap({ vehicles, selectedVehicleId }: Props) {
  const accessToken = getMapboxAccessToken();
  const fleet = useMemo(() => toVehicleFeatureCollection(vehicles), [vehicles]);
  const selected = useMemo(
    () =>
      toVehicleFeatureCollection(
        vehicles.filter((item) => item.id === selectedVehicleId),
      ),
    [selectedVehicleId, vehicles],
  );

  if (accessToken === undefined) {
    return null;
  }

  Mapbox.setAccessToken(accessToken);

  return (
    <View
      style={styles.mapContainer}
      accessibilityLabel="Vehicle map; use the vehicle list below for accessible selection"
    >
      <Mapbox.MapView
        compassEnabled={false}
        scaleBarEnabled={false}
        style={styles.map}
        styleURL={MAP_STYLE}
      >
        <Mapbox.Camera centerCoordinate={[-122.6765, 45.5231]} zoomLevel={10} />
        <Mapbox.ShapeSource id="vehicles" shape={fleet}>
          <Mapbox.CircleLayer
            id="vehicle-stale"
            filter={["==", ["get", "freshness"], "stale"]}
            style={{
              circleColor: "#6B7280",
              circleRadius: 5,
              circleOpacity: 0.75,
            }}
          />
          <Mapbox.SymbolLayer
            id="vehicle-symbol"
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
        <Mapbox.ShapeSource id="selected-vehicle" shape={selected}>
          <Mapbox.CircleLayer
            id="selected-vehicle-halo"
            style={{
              circleColor: "#111827",
              circleRadius: 11,
              circleOpacity: 0.35,
            }}
          />
          <Mapbox.SymbolLayer
            id="selected-vehicle-symbol"
            style={{
              iconImage: "marker-15",
              iconColor: "#111827",
              iconAllowOverlap: true,
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
