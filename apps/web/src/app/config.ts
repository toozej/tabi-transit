import { z } from "zod";

const blankToUndefined = (value: unknown) =>
  typeof value === "string" && value.trim() === "" ? undefined : value;

const optionalPublicToken = z.preprocess(
  blankToUndefined,
  z.string().trim().min(1).optional(),
);

const configSchema = z.object({
  mode: z.enum(["fixture", "remote"]).default("fixture"),
  apiBaseUrl: z.string().default("/v1"),
  // Vite intentionally exposes this value to the browser. It must therefore
  // only ever be a restricted public Maps SDK token, never a server token.
  mapboxAccessToken: optionalPublicToken,
  mapboxStyleUrl: z.preprocess(
    blankToUndefined,
    z
      .string()
      .url()
      .or(z.string().startsWith("mapbox://styles/"))
      .default("mapbox://styles/mapbox/light-v11"),
  ),
});

export type WebConfig = z.infer<typeof configSchema>;

export function readConfig(
  environment: Record<string, string | undefined>,
): WebConfig {
  const config = configSchema.parse({
    mode: environment.VITE_TABI_DATA_MODE,
    apiBaseUrl: environment.VITE_TABI_API_BASE_URL,
    mapboxAccessToken: environment.VITE_MAPBOX_ACCESS_TOKEN,
    mapboxStyleUrl: environment.VITE_MAPBOX_STYLE_URL,
  });
  if (config.mode === "remote" && !config.apiBaseUrl.startsWith("/")) {
    throw new Error("Remote API requests must use a same-origin path.");
  }
  return config;
}

export const webConfig = readConfig(import.meta.env);
