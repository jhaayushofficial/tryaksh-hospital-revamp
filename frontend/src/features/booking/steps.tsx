import { format, isToday, isTomorrow, parseISO } from "date-fns";
import { Check, MapPin, Phone } from "lucide-react";
import type { Clinic, Slot } from "../../lib/types";
import { CLINIC } from "../../lib/clinic";
import { formatDateLong, formatTime, partOfDay, type PartOfDay } from "../../lib/format";
import { cx } from "../../lib/cx";
import { Button, Card, EmptyState, Spinner } from "../../components/ui";

export function ClinicStep({
  clinics,
  onSelect,
}: {
  clinics: Clinic[];
  onSelect: (clinicId: string) => void;
}) {
  if (clinics.length === 0) {
    return (
      <EmptyState
        title="No locations are listed"
        hint={`Call ${CLINIC.phoneDisplay} to book by phone.`}
      />
    );
  }

  return (
    <ul className="space-y-3">
      {clinics.map((clinic) => (
        <Card as="li" key={clinic.id} className="p-4">
          <h2 className="text-base font-medium text-navy">{clinic.name}</h2>
          {clinic.address && (
            <p className="mt-1.5 flex items-start gap-2 text-sm text-ink-soft">
              <MapPin size={15} className="mt-0.5 shrink-0 text-ink-mute" aria-hidden />
              {clinic.address}
            </p>
          )}
          {clinic.phone && (
            <p className="mt-1 flex items-center gap-2 text-sm text-ink-soft">
              <Phone size={15} className="shrink-0 text-ink-mute" aria-hidden />
              {clinic.phone}
            </p>
          )}
          <Button variant="secondary" className="mt-4" onClick={() => onSelect(clinic.id)}>
            Choose this location
          </Button>
        </Card>
      ))}
    </ul>
  );
}

export function DateStep({
  dates,
  selected,
  onSelect,
}: {
  dates: string[];
  selected?: string | null;
  onSelect: (date: string) => void;
}) {
  return (
    <ul className="grid grid-cols-3 gap-2 sm:grid-cols-4">
      {dates.map((date) => {
        const day = parseISO(date);
        const isChosen = selected === date;

        // "Today" and "Tomorrow" are what people say; the rest get a weekday.
        const label = isToday(day) ? "Today" : isTomorrow(day) ? "Tomorrow" : format(day, "EEE");

        return (
          <li key={date}>
            <button
              onClick={() => onSelect(date)}
              aria-pressed={isChosen}
              aria-label={formatDateLong(date)}
              className={cx(
                "w-full rounded-lg border px-2 py-3 text-center transition-colors",
                isChosen
                  ? "border-red bg-red-tint text-red-deep"
                  : "border-line bg-white text-navy hover:border-navy",
              )}
            >
              <span className="block text-xs text-ink-mute">{label}</span>
              <span className="tabular mt-0.5 block text-base font-medium">
                {format(day, "d MMM")}
              </span>
            </button>
          </li>
        );
      })}
    </ul>
  );
}

const PART_ORDER: PartOfDay[] = ["Morning", "Afternoon", "Evening"];

export function SlotStep({
  slots,
  loading,
  selected,
  onSelect,
}: {
  slots: Slot[];
  loading: boolean;
  selected?: string | null;
  onSelect: (slot: Slot) => void;
}) {
  if (loading) return <Spinner label="Checking availability" />;

  const openSlots = slots.filter((slot) => slot.status !== "PAST");

  if (openSlots.length === 0) {
    return (
      <EmptyState
        title="Nothing free on this day"
        hint="Go back and try another date, or call the hospital — walk-ins are seen too."
      />
    );
  }

  const byPart = new Map<PartOfDay, Slot[]>();
  for (const slot of openSlots) {
    const part = partOfDay(slot.start);
    byPart.set(part, [...(byPart.get(part) ?? []), slot]);
  }

  return (
    <div className="space-y-6">
      {PART_ORDER.filter((part) => byPart.has(part)).map((part) => (
        <section key={part}>
          <h2 className="mb-2.5 text-xs font-semibold tracking-[0.14em] text-ink-mute uppercase">
            {part}
          </h2>
          <ul className="grid grid-cols-3 gap-2 sm:grid-cols-4">
            {byPart.get(part)!.map((slot) => {
              const isBooked = slot.status !== "AVAILABLE";
              const isChosen = selected === slot.start;

              return (
                <li key={slot.start}>
                  <button
                    disabled={isBooked}
                    onClick={() => onSelect(slot)}
                    aria-pressed={isChosen}
                    aria-label={`${formatTime(slot.start)}${isBooked ? ", already booked" : ""}`}
                    className={cx(
                      "tabular w-full rounded-lg border py-2.5 text-sm transition-colors",
                      isBooked && "cursor-not-allowed border-line bg-paper-sunk text-ink-mute line-through",
                      !isBooked && isChosen && "border-red bg-red-tint text-red-deep",
                      !isBooked && !isChosen && "border-line bg-white text-navy hover:border-navy",
                    )}
                  >
                    {slot.start}
                  </button>
                </li>
              );
            })}
          </ul>
        </section>
      ))}
      <p className="text-xs text-ink-mute">
        Times are {CLINIC.city} local time. Crossed-out times are already taken.
      </p>
    </div>
  );
}

export function ConfirmedStep({
  reference,
  doctorName,
  clinicName,
  date,
  time,
  onDone,
}: {
  reference: string;
  doctorName?: string;
  clinicName?: string;
  date: string;
  time: string;
  onDone: () => void;
}) {
  return (
    <main className="mx-auto max-w-lg px-5 py-14 text-center">
      <span
        className="mx-auto mb-6 flex h-14 w-14 items-center justify-center rounded-full bg-jade-tint text-jade"
        aria-hidden
      >
        <Check size={26} />
      </span>

      <h1 className="text-2xl">Appointment confirmed</h1>
      <p className="mt-2 text-sm text-ink-soft">
        {doctorName} · {formatDateLong(date)} · {formatTime(time)}
      </p>
      {clinicName && <p className="text-sm text-ink-mute">{clinicName}</p>}

      {/*
        The reference is the patient's only proof of the booking, so it is set
        in the mono face at a size that survives being read aloud on the phone
        or shown at the front desk.
      */}
      <div className="mx-auto mt-8 w-fit rounded-card border border-dashed border-red/40 bg-red-tint px-7 py-4">
        <p className="text-[11px] font-semibold tracking-[0.14em] text-red-deep uppercase">
          Your reference
        </p>
        <p className="tabular mt-1 text-2xl font-medium text-red-deep">{reference}</p>
      </div>

      <p className="mx-auto mt-6 max-w-sm text-xs leading-relaxed text-ink-mute">
        Show this reference at the front desk. Arrive ten minutes early; the fee is payable at
        the clinic. To change or cancel, call{" "}
        <a href={CLINIC.phoneHref} className="font-medium text-red-deep underline underline-offset-2">
          {CLINIC.phoneDisplay}
        </a>
        .
      </p>

      <Button variant="ghost" className="mt-8" onClick={onDone}>
        Back to home
      </Button>
    </main>
  );
}
