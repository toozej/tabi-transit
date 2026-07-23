import {
  applyNearbyLimit,
  alertSchema,
  nearbyStopsSchema,
  routeDetailSchema,
  routeShapeSchema,
  scheduleSchema,
  staticManifestSchema,
  stopDetailSchema,
  type Alert,
  type Arrival,
  type NearbyStops,
  type RouteDetail,
  type RouteShape,
  type Schedule,
  type StaticManifest,
  type Stop,
} from "@/domain/riderInfo";
import { freshnessSchema } from "@/domain/vehicleModels";
import { getApiRuntimeConfig, type ApiRuntimeConfig } from "./config";
import { ApiError } from "./vehicleRepository";
import {
  fixtureAlerts,
  fixtureArrivals,
  fixtureManifest,
  fixtureNearby,
  fixtureRoute,
  fixtureRouteShape,
  fixtureSchedule,
  fixtureStops,
} from "./riderInfoFixtures";

type Request = (input: string, init?: RequestInit) => Promise<Response>;
export class RiderInfoRepository {
  constructor(
    private readonly runtime: ApiRuntimeConfig = getApiRuntimeConfig(),
    private readonly request: Request = fetch,
  ) {}
  async nearby(limitPerMode?: number): Promise<NearbyStops> {
    if (this.runtime.apiMode === "fixture")
      return applyNearbyLimit(fixtureNearby, limitPerMode);
    return this.get(
      `/v1/stops/nearby?lat=45.52&lon=-122.68${limitPerMode ? `&limitPerMode=${limitPerMode}` : ""}`,
      nearbyStopsSchema,
    );
  }
  async stop(id: string): Promise<Stop> {
    if (this.runtime.apiMode === "fixture") {
      const stop = fixtureStops.find((item) => item.id === id);
      if (!stop) throw new ApiError("Stop was not found.", "http");
      return stop;
    }
    return (
      await this.get(`/v1/stops/${encodeURIComponent(id)}`, stopDetailSchema)
    ).stop;
  }
  async arrivals(stopId: string): Promise<Arrival[]> {
    if (this.runtime.apiMode === "fixture")
      return fixtureArrivals.filter((arrival) => arrival.stopId === stopId);
    const value = await this.json(
      `/v1/stops/${encodeURIComponent(stopId)}/arrivals`,
    );
    const parsed = zArrivalCollection.parse(value);
    return parsed.arrivals;
  }
  async route(id: string): Promise<RouteDetail> {
    if (this.runtime.apiMode === "fixture") return fixtureRoute;
    return this.get(`/v1/routes/${encodeURIComponent(id)}`, routeDetailSchema);
  }
  async routeShape(id: string): Promise<RouteShape> {
    if (this.runtime.apiMode === "fixture") return fixtureRouteShape;
    return this.get(
      `/v1/routes/${encodeURIComponent(id)}/shape`,
      routeShapeSchema,
    );
  }
  async schedule(stopId: string): Promise<Schedule> {
    if (this.runtime.apiMode === "fixture") return fixtureSchedule;
    return this.get(
      `/v1/stops/${encodeURIComponent(stopId)}/schedule`,
      scheduleSchema,
    );
  }
  async alerts(): Promise<Alert[]> {
    if (this.runtime.apiMode === "fixture") return fixtureAlerts;
    const value = await this.json("/v1/alerts");
    return zAlertCollection.parse(value).alerts;
  }
  async manifest(): Promise<StaticManifest> {
    if (this.runtime.apiMode === "fixture") return fixtureManifest;
    return this.get("/v1/static/manifest", staticManifestSchema);
  }
  private async json(path: string): Promise<unknown> {
    try {
      const response = await this.request(`${this.runtime.apiBaseUrl}${path}`);
      if (!response.ok)
        throw new ApiError(
          "Rider information is temporarily unavailable.",
          response.status === 503 ? "source_unavailable" : "http",
        );
      return await response.json();
    } catch (error) {
      if (error instanceof ApiError) throw error;
      throw new ApiError("The Tabi API could not be reached.", "offline");
    }
  }
  private async get<T>(
    path: string,
    schema: { parse: (value: unknown) => T },
  ): Promise<T> {
    try {
      return schema.parse(await this.json(path));
    } catch (error) {
      if (error instanceof ApiError) throw error;
      throw new ApiError(
        "The Tabi API returned an invalid response.",
        "invalid_response",
      );
    }
  }
}
// Local schemas avoid trusting a partially deployed server response.
import { z } from "zod";
import { arrivalSchema } from "@/domain/riderInfo";
const zArrivalCollection = z.object({
  arrivals: z.array(arrivalSchema),
  freshness: freshnessSchema,
});
const zAlertCollection = z.object({
  alerts: z.array(alertSchema),
  freshness: freshnessSchema,
});
export const riderInfoRepository = new RiderInfoRepository();
