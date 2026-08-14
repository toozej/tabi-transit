import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["packages/*/src/**/*.test.ts", "tests/unit/**/*.test.ts"],
    environment: "node",
    globals: false,
    passWithNoTests: false,
  },
});
