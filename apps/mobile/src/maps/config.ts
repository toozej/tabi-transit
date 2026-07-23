import Constants from "expo-constants";

type MapExtra = {
  mapboxAccessToken?: unknown;
};

export function getMapboxAccessToken(
  extra: MapExtra = Constants.expoConfig?.extra ?? {},
): string | undefined {
  const token = extra.mapboxAccessToken;
  return typeof token === "string" && token.trim().length > 0
    ? token
    : undefined;
}
