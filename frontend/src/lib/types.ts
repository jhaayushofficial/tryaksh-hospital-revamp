/** Shapes returned by the Tryaksh API, mirrored from the Go handlers. */

export interface Doctor {
  id: string;
  slug: string;
  name: string;
  qualification?: string | null;
  specialization?: string | null;
  experience_years?: number | null;
  bio?: string | null;
  photo_url?: string | null;
  active: boolean;
}

export interface Clinic {
  id: string;
  slug: string;
  name: string;
  address?: string | null;
  phone?: string | null;
  active: boolean;
}

export type SlotStatus = "AVAILABLE" | "BOOKED" | "PAST";

export interface Slot {
  /** Local clinic time, "HH:MM". */
  start: string;
  end: string;
  status: SlotStatus;
}

export type AppointmentStatus = "BOOKED" | "COMPLETED" | "CANCELLED" | "NO_SHOW";

export interface Appointment {
  id: string;
  reference: string;
  doctor_id: string;
  clinic_id: string;
  appointment_date: string;
  start_time: string;
  end_time: string;
  patient_name?: string | null;
  patient_phone?: string | null;
  status: AppointmentStatus;
}

export interface BookingRequest {
  doctor_id: string;
  clinic_id: string;
  appointment_date: string;
  start_time: string;
  end_time: string;
  patient_name: string;
  patient_phone: string;
  patient_email?: string;
  patient_note?: string;
  idempotency_key: string;
}

/** A day's working hours for one doctor at one clinic. */
export interface HoursBlock {
  start?: string;
  end?: string;
  break_start?: string;
  break_end?: string;
}

/** Keyed by lowercase weekday name: monday … sunday. */
export type WeeklySchedule = Record<string, HoursBlock[]>;

export interface DoctorClinic {
  doctor_id: string;
  clinic_id: string;
  default_hours?: WeeklySchedule | null;
}
