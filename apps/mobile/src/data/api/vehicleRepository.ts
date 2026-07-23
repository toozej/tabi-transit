import {
  configSchema,
  vehicleCollectionSchema,
  vehicleDetailSchema,
  vehicleSearchSchema,
  type PublicConfig,
  type Vehicle,
  type VehicleCollection,
  type VehicleFilters,
  type VehicleSearch,
} from "@/domain/vehicleModels";

import { getApiRuntimeConfig, type ApiRuntimeConfig } from "./config";
import { fixtureCollection, fixtureConfig } from "./fixtures";

export class ApiError extends Error {
  constructor(
    message: string,
    readonly kind:
      | "offline"
      | "source_unavailable"
      | "http"
      | "invalid_response",
    readonly retryAfterSeconds?: number,
  ) {
    super(message);
  }
}

type CachedResponse<T> = { etag?: string; value: T };
type Request = (input: string, init?: RequestInit) => Promise<Response>;

export class VehicleRepository {
  private readonly cache = new Map<string, CachedResponse<unknown>>();

  constructor(
    private readonly runtime: ApiRuntimeConfig = getApiRuntimeConfig(),
    private readonly request: Request = fetch,
  ) {}

  async config(): Promise<PublicConfig> {
    if (this.runtime.apiMode === "fixture") return fixtureConfig;
    return this.get("/v1/config", configSchema);
  }

  async vehicles(
    filters: VehicleFilters = { modes: [] },
  ): Promise<VehicleCollection> {
    if (this.runtime.apiMode === "fixture") return fixtureCollection;
    const query = new URLSearchParams({ format: "json" });
    if (filters.modes.length) query.set("modes", filters.modes.join(","));
    if (filters.routeId) query.set("routes", filters.routeId);
    if (filters.freshness) query.set("freshness", filters.freshness);
    return this.get(
      `/v1/vehicles?${query.toString()}`,
      vehicleCollectionSchema,
    );
  }

  async search(query: string): Promise<VehicleSearch> {
    if (this.runtime.apiMode === "fixture") {
      const normalized = query.trim().toLowerCase();
      const vehicles = fixtureCollection.vehicles.filter((vehicle) =>
        [vehicle.id, vehicle.sourceVehicleId].some((candidate) =>
          candidate.toLowerCase().includes(normalized),
        ),
      );
      vehicles.sort(
        (a, b) =>
          Number(b.sourceVehicleId === query.trim()) -
          Number(a.sourceVehicleId === query.trim()),
      );
      return { query, vehicles, freshness: fixtureCollection.freshness };
    }
    return this.get(
      `/v1/vehicles/search?q=${encodeURIComponent(query)}`,
      vehicleSearchSchema,
    );
  }

  async vehicle(id: string): Promise<Vehicle> {
    if (this.runtime.apiMode === "fixture") {
      const vehicle = fixtureCollection.vehicles.find((item) => item.id === id);
      if (!vehicle) throw new ApiError("Vehicle was not found.", "http");
      return vehicle;
    }
    return (
      await this.get(
        `/v1/vehicles/${encodeURIComponent(id)}`,
        vehicleDetailSchema,
      )
    ).vehicle;
  }

  private async get<T>(
    path: string,
    schema: { parse: (value: unknown) => T },
  ): Promise<T> {
    const cached = this.cache.get(path) as CachedResponse<T> | undefined;
    let response: Response;
    try {
      response = await this.request(`${this.runtime.apiBaseUrl}${path}`, {
        headers: cached?.etag ? { "If-None-Match": cached.etag } : undefined,
      });
    } catch {
      throw new ApiError("The Tabi API could not be reached.", "offline");
    }
    if (response.status === 304 && cached) return cached.value;
    if (!response.ok) {
      const retry = Number(response.headers.get("Retry-After"));
      throw new ApiError(
        response.status === 503
          ? "Vehicle positions are temporarily unavailable."
          : "The Tabi API returned an error.",
        response.status === 503 ? "source_unavailable" : "http",
        Number.isFinite(retry) ? retry : undefined,
      );
    }
    let body: unknown;
    try {
      body = await response.json();
      const value = schema.parse(body);
      this.cache.set(path, {
        etag: response.headers.get("ETag") ?? undefined,
        value,
      });
      return value;
    } catch {
      throw new ApiError(
        "The Tabi API returned an invalid response.",
        "invalid_response",
      );
    }
  }
}

export const vehicleRepository = new VehicleRepository();
