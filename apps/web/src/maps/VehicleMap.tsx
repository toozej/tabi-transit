import mapboxgl, { type Map as MapboxMap } from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { useEffect, useRef, useState } from "react";

import type { Vehicle } from "../features/models";
import { selectedVehicleFeature, vehicleFeatureCollection } from "./geojson";

const PORTLAND_CENTER: [number, number] = [-122.6765, 45.5231];
const VEHICLE_SOURCE = "vehicles";
const SELECTED_VEHICLE_SOURCE = "selected-vehicle";
const INTERACTIVE_LAYERS = ["vehicles-fresh", "vehicles-stale"] as const;

type Props = {
  accessToken: string;
  styleUrl: string;
  vehicles: readonly Vehicle[];
  selectedVehicleId?: string;
  onVehicleSelect: (id: string) => void;
};

function updateSources(
  map: MapboxMap,
  vehicles: readonly Vehicle[],
  selectedVehicleId?: string,
) {
  const fleet = map.getSource(VEHICLE_SOURCE) as
    | mapboxgl.GeoJSONSource
    | undefined;
  fleet?.setData(vehicleFeatureCollection(vehicles, selectedVehicleId));

  const selected = map.getSource(SELECTED_VEHICLE_SOURCE) as
    | mapboxgl.GeoJSONSource
    | undefined;
  selected?.setData(
    selectedVehicleFeature(
      vehicles.find((item) => item.id === selectedVehicleId),
    ) ?? {
      type: "FeatureCollection",
      features: [],
    },
  );
}

function addVehicleLayers(map: MapboxMap) {
  map.addSource(VEHICLE_SOURCE, {
    type: "geojson",
    data: { type: "FeatureCollection", features: [] },
  });
  map.addSource(SELECTED_VEHICLE_SOURCE, {
    type: "geojson",
    data: { type: "FeatureCollection", features: [] },
  });

  map.addLayer({
    id: "vehicles-stale",
    type: "circle",
    source: VEHICLE_SOURCE,
    filter: ["!=", ["get", "freshness"], "fresh"],
    paint: {
      "circle-color": "#6B7280",
      "circle-radius": 5,
      "circle-opacity": 0.75,
    },
  });
  map.addLayer({
    id: "vehicles-fresh",
    type: "circle",
    source: VEHICLE_SOURCE,
    filter: ["==", ["get", "freshness"], "fresh"],
    paint: {
      "circle-color": [
        "match",
        ["get", "mode"],
        "light_rail",
        "#DC2626",
        "streetcar",
        "#7C3AED",
        "commuter_rail",
        "#2563EB",
        "#15803D",
      ],
      "circle-radius": 7,
      "circle-stroke-color": "#FFFFFF",
      "circle-stroke-width": 1.5,
    },
  });
  map.addLayer({
    id: "selected-vehicle-halo",
    type: "circle",
    source: SELECTED_VEHICLE_SOURCE,
    paint: {
      "circle-color": "#111827",
      "circle-radius": 13,
      "circle-opacity": 0.35,
    },
  });
  map.addLayer({
    id: "selected-vehicle",
    type: "circle",
    source: SELECTED_VEHICLE_SOURCE,
    paint: {
      "circle-color": "#111827",
      "circle-radius": 7,
      "circle-stroke-color": "#FFFFFF",
      "circle-stroke-width": 2,
    },
  });
}

/**
 * Browser Mapbox adapter. The list and selected-detail controls remain the
 * accessible path; this canvas is a progressive visual enhancement.
 */
export function VehicleMap({
  accessToken,
  styleUrl,
  vehicles,
  selectedVehicleId,
  onVehicleSelect,
}: Props) {
  const container = useRef<HTMLDivElement>(null);
  const map = useRef<MapboxMap | undefined>(undefined);
  const selectHandler = useRef(onVehicleSelect);
  const [loadFailed, setLoadFailed] = useState(false);

  useEffect(() => {
    selectHandler.current = onVehicleSelect;
  }, [onVehicleSelect]);

  useEffect(() => {
    if (!container.current) return;

    mapboxgl.accessToken = accessToken;
    const instance = new mapboxgl.Map({
      container: container.current,
      style: styleUrl,
      center: PORTLAND_CENTER,
      zoom: 10,
      attributionControl: true,
    });
    map.current = instance;

    const selectFeature = (event: mapboxgl.MapLayerMouseEvent) => {
      const feature = event.features?.[0] as
        | { properties?: Record<string, unknown> }
        | undefined;
      const id = feature?.properties?.id;
      if (typeof id === "string") selectHandler.current(id);
    };
    const pointer = () => {
      instance.getCanvas().style.cursor = "pointer";
    };
    const resetPointer = () => {
      instance.getCanvas().style.cursor = "";
    };
    const onLoad = () => {
      addVehicleLayers(instance);
      updateSources(instance, vehicles, selectedVehicleId);
    };
    const onError = () => setLoadFailed(true);

    instance.on("load", onLoad);
    instance.on("error", onError);
    for (const layer of INTERACTIVE_LAYERS) {
      instance.on("click", layer, selectFeature);
      instance.on("mouseenter", layer, pointer);
      instance.on("mouseleave", layer, resetPointer);
    }

    return () => {
      map.current = undefined;
      instance.remove();
    };
  }, [accessToken, styleUrl]);

  useEffect(() => {
    const instance = map.current;
    if (!instance) return;
    if (instance.isStyleLoaded()) {
      updateSources(instance, vehicles, selectedVehicleId);
      return;
    }
    instance.once("load", () =>
      updateSources(instance, vehicles, selectedVehicleId),
    );
  }, [selectedVehicleId, vehicles]);

  return (
    <section className="vehicle-map" aria-label="Vehicle map">
      <div
        className="vehicle-map-canvas"
        ref={container}
        role="img"
        aria-label="Visual vehicle map. Use the vehicle list below to select a vehicle."
      />
      {loadFailed && (
        <p className="notice" role="alert">
          The vehicle map could not load. The accessible vehicle list remains
          available.
        </p>
      )}
    </section>
  );
}

export default VehicleMap;
