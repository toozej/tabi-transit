import { describe, expect, it } from "vitest";

import { nativeTokens } from "./native.js";
import { designTokens } from "./tokens.js";
import { createWebCss } from "./web.js";

describe("design tokens", () => {
  it("derives browser custom properties from the shared token source", () => {
    const css = createWebCss();

    expect(css).toContain(`--tabi-color-brand: ${designTokens.color.brand};`);
    expect(css).toContain(
      `--tabi-font-size-body: ${designTokens.typography.fontSize.body}px;`,
    );
    expect(css).toContain(`--tabi-space-4: ${designTokens.space[4]}px;`);
    expect(css).toContain(
      `--tabi-breakpoint-desktop: ${designTokens.breakpoint.desktop}px;`,
    );
  });

  it("keeps the native map numeric and aligned with shared semantic values", () => {
    expect(nativeTokens.color.warning).toBe(designTokens.color.warning);
    expect(nativeTokens.space[4]).toBe(16);
    expect(nativeTokens.radius.pill).toBe(9999);
    expect(nativeTokens.typography.fontFamily).toBe(
      designTokens.typography.fontFamily.native,
    );
  });
});
