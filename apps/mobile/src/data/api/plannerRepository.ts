import {
  itinerarySchema,
  rankItineraries,
  type Itinerary,
  type PlannerDraft,
  type PlanEndpoint,
} from "@/domain/tripPlanning";

import { getApiRuntimeConfig, type ApiRuntimeConfig } from "./config";
import { fixtureItineraries } from "./plannerFixtures";

export class PlannerFeatureDisabledError extends Error {
  constructor() {
    super(
      "Trip planning is unavailable until the provider, terms, and public API contract are approved.",
    );
    this.name = "PlannerFeatureDisabledError";
  }
}

export class PlannerResponseError extends Error {
  constructor(
    message: string,
    readonly kind: "offline" | "unavailable" | "http" | "invalid_response",
  ) {
    super(message);
    this.name = "PlannerResponseError";
  }
}

type Request = (input: string, init?: RequestInit) => Promise<Response>;

type RemotePlanResponse = {
  source: "trimet-web-services";
  freshness: { isRealtime: boolean };
  itineraries: Array<{
    id: string;
    durationSeconds: number;
    transfers: number;
    walkingMeters: number;
    legs: Array<{
      mode: string;
      routeId?: string;
      fromName: string;
      toName: string;
      startAt?: string;
      endAt?: string;
      distanceMeters?: number;
    }>;
  }>;
};

function remoteEndpoint(endpoint: PlanEndpoint) {
  if (endpoint.kind !== "stop") {
    throw new PlannerFeatureDisabledError();
  }
  return { type: "stop", stopId: endpoint.id, label: endpoint.label };
}

function remoteRequest(draft: PlannerDraft) {
  if (!draft.origin || !draft.destination)
    throw new PlannerFeatureDisabledError();
  return {
    origin: remoteEndpoint(draft.origin),
    destination: remoteEndpoint(draft.destination),
    time: { mode: draft.constraints.timeMode, value: draft.constraints.time },
    preferences: {
      modes: draft.constraints.modes,
      maxTransfers: draft.constraints.maxTransfers,
      maxWalkingMeters: draft.constraints.maxWalkingMeters,
      wheelchairAccessible: draft.constraints.wheelchairAccessible,
      optimize: draft.constraints.optimization,
    },
  };
}

function parseRemotePlan(value: unknown, draft: PlannerDraft): Itinerary[] {
  // This intentionally describes only the normalized public response that the
  // timeline needs. It is not a provider DTO and does not retain coordinates.
  const response = value as RemotePlanResponse;
  if (
    response?.source !== "trimet-web-services" ||
    !Array.isArray(response.itineraries) ||
    typeof response.freshness?.isRealtime !== "boolean"
  ) {
    throw new PlannerResponseError(
      "The planner returned an invalid response.",
      "invalid_response",
    );
  }
  return response.itineraries.map((item) => {
    if (
      typeof item.id !== "string" ||
      !Number.isInteger(item.durationSeconds) ||
      !Number.isInteger(item.transfers) ||
      !Number.isInteger(item.walkingMeters) ||
      !Array.isArray(item.legs) ||
      item.legs.length === 0
    ) {
      throw new PlannerResponseError(
        "The planner returned an invalid response.",
        "invalid_response",
      );
    }
    const legs = item.legs.map((leg, index) => {
      const mode = normalizedMode(leg.mode);
      if (
        mode === undefined ||
        typeof leg.fromName !== "string" ||
        typeof leg.toName !== "string"
      )
        throw new PlannerResponseError(
          "The planner returned an invalid response.",
          "invalid_response",
        );
      const startAt = leg.startAt ?? draft.constraints.time;
      const endAt = leg.endAt ?? startAt;
      return {
        id: `${item.id}:leg:${index + 1}`,
        mode,
        startLabel: leg.fromName || draft.origin?.label || "Origin",
        endLabel: leg.toName || draft.destination?.label || "Destination",
        startAt,
        endAt,
        durationSeconds:
          Math.max(0, Date.parse(endAt) - Date.parse(startAt)) / 1000,
        walkingMeters: mode === "walk" ? leg.distanceMeters : undefined,
        routeLabel: leg.routeId,
        realtime: response.freshness.isRealtime ? "estimated" : "scheduled",
      };
    });
    return itinerarySchema.parse({
      id: item.id,
      departureAt: legs[0]?.startAt,
      arrivalAt: legs.at(-1)?.endAt,
      durationSeconds: item.durationSeconds,
      transfers: item.transfers,
      walkingMeters: item.walkingMeters,
      source: response.source,
      freshness: {
        status: response.freshness.isRealtime ? "fresh" : "unknown",
        message: response.freshness.isRealtime
          ? "Live planning data from TriMet."
          : "Planning data may not include real-time updates.",
      },
      legs,
    });
  });
}

