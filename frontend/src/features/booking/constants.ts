import { addDays, format } from "date-fns";

/** The booking steps, in the order a patient walks them. */
export const STEPS = ["doctor", "clinic", "date", "slot", "details"] as const;

export type Step = (typeof STEPS)[number];

/**
 * Which URL parameters stop being valid once a given step is (re)entered.
 * Going back to the date screen must clear the slot, or the confirmation
 * would quote a time the patient never picked.
 */
export const STALE_PARAMS_BY_STEP: Record<Step, string[]> = {
  doctor: ["doctor", "clinic", "date", "slot", "slotEnd"],
  clinic: ["clinic", "date", "slot", "slotEnd"],
  date: ["date", "slot", "slotEnd"],
  slot: ["slot", "slotEnd"],
  details: [],
};

/** How many days ahead the date picker offers. */
export const BOOKABLE_DAYS = 14;

export function upcomingDates(from = new Date()): string[] {
  return Array.from({ length: BOOKABLE_DAYS }, (_, index) =>
    format(addDays(from, index), "yyyy-MM-dd"),
  );
}
