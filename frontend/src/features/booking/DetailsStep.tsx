import { useEffect, useRef, useState } from "react";
import {
  RecaptchaVerifier,
  signInWithPhoneNumber,
  type ConfirmationResult,
} from "firebase/auth";
import { getFirebaseAuth } from "../../lib/firebase";
import * as api from "../../lib/api";
import { logger, report } from "../../lib/logger";
import type { Doctor } from "../../lib/types";
import { Alert, Button, Field } from "../../components/ui";

const PHONE_DIGITS = 10;
const CODE_DIGITS = 6;

interface DetailsStepProps {
  doctor: Doctor;
  clinicId: string;
  date: string;
  slotStart: string;
  slotEnd: string;
  onBooked: (reference: string) => void;
}

/**
 * Name, number, code, booked.
 *
 * The number is verified through Firebase phone auth; the resulting ID token
 * is exchanged for a Tryaksh session token, which is what actually authorises
 * the booking. The patient never sees that two systems are involved.
 */
export function DetailsStep({
  doctor,
  clinicId,
  date,
  slotStart,
  slotEnd,
  onBooked,
}: DetailsStepProps) {
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [awaitingCode, setAwaitingCode] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const confirmationRef = useRef<ConfirmationResult | null>(null);
  const verifierRef = useRef<RecaptchaVerifier | null>(null);

  // The verifier attaches itself to a DOM node, so it is created after mount
  // and torn down with the step rather than left on `window`. Firebase
  // initialization can fail (bad config); that surfaces as the same on-screen
  // error a failed send would, rather than crashing the step.
  useEffect(() => {
    try {
      verifierRef.current = new RecaptchaVerifier(getFirebaseAuth(), "recaptcha-container", {
        size: "invisible",
      });
    } catch (cause) {
      setError(report("failed to set up phone verification", cause, { doctorId: doctor.id }));
    }

    return () => {
      verifierRef.current?.clear();
      verifierRef.current = null;
    };
  }, [doctor.id]);

  const sendCode = async () => {
    setError("");
    setBusy(true);
    try {
      const verifier = verifierRef.current;
      if (!verifier) throw new Error("Verification isn't ready yet. Reload the page.");

      const e164 = phone.startsWith("+") ? phone : `+91${phone}`;
      confirmationRef.current = await signInWithPhoneNumber(getFirebaseAuth(), e164, verifier);

      logger.info("verification code sent", { doctorId: doctor.id, date });
      setAwaitingCode(true);
    } catch (cause) {
      setError(report("failed to send verification code", cause, { date }));
    } finally {
      setBusy(false);
    }
  };

  const confirmAndBook = async () => {
    setError("");
    setBusy(true);
    try {
      const confirmation = confirmationRef.current;
      if (!confirmation) throw new Error("That code has expired. Request a new one.");

      const credential = await confirmation.confirm(code);
      const idToken = await credential.user.getIdToken();
      const { token } = await api.firebaseLogin(idToken);

      const appointment = await api.bookAppointment(
        {
          doctor_id: doctor.id,
          clinic_id: clinicId,
          appointment_date: date,
          start_time: slotStart,
          end_time: slotEnd,
          patient_name: name.trim(),
          patient_phone: phone,
          // Keyed to the slot, so a double tap or a retry cannot produce two
          // appointments for the same person at the same time.
          idempotency_key: `${phone}-${date}-${slotStart}`,
        },
        token,
      );

      logger.info("appointment booked", { reference: appointment.reference, date });
      onBooked(appointment.reference);
    } catch (cause) {
      setError(report("failed to confirm booking", cause, { date, slotStart }));
    } finally {
      setBusy(false);
    }
  };

  if (!awaitingCode) {
    return (
      <>
        {error && <Alert>{error}</Alert>}

        <Field
          label="Full name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="e.g. Ritu Kumari"
          autoComplete="name"
        />
        <Field
          label="Mobile number"
          type="tel"
          inputMode="numeric"
          value={phone}
          onChange={(event) => setPhone(event.target.value.replace(/\D/g, "").slice(0, PHONE_DIGITS))}
          placeholder="10-digit number"
          autoComplete="tel-national"
          hint="We send a one-time code to this number to confirm the booking."
          className="tabular"
        />

        <div id="recaptcha-container" />

        <Button
          block
          loading={busy}
          disabled={!name.trim() || phone.length < PHONE_DIGITS}
          onClick={sendCode}
        >
          Send verification code
        </Button>
      </>
    );
  }

  return (
    <>
      {error && <Alert>{error}</Alert>}

      <p className="mb-4 text-sm text-ink-soft">
        Enter the {CODE_DIGITS}-digit code sent to{" "}
        <span className="tabular font-medium text-navy">+91 {phone}</span>.
      </p>

      <Field
        label="Verification code"
        inputMode="numeric"
        value={code}
        onChange={(event) => setCode(event.target.value.replace(/\D/g, "").slice(0, CODE_DIGITS))}
        placeholder="000000"
        autoComplete="one-time-code"
        maxLength={CODE_DIGITS}
        className="tabular text-center text-lg tracking-[0.3em]"
      />

      <Button block loading={busy} disabled={code.length < CODE_DIGITS} onClick={confirmAndBook}>
        Confirm appointment
      </Button>

      <button
        onClick={() => {
          setAwaitingCode(false);
          setCode("");
          setError("");
        }}
        className="mt-4 w-full rounded text-xs text-ink-mute transition-colors hover:text-navy"
      >
        Wrong number? Go back and change it
      </button>
    </>
  );
}
