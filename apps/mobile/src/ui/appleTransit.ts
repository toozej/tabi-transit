import { tabi } from "./tabi";

/** @deprecated Use `tabi`. Kept as an alias while existing feature WIP moves. */
export const appleTransit = {
  color: {
    paper: tabi.color.canvas,
    card: tabi.color.surface,
    ink: tabi.color.ink,
    mutedInk: tabi.color.mutedInk,
    rule: tabi.color.border,
    line: tabi.color.accent,
    bus: tabi.color.bus,
    rail: tabi.color.rail,
  },
  type: tabi.type,
} as const;
