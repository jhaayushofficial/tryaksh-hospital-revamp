const API_BASE = (import.meta as any).env?.VITE_API_BASE || "http://localhost:8080/api/v1";

export async function fetchDoctors() {
  const res = await fetch(`${API_BASE}/doctors`);
  if (!res.ok) throw new Error("Failed to fetch doctors");
  return res.json();
}

export async function fetchClinics() {
  const res = await fetch(`${API_BASE}/locations`);
  if (!res.ok) throw new Error("Failed to fetch clinics");
  return res.json();
}

export async function fetchSlots(doctorId: string, clinicId: string, date: string) {
  const res = await fetch(`${API_BASE}/slots?doctor_id=${doctorId}&clinic_id=${clinicId}&date=${date}`);
  if (!res.ok) throw new Error("Failed to fetch slots");
  return res.json();
}

export async function requestOTP(phone: string) {
  const res = await fetch(`${API_BASE}/auth/request-code`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ phone }),
  });
  if (!res.ok) throw new Error("Failed to request OTP");
  return res.json();
}

export async function verifyOTP(phone: string, code: string) {
  const res = await fetch(`${API_BASE}/auth/verify-code`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ phone, code }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error?.message || err.error || "Failed to verify OTP");
  }
  return res.json();
}

export async function firebaseLogin(idToken: string) {
  const res = await fetch(`${API_BASE}/auth/firebase-login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ idToken }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error?.message || err.error || "Failed to login with Firebase");
  }
  return res.json();
}

export async function bookAppointment(data: {
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
}, token: string) {
  const res = await fetch(`${API_BASE}/appointments`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${token}`
    },
    body: JSON.stringify(data),
  });
  
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error?.message || err.error || "Failed to book appointment");
  }
  return res.json();
}
