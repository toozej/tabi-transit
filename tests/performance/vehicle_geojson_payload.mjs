#!/usr/bin/env node
import { performance } from "node:perf_hooks";

const fleetSizes = [1000, 1500, 3000];

function vehicle(index) {
  return {
    id: `fixture:vehicle:${index}`,
    coordinate: [-122.7 + (index % 100) * 0.001, 45.5 + (index % 100) * 0.001],
    mode: index % 2 === 0 ? "bus" : "light_rail",
    freshness: index % 10 === 0 ? "aging" : "fresh",
  };
}

for (const size of fleetSizes) {
  const vehicles = Array.from({ length: size }, (_, index) => vehicle(index));
  const started = performance.now();
  const collection = {
    type: "FeatureCollection",
    features: vehicles.map((item) => ({
      type: "Feature",
      id: item.id,
      geometry: { type: "Point", coordinates: item.coordinate },
      properties: { id: item.id, mode: item.mode, freshness: item.freshness },
    })),
  };
  const payload = JSON.stringify(collection);
  const elapsedMs = performance.now() - started;
  if (
    collection.features.length !== size ||
    collection.features.some(
      (feature) => feature.geometry.coordinates.length !== 2,
    )
  ) {
    throw new Error(`invalid GeoJSON generated for ${size} vehicles`);
  }
  console.log(
    `${size} vehicles: ${Buffer.byteLength(payload)} bytes; ${elapsedMs.toFixed(2)} ms build/stringify`,
  );
}
