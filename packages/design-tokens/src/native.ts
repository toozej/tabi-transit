import { designTokens } from "./tokens.js";

/**
 * React Native-friendly token map. Values remain unitless where React Native
 * expects numbers; platform-specific font family choices live in the source.
 */
export const nativeTokens = {
  color: designTokens.color,
  typography: {
    fontFamily: designTokens.typography.fontFamily.native,
    monoFontFamily: designTokens.typography.fontFamily.monoNative,
    fontSize: designTokens.typography.fontSize,
    lineHeight: designTokens.typography.lineHeight,
    fontWeight: designTokens.typography.fontWeight,
  },
  space: designTokens.space,
  radius: designTokens.radius,
  elevation: designTokens.elevation,
  motion: designTokens.motion,
  icon: designTokens.icon,
  breakpoint: designTokens.breakpoint,
} as const;

export type NativeTokens = typeof nativeTokens;
