import { Stack } from "expo-router";

import { AppProviders } from "@/bootstrap/AppProviders";

export default function RootLayout() {
  return (
    <AppProviders>
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Screen name="(tabs)" />
        <Stack.Screen name="stop/[stopId]" />
        <Stack.Screen name="route/[routeId]" />
      </Stack>
    </AppProviders>
  );
}
