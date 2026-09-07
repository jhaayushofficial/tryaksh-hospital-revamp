import { request } from "../../lib/http";
import type {
  Appointment,
  AppointmentStatus,
  Clinic,
  Doctor,
  DoctorClinic,
  WeeklySchedule,
} from "../../lib/types";

/**
 * Admin endpoints.
 *
 * The session token is kept in localStorage because the admin panel is a
 * separate origin from the API in development, where the HttpOnly cookie the
 * backend also sets is not sent.
 */
const TOKEN_KEY = "admin_token";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function logout(): void {
  localStorage.removeItem(TOKEN_KEY);
}

function adminRequest<T>(path: string, options: Parameters<typeof request>[1] = {}) {
  return request<T>(`/admin${path}`, { ...options, token: getToken() });
}

export async function login(password: string): Promise<void> {
  const { token } = await request<{ token: string }>("/admin/login", { body: { password } });
  localStorage.setItem(TOKEN_KEY, token);
}

export function fetchAllDoctors() {
  return adminRequest<Doctor[]>("/doctors");
}

export function createDoctor(doctor: Partial<Doctor>) {
  return adminRequest<Doctor>("/doctors", { body: doctor });
}

export function fetchAllClinics() {
  return adminRequest<Clinic[]>("/clinics");
}

export function createClinic(clinic: Partial<Clinic>) {
  return adminRequest<Clinic>("/clinics", { body: clinic });
}

export function fetchAppointments(filters: Record<string, string> = {}) {
  const query = new URLSearchParams(filters).toString();
  return adminRequest<Appointment[]>(`/appointments${query ? `?${query}` : ""}`);
}

export function updateAppointmentStatus(
  id: string,
  status: AppointmentStatus,
  reason?: string,
) {
  return adminRequest<Appointment>(`/appointments/${id}/status`, {
    method: "PUT",
    body: { status, reason },
  });
}

export function fetchDoctorClinic(doctorId: string, clinicId: string) {
  return adminRequest<DoctorClinic>(`/doctor-clinics/${doctorId}/${clinicId}`);
}

export function updateDoctorClinicHours(
  doctorId: string,
  clinicId: string,
  defaultHours: WeeklySchedule,
) {
  return adminRequest<DoctorClinic>(`/doctor-clinics/${doctorId}/${clinicId}`, {
    method: "PUT",
    body: { default_hours: defaultHours },
  });
}

export function bulkGenerateAvailability(
  doctorId: string,
  clinicId: string,
  from: string,
  to: string,
) {
  return adminRequest<{ created_blocks: number }>("/availability/bulk", {
    body: { doctor_id: doctorId, clinic_id: clinicId, from, to },
  });
}
