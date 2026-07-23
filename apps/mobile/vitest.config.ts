import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

const appRoot = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  resolve: {
    alias: {
      "@": path.join(appRoot, "src"),
      "expo-constants": path.join(appRoot, "src/test/expo-constants.ts"),
      "react-native": path.join(appRoot, "src/test/react-native.ts"),
    },
  },
  test: {
    environment: "node",
    // The RNTL/Vitest bridge is kept as an explicit device/harness gate; the
    // current dependency pair cannot parse React Native's Flow entrypoint.
    include: ["src/**/*.test.ts"],
  },
});
