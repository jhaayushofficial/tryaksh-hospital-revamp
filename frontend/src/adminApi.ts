const API_BASE = "http://localhost:8080/api/v1";

// Simple wrapper to include token
async function fetchAdmin(endpoint: string, options: RequestInit = {}) {
  const token = localStorage.getItem("admin_token");
  const headers: Record<string, string> = {
    ...((options.headers as any) || {}),
  };
  
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}/admin${endpoint}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || err.message || "Admin API error");
  }

  // Handle empty responses (like DELETE)
  if (res.status === 204) return null;
  const text = await res.text();
  return text ? JSON.parse(text) : null;
}

export async function login(password: string) {
  const res = await fetch(`${API_BASE}/admin/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ password }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || "Login failed");
  }
  const data = await res.json();
  localStorage.setItem("admin_token", data.token);
  return data;
}

export function logout() {
  localStorage.removeItem("admin_token");
}

export async function fetchAllDoctors() {
  return fetchAdmin("/doctors");
}

export async function createDoctor(data: any) {
  return fetchAdmin("/doctors", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
}

export async function fetchAllClinics() {
  return fetchAdmin("/clinics");
}

export async function createClinic(data: any) {
  return fetchAdmin("/clinics", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
}

export async function fetchAdminAppointments(params: Record<string, string>) {
  const q = new URLSearchParams(params).toString();
  return fetchAdmin(`/appointments?${q}`);
}

export async function updateAppointmentStatus(id: string, status: string, reason?: string) {
  return fetchAdmin(`/appointments/${id}/status`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ status, reason }),
  });
}
