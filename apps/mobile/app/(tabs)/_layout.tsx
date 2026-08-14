import MaterialCommunityIcons from "@expo/vector-icons/MaterialCommunityIcons";
import { Tabs } from "expo-router";
import { Platform } from "react-native";

import { tabi } from "@/ui/tabi";

export default function TabLayout() {
  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: tabi.color.accent,
        tabBarInactiveTintColor: tabi.color.mutedInk,
        tabBarHideOnKeyboard: true,
        tabBarItemStyle: { paddingVertical: Platform.OS === "android" ? 4 : 0 },
        tabBarLabelStyle: {
          fontFamily: tabi.type.body,
          fontSize: Platform.OS === "ios" ? 10 : 11,
          fontWeight: "700",
        },
        tabBarStyle: {
          backgroundColor: tabi.color.surface,
          borderTopColor: tabi.color.border,
          height: Platform.OS === "android" ? 66 : undefined,
        },
      }}
    >
      <Tabs.Screen
        name="nearby"
        options={{
          title: "Nearby",
          tabBarIcon: ({ color, focused, size }) => (
            <MaterialCommunityIcons
              name={focused ? "map-marker-radius" : "map-marker-radius-outline"}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="map"
        options={{
          title: "Map",
          tabBarIcon: ({ color, focused, size }) => (
            <MaterialCommunityIcons
              name={focused ? "map" : "map-outline"}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="plan"
        options={{
          title: "Plan",
          tabBarIcon: ({ color, focused, size }) => (
            <MaterialCommunityIcons
              name={
                focused ? "transit-connection" : "transit-connection-variant"
              }
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="alerts"
        options={{
          title: "Alerts",
          tabBarIcon: ({ color, focused, size }) => (
            <MaterialCommunityIcons
              name={focused ? "bell" : "bell-outline"}
              color={color}
              size={size}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="saved"
        options={{
          title: "Saved",
          tabBarIcon: ({ color, focused, size }) => (
            <MaterialCommunityIcons
              name={focused ? "bookmark" : "bookmark-outline"}
              color={color}
              size={size}
            />
          ),
        }}
      />
    </Tabs>
  );
}
