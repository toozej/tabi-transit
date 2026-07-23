import type { ExpoConfig, ConfigContext } from "expo/config";

const mapboxAccessToken = process.env.EXPO_PUBLIC_MAPBOX_ACCESS_TOKEN;
const apiMode = process.env.EXPO_PUBLIC_API_MODE ?? "fixture";
const apiBaseUrl = process.env.EXPO_PUBLIC_API_BASE_URL ?? "";

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: "Tabi",
  slug: "tabi",
  scheme: "tabi",
  version: "0.0.0",
  orientation: "portrait",
  userInterfaceStyle: "automatic",
  newArchEnabled: true,
  ios: {
    bundleIdentifier: "app.tabi.transit",
    supportsTablet: true,
  },
  android: {
    package: "app.tabi.transit",
  },
  experiments: {
    typedRoutes: true,
  },
  plugins: ["expo-router", "expo-sqlite", "@rnmapbox/maps"],
  extra: {
    mapboxAccessToken: mapboxAccessToken ?? "",
    mapboxConfigured: Boolean(mapboxAccessToken),
    apiMode,
    apiBaseUrl,
  },
});
