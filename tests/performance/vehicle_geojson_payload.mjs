#!/usr/bin/env node
/* global console */
import { Buffer } from "node:buffer";
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

function jsonPayload(vehicles) {
  return JSON.stringify({ vehicles });
}

function geoJsonPayload(vehicles) {
  return JSON.stringify({
    type: "FeatureCollection",
    features: vehicles.map((item) => ({
      type: "Feature",
      id: item.id,
      geometry: { type: "Point", coordinates: item.coordinate },
      properties: { id: item.id, mode: item.mode, freshness: item.freshness },
    })),
  });
}

function measure(build) {
  const started = performance.now();
  const payload = build();
  return { payload, elapsedMs: performance.now() - started };
}

for (const size of fleetSizes) {
  const vehicles = Array.from({ length: size }, (_, index) => vehicle(index));
  const json = measure(() => jsonPayload(vehicles));
  const geoJson = measure(() => geoJsonPayload(vehicles));

  const collection = JSON.parse(geoJson.payload);
  if (
    collection.features.length !== size ||
    collection.features.some(
      (feature) => feature.geometry.coordinates.length !== 2,
    )
  ) {
    throw new Error(`invalid GeoJSON generated for ${size} vehicles`);
  }
  console.log(
    `${size} vehicles: JSON ${Buffer.byteLength(json.payload)} bytes / ${json.elapsedMs.toFixed(2)} ms; GeoJSON ${Buffer.byteLength(geoJson.payload)} bytes / ${geoJson.elapsedMs.toFixed(2)} ms`,
  );
}
