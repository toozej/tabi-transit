import { Stack } from "expo-router";
import { useEffect } from "react";

import { AppProviders } from "@/bootstrap/AppProviders";
import { bootstrapSavedRepository } from "@/data/local/bootstrapSavedRepository";

export default function RootLayout() {
  useEffect(() => {
    void bootstrapSavedRepository();
  }, []);
  return (
    <AppProviders>
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Screen name="(tabs)" />
        <Stack.Screen name="stop/[stopId]" />
        <Stack.Screen name="route/[routeId]" />
        <Stack.Screen name="settings/notifications" />
      </Stack>
    </AppProviders>
  );
}
