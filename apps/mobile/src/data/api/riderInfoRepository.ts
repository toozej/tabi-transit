import {
  applyNearbyLimit,
  alertSchema,
  nearbyStopsSchema,
  routeDetailSchema,
  routeShapeSchema,
  routeStopCollectionSchema,
  scheduleSchema,
  staticManifestSchema,
  stopDetailSchema,
  type Alert,
  type Arrival,
  type NearbyStops,
  type RouteDetail,
  type RouteShape,
  type RouteStopCollection,
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
  fixtureRouteStops,
  fixtureSchedule,
  fixtureStops,
} from "./riderInfoFixtures";

type Request = (input: string, init?: RequestInit) => Promise<Response>;
export type NearbyCoordinate = {
  latitude: number;
  longitude: number;
};

function requireFixtureRoute(id: string): void {
  if (fixtureRoute.route.id !== id)
    throw new ApiError("Route was not found.", "http");
}

function isValidNearbyCoordinate(
  coordinate: NearbyCoordinate | undefined,
): coordinate is NearbyCoordinate {
  return Boolean(
    coordinate &&
    Number.isFinite(coordinate.latitude) &&
    Number.isFinite(coordinate.longitude) &&
    coordinate.latitude >= -90 &&
    coordinate.latitude <= 90 &&
    coordinate.longitude >= -180 &&
    coordinate.longitude <= 180,
  );
}

export class RiderInfoRepository {
  constructor(
    private readonly runtime: ApiRuntimeConfig = getApiRuntimeConfig(),
    private readonly request: Request = fetch,
  ) {}
  async nearby(
    coordinate: NearbyCoordinate | undefined,
    limitPerMode?: number,
  ): Promise<NearbyStops> {
    if (this.runtime.apiMode === "fixture")
      return applyNearbyLimit(fixtureNearby, limitPerMode);
    if (!isValidNearbyCoordinate(coordinate))
      throw new ApiError(
        "A valid current location is required for nearby stops.",
        "http",
      );
    const query = new URLSearchParams({
      lat: String(coordinate.latitude),
      lon: String(coordinate.longitude),
    });
    if (limitPerMode) query.set("limitPerMode", String(limitPerMode));
    return this.get(`/v1/stops/nearby?${query.toString()}`, nearbyStopsSchema);
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
    if (this.runtime.apiMode === "fixture") {
      requireFixtureRoute(id);
      return fixtureRoute;
    }
    return this.get(`/v1/routes/${encodeURIComponent(id)}`, routeDetailSchema);
  }
  async routeShape(id: string): Promise<RouteShape> {
    if (this.runtime.apiMode === "fixture") {
      requireFixtureRoute(id);
      return fixtureRouteShape;
    }
    return this.get(
      `/v1/routes/${encodeURIComponent(id)}/shape`,
      routeShapeSchema,
    );
  }
  async routeStops(
    id: string,
    directionId?: 0 | 1,
  ): Promise<RouteStopCollection> {
    if (this.runtime.apiMode === "fixture") {
      const collection = fixtureRouteStops.find(
        (item) =>
          item.routeId === id &&
          (directionId === undefined || item.directionId === directionId),
      );
      if (!collection)
        throw new ApiError("Route stops were not found.", "http");
      return collection;
    }
    const query =
      directionId === undefined ? "" : `?directionId=${directionId}`;
    return this.get(
      `/v1/routes/${encodeURIComponent(id)}/stops${query}`,
      routeStopCollectionSchema,
    );
  }
  async schedule(stopId: string, serviceDate: string): Promise<Schedule> {
    if (this.runtime.apiMode === "fixture") {
      if (fixtureSchedule.stopId !== stopId)
        throw new ApiError("Schedule was not found.", "http");
      return {
        ...fixtureSchedule,
        serviceDate,
        schedule: fixtureSchedule.schedule.map((entry) => ({
          ...entry,
          serviceDate,
        })),
      };
    }
    return this.get(
      `/v1/stops/${encodeURIComponent(stopId)}/schedule?serviceDate=${encodeURIComponent(serviceDate)}`,
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
    let response: Response;
    try {
      response = await this.request(`${this.runtime.apiBaseUrl}${path}`);
    } catch (error) {
      if (error instanceof ApiError) throw error;
      throw new ApiError("The Tabi API could not be reached.", "offline");
    }
    if (!response.ok)
      throw new ApiError(
        "Rider information is temporarily unavailable.",
        response.status === 503 ? "source_unavailable" : "http",
      );
    return response.json();
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
