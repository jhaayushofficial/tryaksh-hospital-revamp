import { format, parseISO } from "date-fns";

/** Formatting helpers shared by the booking flow and the admin panel. */

/** "2026-09-07" → "Mon, 7 Sep". */
export function formatDateShort(isoDate: string): string {
  return format(parseISO(isoDate), "EEE, d MMM");
}

/** "2026-09-07" → "Monday, 7 September 2026". */
export function formatDateLong(isoDate: string): string {
  return format(parseISO(isoDate), "EEEE, d MMMM yyyy");
}

/** Accepts "14:30", "14:30:00" or a timestamp; returns "2:30 pm". */
export function formatTime(value: string): string {
  const clock = value.includes("T") ? value.split("T")[1] : value;
  const [hours, minutes] = clock.split(":");
  const h = Number(hours);
  if (Number.isNaN(h)) return value;

  const suffix = h < 12 ? "am" : "pm";
  const display = h % 12 === 0 ? 12 : h % 12;
  return `${display}:${minutes} ${suffix}`;
}

/** Strips a timestamp down to its date, for API values that carry both. */
export function toDateOnly(value: string): string {
  return value.includes("T") ? value.split("T")[0] : value;
}

/**
 * Groups slots into the parts of the day patients actually think in, so a
 * long list of times reads as "morning / afternoon / evening" instead of one
 * undifferentiated grid.
 */
export type PartOfDay = "Morning" | "Afternoon" | "Evening";

export function partOfDay(time: string): PartOfDay {
  const hour = Number(time.split(":")[0]);
  if (hour < 12) return "Morning";
  if (hour < 17) return "Afternoon";
  return "Evening";
}