function normalizedMode(value: string) {
  switch (value) {
    case "walk":
    case "bus":
    case "light_rail":
    case "streetcar":
    case "commuter_rail":
      return value;
    // The backend's private TriMet adapter exposes these source names today;
    // map them at the mobile boundary, never in a screen component.
    case "rail":
      return "light_rail";
    case "tram":
      return "streetcar";
    default:
      return undefined;
  }
}

/**
 * Fixture mode is deterministic. Remote mode calls only Tabi's normalized API;
 * it never calls Mapbox or TriMet directly and only supports opaque stop IDs.
 */
export class PlannerRepository {
  constructor(
    private readonly runtime: ApiRuntimeConfig = getApiRuntimeConfig(),
    private readonly request: Request = fetch,
  ) {}

  async plan(draft: PlannerDraft): Promise<Itinerary[]> {
    if (!draft.origin || !draft.destination)
      throw new PlannerFeatureDisabledError();
    if (this.runtime.apiMode === "remote") return this.remotePlan(draft);
    const parsed = fixtureItineraries.map((value) =>
      itinerarySchema.parse(scheduleFixtureItinerary(value, draft)),
    );
    return rankItineraries(parsed, draft.constraints).itineraries;
  }

  private async remotePlan(draft: PlannerDraft): Promise<Itinerary[]> {
    // Validate before constructing a request so unsupported local-only
    // endpoints can never reach the network boundary.
    const body = JSON.stringify(remoteRequest(draft));
    let response: Response;
    try {
      response = await this.request(
        `${this.runtime.apiBaseUrl}/v1/journeys/plan`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body,
        },
      );
    } catch {
      throw new PlannerResponseError(
        "The Tabi API could not be reached.",
        "offline",
      );
    }
    if (!response.ok) {
      throw new PlannerResponseError(
        response.status === 503
          ? "Trip planning is unavailable."
          : "The Tabi API returned an error.",
        response.status === 503 ? "unavailable" : "http",
      );
    }
    try {
      return parseRemotePlan(await response.json(), draft);
    } catch (error) {
      if (error instanceof PlannerResponseError) throw error;
      throw new PlannerResponseError(
        "The planner returned an invalid response.",
        "invalid_response",
      );
    }
  }
}

function scheduleFixtureItinerary(
  itinerary: Itinerary,
  draft: PlannerDraft,
): Itinerary {
  const anchor =
    draft.constraints.timeMode === "depart_at"
      ? itinerary.departureAt
      : itinerary.arrivalAt;
  const shiftMilliseconds =
    Date.parse(draft.constraints.time) - Date.parse(anchor);
  const shift = (value: string) =>
    new Date(Date.parse(value) + shiftMilliseconds).toISOString();
  return {
    ...itinerary,
    departureAt: shift(itinerary.departureAt),
    arrivalAt: shift(itinerary.arrivalAt),
    legs: itinerary.legs.map((leg) => ({
      ...leg,
      startAt: shift(leg.startAt),
      endAt: shift(leg.endAt),
    })),
  };
}

export const plannerRepository = new PlannerRepository();
