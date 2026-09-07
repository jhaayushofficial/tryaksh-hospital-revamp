import { request } from "./http";
import type { Appointment, BookingRequest, Clinic, Doctor, Slot } from "./types";

/** Public (patient-facing) endpoints. */

export function fetchDoctors(signal?: AbortSignal) {
  return request<Doctor[]>("/doctors", { signal });
}

export function fetchClinics(signal?: AbortSignal) {
  return request<Clinic[]>("/locations", { signal });
}

export function fetchSlots(
  doctorId: string,
  clinicId: string,
  date: string,
  signal?: AbortSignal,
) {
  const query = new URLSearchParams({ doctor_id: doctorId, clinic_id: clinicId, date });
  return request<{ slots: Slot[] }>(`/slots?${query}`, { signal });
}

/** Exchanges a verified Firebase ID token for a Tryaksh session token. */
export function firebaseLogin(idToken: string) {
  return request<{ token: string; phone: string }>("/auth/firebase-login", {
    body: { idToken },
  });
}

export function bookAppointment(booking: BookingRequest, token: string) {
  return request<Appointment>("/appointments", { body: booking, token });
}
