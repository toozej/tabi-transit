/**
 * The sole value source for Tabi's visual tokens. Platform exports derive from
 * this object so browser and native clients retain the same scale and states.
 */
export const designTokens = {
  color: {
    canvas: "#fffaf5",
    surface: "#ffffff",
    surfaceRaised: "#fffdf9",
    text: "#17221d",
    textMuted: "#526158",
    border: "#bdc7bc",
    focus: "#075d8c",
    brand: "#a61b34",
    brandStrong: "#7d1025",
    brandOn: "#ffffff",
    transit: "#176b44",
    transitOn: "#ffffff",
    info: "#075d8c",
    infoOn: "#ffffff",
    success: "#176b44",
    successOn: "#ffffff",
    warning: "#8a4b00",
    warningOn: "#ffffff",
    danger: "#b42318",
    dangerOn: "#ffffff",
    disabled: "#d9dfd8",
    disabledText: "#66736a",
  },
  typography: {
    fontFamily: {
      web: 'Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif',
      native: "System",
      monoWeb:
        "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace",
      monoNative: "Menlo",
    },
    fontSize: {
      caption: 12,
      body: 16,
      label: 16,
      title: 20,
      heading: 28,
      display: 36,
    },
    lineHeight: {
      compact: 1.2,
      normal: 1.5,
      relaxed: 1.65,
    },
    fontWeight: {
      regular: 400,
      medium: 500,
      semibold: 600,
      bold: 700,
    },
  },
  space: {
    0: 0,
    1: 4,
    2: 8,
    3: 12,
    4: 16,
    5: 20,
    6: 24,
    8: 32,
    10: 40,
    12: 48,
  },
  radius: {
    none: 0,
    small: 4,
    medium: 8,
    large: 16,
    pill: 9999,
  },
  elevation: {
    flat: 0,
    raised: 1,
    floating: 2,
    overlay: 3,
  },
  motion: {
    instant: 0,
    fast: 120,
    normal: 200,
    slow: 320,
    easingStandard: "cubic-bezier(0.2, 0, 0, 1)",
  },
  icon: {
    small: 16,
    medium: 20,
    large: 24,
  },
  breakpoint: {
    compact: 320,
    mobile: 375,
    tablet: 768,
    desktop: 1024,
    wide: 1440,
  },
} as const;

export type DesignTokens = typeof designTokens;
