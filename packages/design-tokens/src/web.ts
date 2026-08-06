import { designTokens, type DesignTokens } from "./tokens.js";

function declaration(name: string, value: string | number): string {
  return `  --tabi-${name}: ${value};`;
}

function pixel(value: number): string {
  return `${value}px`;
}

/** Create the CSS custom-property sheet from the platform-neutral token source. */
export function createWebCss(tokens: DesignTokens = designTokens): string {
  const declarations = [
    ...Object.entries(tokens.color).map(([name, value]) =>
      declaration(`color-${name}`, value),
    ),
    declaration("font-family-sans", tokens.typography.fontFamily.web),
    declaration("font-family-mono", tokens.typography.fontFamily.monoWeb),
    ...Object.entries(tokens.typography.fontSize).map(([name, value]) =>
      declaration(`font-size-${name}`, pixel(value)),
    ),
    ...Object.entries(tokens.typography.lineHeight).map(([name, value]) =>
      declaration(`line-height-${name}`, value),
    ),
    ...Object.entries(tokens.typography.fontWeight).map(([name, value]) =>
      declaration(`font-weight-${name}`, value),
    ),
    ...Object.entries(tokens.space).map(([name, value]) =>
      declaration(`space-${name}`, pixel(value)),
    ),
    ...Object.entries(tokens.radius).map(([name, value]) =>
      declaration(`radius-${name}`, pixel(value)),
    ),
    ...Object.entries(tokens.elevation).map(([name, value]) =>
      declaration(`elevation-${name}`, value),
    ),
    ...Object.entries(tokens.motion)
      .filter(([name]) => name !== "easingStandard")
      .map(([name, value]) => declaration(`motion-${name}`, `${value}ms`)),
    declaration("motion-easing-standard", tokens.motion.easingStandard),
    ...Object.entries(tokens.icon).map(([name, value]) =>
      declaration(`icon-${name}`, pixel(value)),
    ),
    ...Object.entries(tokens.breakpoint).map(([name, value]) =>
      declaration(`breakpoint-${name}`, pixel(value)),
    ),
  ];

  return `/* Generated from src/tokens.ts. Do not edit manually. */\n:root {\n${declarations.join("\n")}\n}\n`;
}

export const webCss = createWebCss();
