import { create } from "zustand";

import {
  defaultPlannerConstraints,
  swapEndpoints,
  type PlanEndpoint,
  type PlannerConstraints,
  type PlannerDraft,
} from "@/domain/tripPlanning";

type PlannerState = {
  draft: PlannerDraft;
  locationPermission: "unknown" | "denied" | "granted";
  setOrigin: (endpoint?: PlanEndpoint) => void;
  setDestination: (endpoint?: PlanEndpoint) => void;
  setConstraints: (constraints: Partial<PlannerConstraints>) => void;
  swap: () => void;
  setLocationPermission: (value: PlannerState["locationPermission"]) => void;
};

export const usePlannerStore = create<PlannerState>((set) => ({
  draft: { constraints: defaultPlannerConstraints },
  locationPermission: "unknown",
  setOrigin: (origin) =>
    set((state) => ({ draft: { ...state.draft, origin } })),
  setDestination: (destination) =>
    set((state) => ({ draft: { ...state.draft, destination } })),
  setConstraints: (constraints) =>
    set((state) => ({
      draft: {
        ...state.draft,
        constraints: { ...state.draft.constraints, ...constraints },
      },
    })),
  swap: () => set((state) => ({ draft: swapEndpoints(state.draft) })),
  setLocationPermission: (locationPermission) => set({ locationPermission }),
}));
