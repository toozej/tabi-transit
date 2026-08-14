import { Platform, StyleSheet } from "react-native";

/**
 * Tabi's shared product language. Layout, color, and hierarchy stay consistent
 * across platforms; typography, touch targets, corners, and elevation follow
 * the host OS so the app feels at home on both iOS and Android.
 */
export const tabi = {
  color: {
    canvas: "#F5F3EE",
    surface: "#FFFFFF",
    surfaceMuted: "#ECE9E2",
    ink: "#172033",
    mutedInk: "#667085",
    faintInk: "#8A929F",
    border: "#D8D5CE",
    accent: "#C84B40",
    accentPressed: "#A93B33",
    accentSoft: "#F8E8E5",
    bus: "#35658F",
    rail: "#4C7353",
    warning: "#9A6700",
    warningSoft: "#FFF4D6",
    danger: "#B42318",
    dangerSoft: "#FEECEB",
    success: "#287044",
    white: "#FFFFFF",
  },
  type: {
    body: Platform.select({ ios: "System", android: "sans-serif" }),
    display: Platform.select({ ios: "System", android: "sans-serif" }),
    utility: Platform.select({ ios: "Menlo", android: "monospace" }),
  },
  radius: {
    small: Platform.select({ ios: 10, android: 8, default: 9 }),
    medium: Platform.select({ ios: 16, android: 14, default: 15 }),
    large: Platform.select({ ios: 24, android: 20, default: 22 }),
    pill: 999,
  },
  touchTarget: Platform.select({ ios: 44, android: 48, default: 44 }),
  shadow: Platform.select({
    ios: {
      shadowColor: "#172033",
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.1,
      shadowRadius: 14,
    },
    android: { elevation: 3 },
    default: {},
  }),
} as const;

export const tabiCommonStyles = StyleSheet.create({
  screen: { backgroundColor: tabi.color.canvas, flex: 1 },
  page: {
    gap: 18,
    paddingHorizontal: 20,
    paddingBottom: 40,
    paddingTop: 12,
  },
  eyebrow: {
    color: tabi.color.accent,
    fontFamily: tabi.type.utility,
    fontSize: 11,
    fontWeight: "700",
    letterSpacing: 1.4,
  },
  title: {
    color: tabi.color.ink,
    fontFamily: tabi.type.display,
    fontSize: 34,
    fontWeight: "800",
    letterSpacing: -0.8,
    lineHeight: 39,
  },
  subtitle: {
    color: tabi.color.mutedInk,
    fontFamily: tabi.type.body,
    fontSize: 16,
    lineHeight: 23,
  },
  sectionTitle: {
    color: tabi.color.ink,
    fontFamily: tabi.type.display,
    fontSize: 21,
    fontWeight: "700",
    letterSpacing: -0.25,
  },
  body: {
    color: tabi.color.ink,
    fontFamily: tabi.type.body,
    fontSize: 15,
    lineHeight: 21,
  },
  secondary: {
    color: tabi.color.mutedInk,
    fontFamily: tabi.type.body,
    fontSize: 13,
    lineHeight: 18,
  },
  card: {
    backgroundColor: tabi.color.surface,
    borderColor: tabi.color.border,
    borderRadius: tabi.radius.medium,
    borderWidth: StyleSheet.hairlineWidth,
    padding: 16,
  },
  pressed: { opacity: 0.68 },
});
