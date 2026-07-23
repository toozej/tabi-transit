import { gzipSync } from "node:zlib";

const stopCount = 4_000;
const routeCount = 180;
const tripCount = 25_000;
const stopTimeCount = 180_000;

function timed(operation) {
  const started = process.hrtime.bigint();
  const result = operation();
  return {
    result,
    milliseconds: Number(process.hrtime.bigint() - started) / 1_000_000,
  };
}

const stops = Array.from({ length: stopCount }, (_, index) => ({
  id: `stop:${index}`,
  name: `Fixture stop ${index}`,
  latitudeE6: 45_500_000 + (index % 1_000),
  longitudeE6: -122_700_000 + (index % 1_000),
}));
const routes = Array.from({ length: routeCount }, (_, index) => ({
  id: `route:${index}`,
  shortName: String(index),
  mode: index % 6 === 0 ? "light_rail" : "bus",
}));
const trips = Array.from({ length: tripCount }, (_, index) => ({
  id: `trip:${index}`,
  routeId: `route:${index % routeCount}`,
  serviceId: `weekday:${index % 7}`,
  headsign: `Fixture destination ${index % 40}`,
}));
const stopTimes = Array.from({ length: stopTimeCount }, (_, index) => ({
  tripId: `trip:${index % tripCount}`,
  stopId: `stop:${index % stopCount}`,
  serviceDaySeconds: 18_000 + ((index * 97) % 90_000),
  sequence: Math.floor(index / tripCount),
}));

// JSON artifact models the current API-oriented static artifact. The normalized
// representation models SQLite table rows only; it is not a claim about a
// device SQLite file or native query speed.
const jsonArtifact = {
  version: "measurement-v1",
  stops,
  routes,
  trips,
  stopTimes,
};
const normalizedRows = { stops, routes, trips, stopTimes };
const json = timed(() => JSON.stringify(jsonArtifact));
const normalized = timed(() => JSON.stringify(normalizedRows));
const jsonParsed = timed(() => JSON.parse(json.result));
const normalizedParsed = timed(() => JSON.parse(normalized.result));
const lookup = timed(() => {
  const target = "stop:101";
  return normalizedParsed.result.stopTimes.filter(
    (row) => row.stopId === target,
  ).length;
});

function report(name, value) {
  const bytes = Buffer.byteLength(value);
  return { name, bytes, gzipBytes: gzipSync(value).length };
}

console.log(
  JSON.stringify(
    {
      dataset: { stopCount, routeCount, tripCount, stopTimeCount },
      artifacts: [
        report("json", json.result),
        report("normalized-sqlite-row-proxy", normalized.result),
      ],
      proxies: {
        jsonSerializeMs: json.milliseconds,
        jsonParseMs: jsonParsed.milliseconds,
        normalizedSerializeMs: normalized.milliseconds,
        normalizedParseMs: normalizedParsed.milliseconds,
        normalizedStopLookupMs: lookup.milliseconds,
        normalizedStopLookupRows: lookup.result,
      },
    },
    null,
    2,
  ),
);
