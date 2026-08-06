import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";

export default defineConfig(({ mode }) => {
  // Read the unprefixed proxy target only in Vite's Node process. It is never
  // included in the browser bundle; browser requests remain relative to /v1.
  const environment = loadEnv(mode, process.cwd(), "");
  const apiProxy = environment.TABI_WEB_API_PROXY;
  return {
    plugins: [react()],
    server: {
      proxy: apiProxy
        ? { "/v1": { target: apiProxy, changeOrigin: true } }
        : undefined,
    },
    build: { sourcemap: true },
  };
});
