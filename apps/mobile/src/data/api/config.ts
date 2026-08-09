import Constants from "expo-constants";
import { z } from "zod";

const apiConfigSchema = z.object({
  apiMode: z.enum(["fixture", "remote"]).default("fixture"),
  apiBaseUrl: z.preprocess(
    (value) => (value === "" ? undefined : value),
    z.string().url().optional(),
  ),
});

export type ApiRuntimeConfig = z.infer<typeof apiConfigSchema>;

export function getApiRuntimeConfig(
  extra: Record<string, unknown> = Constants.expoConfig?.extra ?? {},
): ApiRuntimeConfig {
  const parsed = apiConfigSchema.parse(extra);
  if (parsed.apiMode === "remote" && !parsed.apiBaseUrl) {
    throw new Error(
      "A Tabi API base URL is required when remote mode is enabled.",
    );
  }
  return parsed;
}
