import {
  itinerarySchema,
  rankItineraries,
  type Itinerary,
  type PlannerDraft,
} from "@/domain/tripPlanning";

import { fixtureItineraries } from "./plannerFixtures";

export class PlannerFeatureDisabledError extends Error {
  constructor() {
    super(
      "Trip planning is unavailable until the provider, terms, and public API contract are approved.",
    );
    this.name = "PlannerFeatureDisabledError";
  }
}

/**
 * A deliberately fixture-only boundary. It keeps mobile independent of Mapbox
 * and TriMet until D-001/D-004 and the backend planner contract are complete.
 */
export class PlannerRepository {
  async plan(draft: PlannerDraft): Promise<Itinerary[]> {
    if (!draft.origin || !draft.destination)
      throw new PlannerFeatureDisabledError();
    const parsed = fixtureItineraries.map((value) =>
      itinerarySchema.parse(value),
    );
    return rankItineraries(parsed, draft.constraints).itineraries;
  }
}

export const plannerRepository = new PlannerRepository();
