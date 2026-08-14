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
  icon: "./assets/images/tabi-app-icon.png",
  orientation: "portrait",
  userInterfaceStyle: "automatic",
  splash: {
    image: "./assets/images/tabi-app-icon.png",
    resizeMode: "contain",
    backgroundColor: "#F5F3EE",
  },
  newArchEnabled: true,
  ios: {
    bundleIdentifier: "app.tabi.transit",
    supportsTablet: true,
    icon: "./assets/images/tabi-app-icon.png",
  },
  android: {
    package: "app.tabi.transit",
    adaptiveIcon: {
      foregroundImage: "./assets/images/tabi-app-icon.png",
      backgroundColor: "#16191F",
    },
  },
  web: {
    favicon: "./assets/images/tabi-app-icon.png",
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
