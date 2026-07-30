import type { ItineraryLeg } from "./tripPlanning";

type Coordinate = readonly [longitude: number, latitude: number];

function toMapsCoordinate([longitude, latitude]: Coordinate): string {
  if (
    !Number.isFinite(longitude) ||
    !Number.isFinite(latitude) ||
    longitude < -180 ||
    longitude > 180 ||
    latitude < -90 ||
    latitude > 90
  ) {
    throw new Error("Walking directions require valid route coordinates.");
  }
  // Mapping URLs take latitude,longitude while itinerary geometry is GeoJSON
  // longitude,latitude.
  return `${latitude},${longitude}`;
}

/**
 * Builds a walking-only universal maps URL after the rider explicitly asks to
 * leave Tabi. Coordinates remain on device until that action.
 */
export function createExternalWalkingDirectionsLink(
  leg: Pick<ItineraryLeg, "mode" | "geometry">,
): string | undefined {
  if (leg.mode !== "walk" || !leg.geometry || leg.geometry.length < 2)
    return undefined;

  const originCoordinate = leg.geometry[0];
  const destinationCoordinate = leg.geometry[leg.geometry.length - 1];
  if (!originCoordinate || !destinationCoordinate) return undefined;
  const origin = toMapsCoordinate(originCoordinate);
  const destination = toMapsCoordinate(destinationCoordinate);
  const params = new URLSearchParams({
    api: "1",
    origin,
    destination,
    travelmode: "walking",
  });
  return `https://www.google.com/maps/dir/?${params.toString()}`;
}
