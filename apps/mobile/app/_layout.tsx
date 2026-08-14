import { Stack } from "expo-router";
import { useEffect } from "react";
import { Platform } from "react-native";
import { StatusBar } from "expo-status-bar";

import { AppProviders } from "@/bootstrap/AppProviders";
import { bootstrapSavedRepository } from "@/data/local/bootstrapSavedRepository";
import { tabi } from "@/ui/tabi";

export default function RootLayout() {
  useEffect(() => {
    void bootstrapSavedRepository();
  }, []);
  return (
    <AppProviders>
      <StatusBar style="dark" />
      <Stack
        screenOptions={{
          animation: Platform.OS === "ios" ? "default" : "fade_from_bottom",
          contentStyle: { backgroundColor: tabi.color.canvas },
          headerBackButtonDisplayMode: "minimal",
          headerShadowVisible: false,
          headerStyle: { backgroundColor: tabi.color.canvas },
          headerTintColor: tabi.color.accent,
          headerTitleStyle: {
            color: tabi.color.ink,
            fontFamily: tabi.type.body,
            fontWeight: "700",
          },
        }}
      >
        <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
        <Stack.Screen name="stop/[stopId]" options={{ title: "Stop" }} />
        <Stack.Screen name="route/[routeId]" options={{ title: "Route" }} />
        <Stack.Screen
          name="settings/notifications"
          options={{ title: "Notifications" }}
        />
      </Stack>
    </AppProviders>
  );
}
