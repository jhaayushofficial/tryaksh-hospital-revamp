import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import * as api from "../../lib/api";
import { report } from "../../lib/logger";
import type { Clinic, Doctor, Slot } from "../../lib/types";
import { formatDateShort } from "../../lib/format";
import { Alert } from "../../components/ui";
import { HomePage } from "../home/HomePage";
import { BookingShell } from "./BookingShell";
import { DetailsStep } from "./DetailsStep";
import { ClinicStep, ConfirmedStep, DateStep, SlotStep } from "./steps";
import { STALE_PARAMS_BY_STEP, upcomingDates, type Step } from "./constants";

/**
 * Booking state lives in the URL.
 *
 * That is what makes the browser's back button, a refresh and a shared link
 * all behave the way a patient expects, and it is why there is no reducer here.
 */
type WizardStep = Step | "home" | "confirmed";

/** Slots belong to one doctor/clinic/date triple; the key is how we know
 *  whether what we hold is still the answer to the question being asked. */
function slotKeyFor(doctorId: string | null, clinicId: string | null, date: string | null) {
  return `${doctorId}|${clinicId}|${date}`;
}

export function BookingWizard() {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  const [doctors, setDoctors] = useState<Doctor[]>([]);
  const [clinics, setClinics] = useState<Clinic[]>([]);
  const [loadedSlots, setLoadedSlots] = useState<{ key: string; slots: Slot[] } | null>(null);
  const [loadingDirectory, setLoadingDirectory] = useState(true);
  const [error, setError] = useState("");

  const step = (searchParams.get("step") ?? "home") as WizardStep;
  const doctorId = searchParams.get("doctor");
  const clinicId = searchParams.get("clinic");
  const date = searchParams.get("date");
  const slotStart = searchParams.get("slot");
  const slotEnd = searchParams.get("slotEnd");

  useEffect(() => {
    const controller = new AbortController();

    Promise.all([api.fetchDoctors(controller.signal), api.fetchClinics(controller.signal)])
      .then(([fetchedDoctors, fetchedClinics]) => {
        setDoctors(fetchedDoctors ?? []);
        setClinics(fetchedClinics ?? []);
      })
      .catch((cause) => {
        if (controller.signal.aborted) return;
        setError(report("failed to load doctors and locations", cause));
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoadingDirectory(false);
      });

    return () => controller.abort();
  }, []);

  const slotKey = slotKeyFor(doctorId, clinicId, date);

  useEffect(() => {
    if (step !== "slot" || !doctorId || !clinicId || !date) return;

    const controller = new AbortController();
    const key = slotKeyFor(doctorId, clinicId, date);

    api
      .fetchSlots(doctorId, clinicId, date, controller.signal)
      .then((data) => setLoadedSlots({ key, slots: data.slots ?? [] }))
      .catch((cause) => {
        if (controller.signal.aborted) return;
        setError(report("failed to load slots", cause, { doctorId, clinicId, date }));
        // Record the failure against this key so the screen stops spinning.
        setLoadedSlots({ key, slots: [] });
      });

    return () => controller.abort();
  }, [step, doctorId, clinicId, date]);

  // Anything held for a different doctor, clinic or day is stale, so the slot
  // screen is loading until the answer for the current one arrives.
  const slotsReady = loadedSlots?.key === slotKey;
  const slots = slotsReady ? loadedSlots.slots : [];

  /**
   * Moves to a step, dropping anything chosen after it. Going back to the date
   * screen must not leave the previously selected slot in the URL, or the
   * confirmation would quote a time the patient never picked.
   */
  const goTo = useCallback(
    (target: WizardStep, updates: Record<string, string> = {}) => {
      setError("");
      setSearchParams((current) => {
        const next = new URLSearchParams(current);
        next.set("step", target);

        for (const stale of STALE_PARAMS_BY_STEP[target as Step] ?? []) {
          next.delete(stale);
        }
        for (const [key, value] of Object.entries(updates)) {
          next.set(key, value);
        }
        return next;
      });
      window.scrollTo({ top: 0, behavior: "smooth" });
    },
    [setSearchParams],
  );

  const doctor = useMemo(() => doctors.find((d) => d.id === doctorId), [doctors, doctorId]);
  const clinic = useMemo(() => clinics.find((c) => c.id === clinicId), [clinics, clinicId]);
  const dates = useMemo(() => upcomingDates(), []);

  const banner = error ? <Alert>{error}</Alert> : null;

  if (step === "home") {
    return (
      <>
        {error && <div className="mx-auto max-w-6xl px-5 pt-5">{banner}</div>}
        <HomePage
          doctors={doctors}
          loading={loadingDirectory}
          onPickDoctor={(id) => goTo("clinic", { doctor: id })}
        />
      </>
    );
  }

  if (step === "confirmed" && date && slotStart) {
    return (
      <ConfirmedStep
        reference={searchParams.get("ref") ?? "—"}
        doctorName={doctor?.name}
        clinicName={clinic?.name}
        date={date}
        time={slotStart}
        onDone={() => navigate("/")}
      />
    );
  }

  if (step === "clinic" && doctor) {
    return (
      <BookingShell
        step="clinic"
        title="Where would you like to be seen?"
        onBack={() => navigate("/")}
        doctor={doctor}
      >
        {banner}
        <ClinicStep clinics={clinics} onSelect={(id) => goTo("date", { clinic: id })} />
      </BookingShell>
    );
  }

  if (step === "date" && doctor && clinic) {
    return (
      <BookingShell
        step="date"
        title="Pick a day"
        onBack={() => goTo("clinic")}
        doctor={doctor}
        clinic={clinic}
      >
        {banner}
        <DateStep dates={dates} selected={date} onSelect={(value) => goTo("slot", { date: value })} />
      </BookingShell>
    );
  }

  if (step === "slot" && doctor && clinic && date) {
    return (
      <BookingShell
        step="slot"
        title={`Free times on ${formatDateShort(date)}`}
        onBack={() => goTo("date")}
        doctor={doctor}
        clinic={clinic}
      >
        {banner}
        <SlotStep
          slots={slots}
          loading={!slotsReady}
          selected={slotStart}
          onSelect={(slot) => goTo("details", { slot: slot.start, slotEnd: slot.end })}
        />
      </BookingShell>
    );
  }

  if (step === "details" && doctor && clinic && date && slotStart && slotEnd) {
    return (
      <BookingShell
        step="details"
        title="Your details"
        onBack={() => goTo("slot")}
        doctor={doctor}
        clinic={clinic}
        date={date}
      >
        <DetailsStep
          doctor={doctor}
          clinicId={clinic.id}
          date={date}
          slotStart={slotStart}
          slotEnd={slotEnd}
          onBooked={(reference) => goTo("confirmed", { ref: reference })}
        />
      </BookingShell>
    );
  }

  // The URL names a step whose earlier choices are missing — a hand-edited
  // link, or a doctor who has since been deactivated. Start over rather than
  // render an empty screen.
  return (
    <BookingShell step="doctor" title="Let's start again" onBack={() => navigate("/")}>
      {banner}
      <Alert>
        {loadingDirectory
          ? "Loading your booking…"
          : "That booking link is incomplete. Go back and choose a doctor to begin."}
      </Alert>
    </BookingShell>
  );
}
