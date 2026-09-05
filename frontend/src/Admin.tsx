import { useState, useEffect } from "react";
import { useNavigate, Routes, Route, Link } from "react-router-dom";
import * as adminApi from "./adminApi";
import { Loader2, Users, MapPin, Calendar, LogOut, Clock } from "lucide-react";

function Login({ onLogin }: { onLogin: () => void }) {
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleLogin = async () => {
    setLoading(true);
    try {
      await adminApi.login(password);
      onLogin();
    } catch (err: any) {
      setError(err.message);
    }
    setLoading(false);
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-navy px-6">
      <div className="bg-white p-8 rounded-lg shadow-lg w-full max-w-sm">
        <h2 className="text-2xl font-lora text-navy mb-6 text-center">Admin Login</h2>
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleLogin()}
          placeholder="Admin Password"
          className="w-full px-4 py-2 border border-[#E4E2D9] rounded mb-4 focus:outline-none focus:border-red"
        />
        {error && <p className="text-red text-sm mb-4">{error}</p>}
        <button
          onClick={handleLogin}
          disabled={loading || !password}
          className="w-full bg-red text-white py-2 rounded flex items-center justify-center gap-2"
        >
          {loading && <Loader2 size={16} className="animate-spin" />}
          Login
        </button>
      </div>
    </div>
  );
}

function Dashboard() {
  const [doctors, setDoctors] = useState<any[]>([]);
  const [clinics, setClinics] = useState<any[]>([]);

  useEffect(() => {
    adminApi.fetchAllDoctors().then(setDoctors).catch(console.error);
    adminApi.fetchAllClinics().then(setClinics).catch(console.error);
  }, []);

  return (
    <div className="p-6">
      <h1 className="text-2xl font-lora text-navy mb-6">Overview</h1>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
        <div className="bg-white p-6 rounded shadow border border-[#E4E2D9]">
          <h2 className="text-lg font-medium text-navy flex items-center gap-2 mb-2">
            <Users size={18} /> Doctors
          </h2>
          <p className="text-3xl font-bold text-red-deep">{doctors?.length || 0}</p>
        </div>
        <div className="bg-white p-6 rounded shadow border border-[#E4E2D9]">
          <h2 className="text-lg font-medium text-navy flex items-center gap-2 mb-2">
            <MapPin size={18} /> Clinics
          </h2>
          <p className="text-3xl font-bold text-red-deep">{clinics?.length || 0}</p>
        </div>
      </div>
    </div>
  );
}

function Doctors() {
  const [doctors, setDoctors] = useState<any[]>([]);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");

  const load = () => adminApi.fetchAllDoctors().then(setDoctors).catch(console.error);
  useEffect(() => { load(); }, []);

  const handleAdd = async () => {
    try {
      await adminApi.createDoctor({
        slug: slug,
        name: name,
        qualification: "MBBS",
        specialization: "General",
        experience_years: 5,
        bio: "Bio here",
        active: true
      });
      setName("");
      setSlug("");
      load();
    } catch (e: any) { alert(e.message); }
  };

  return (
    <div className="p-6">
      <h1 className="text-2xl font-lora text-navy mb-6">Doctors</h1>
      <div className="bg-white p-6 rounded border border-[#E4E2D9] mb-6">
        <h2 className="text-lg mb-4">Add Doctor</h2>
        <div className="flex gap-4">
          <input value={name} onChange={e => setName(e.target.value)} placeholder="Name" className="border px-3 py-2 rounded" />
          <input value={slug} onChange={e => setSlug(e.target.value)} placeholder="Slug (e.g. dr-smith)" className="border px-3 py-2 rounded" />
          <button onClick={handleAdd} className="bg-navy text-white px-4 py-2 rounded">Add</button>
        </div>
      </div>
      <div className="space-y-3">
        {doctors?.map(d => (
          <div key={d.id} className="bg-white p-4 rounded border border-[#E4E2D9]">
            <p className="font-bold text-navy">{d.name} <span className="text-sm font-normal text-gray-500">({d.slug})</span></p>
            <p className="text-sm text-gray-600">{d.specialization}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

function Clinics() {
  const [clinics, setClinics] = useState<any[]>([]);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");

  const load = () => adminApi.fetchAllClinics().then(setClinics).catch(console.error);
  useEffect(() => { load(); }, []);

  const handleAdd = async () => {
    try {
      await adminApi.createClinic({
        slug: slug,
        name: name,
        address: "Address",
        phone: "0000000000",
        active: true
      });
      setName("");
      setSlug("");
      load();
    } catch (e: any) { alert(e.message); }
  };

  return (
    <div className="p-6">
      <h1 className="text-2xl font-lora text-navy mb-6">Clinics</h1>
      <div className="bg-white p-6 rounded border border-[#E4E2D9] mb-6">
        <h2 className="text-lg mb-4">Add Clinic</h2>
        <div className="flex gap-4">
          <input value={name} onChange={e => setName(e.target.value)} placeholder="Name" className="border px-3 py-2 rounded" />
          <input value={slug} onChange={e => setSlug(e.target.value)} placeholder="Slug" className="border px-3 py-2 rounded" />
          <button onClick={handleAdd} className="bg-navy text-white px-4 py-2 rounded">Add</button>
        </div>
      </div>
      <div className="space-y-3">
        {clinics?.map(c => (
          <div key={c.id} className="bg-white p-4 rounded border border-[#E4E2D9]">
            <p className="font-bold text-navy">{c.name} <span className="text-sm font-normal text-gray-500">({c.slug})</span></p>
            <p className="text-sm text-gray-600">{c.address}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

function Appointments() {
  const [appointments, setAppointments] = useState<any[]>([]);

  const load = () => adminApi.fetchAdminAppointments({}).then(setAppointments).catch(console.error);
  useEffect(() => { load(); }, []);

  const handleStatus = async (id: string, status: string) => {
    try {
      await adminApi.updateAppointmentStatus(id, status);
      load();
    } catch (e: any) { alert(e.message); }
  };

  return (
    <div className="p-6">
      <h1 className="text-2xl font-lora text-navy mb-6">Appointments</h1>
      <div className="space-y-3">
        {appointments?.map(a => (
          <div key={a.id} className="bg-white p-4 rounded border border-[#E4E2D9] flex justify-between items-center">
            <div>
              <p className="font-bold text-navy">{a.patient_name} <span className="text-sm font-normal text-gray-500">({a.patient_phone})</span></p>
              <p className="text-sm text-gray-600">
                {a.appointment_date ? a.appointment_date.split('T')[0] : ''} at {a.start_time ? a.start_time.split('T')[1].substring(0,5) : ''}
              </p>
              <span className={`text-xs px-2 py-1 rounded mt-2 inline-block ${a.status === 'BOOKED' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-800'}`}>
                {a.status}
              </span>
            </div>
            {a.status === 'BOOKED' && (
              <div className="flex gap-2">
                <button onClick={() => handleStatus(a.id, 'COMPLETED')} className="text-sm bg-green-100 text-green-800 px-3 py-1 rounded">Complete</button>
                <button onClick={() => handleStatus(a.id, 'CANCELLED')} className="text-sm bg-red-100 text-red-800 px-3 py-1 rounded">Cancel</button>
              </div>
            )}
          </div>
        ))}
        {appointments.length === 0 && <p>No appointments found.</p>}
      </div>
    </div>
  );
}

const DAYS_OF_WEEK = ["monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"];

function AvailabilityAdmin() {
  const [doctors, setDoctors] = useState<any[]>([]);
  const [clinics, setClinics] = useState<any[]>([]);
  const [doctorId, setDoctorId] = useState("");
  const [clinicId, setClinicId] = useState("");
  
  const [schedule, setSchedule] = useState<any>({});
  
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    adminApi.fetchAllDoctors().then(d => { setDoctors(d); if (d.length) setDoctorId(d[0].id); }).catch(console.error);
    adminApi.fetchAllClinics().then(c => { setClinics(c); if (c.length) setClinicId(c[0].id); }).catch(console.error);
  }, []);

  useEffect(() => {
    if (!doctorId || !clinicId) return;
    adminApi.fetchDoctorClinic(doctorId, clinicId)
      .then(dc => {
        if (dc && dc.default_hours) setSchedule(dc.default_hours);
        else setSchedule({});
      })
      .catch(e => {
        if (!e.message.includes("not found")) console.error(e);
        setSchedule({});
      });
  }, [doctorId, clinicId]);

  const updateScheduleForDay = (day: string, field: string, value: string) => {
    const newSchedule = { ...schedule };
    if (!newSchedule[day]) newSchedule[day] = [{}];
    if (newSchedule[day].length === 0) newSchedule[day].push({});
    
    if (value === "") {
      delete newSchedule[day][0][field];
    } else {
      newSchedule[day][0][field] = value;
    }
    
    // Clear out empty days
    if (Object.keys(newSchedule[day][0]).length === 0) {
      newSchedule[day] = [];
    }
    setSchedule(newSchedule);
  };

  const clearDay = (day: string) => {
    const newSchedule = { ...schedule };
    newSchedule[day] = [];
    setSchedule(newSchedule);
  };

  const handleSaveSchedule = async () => {
    if (!doctorId || !clinicId) return;
    setLoading(true);
    setMessage("");
    try {
      await adminApi.updateDoctorClinicHours(doctorId, clinicId, schedule);
      setMessage("Successfully saved schedule template!");
    } catch (e: any) {
      setMessage(`Error saving: ${e.message}`);
    }
    setLoading(false);
  };

  const handleBulkGenerate = async () => {
    if (!doctorId || !clinicId) return;
    setLoading(true);
    setMessage("");
    try {
      const from = new Date().toISOString().split("T")[0];
      const to = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split("T")[0];
      
      const res = await fetch("http://localhost:8080/api/v1/admin/availability/bulk", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${localStorage.getItem("admin_token")}`
        },
        body: JSON.stringify({ doctor_id: doctorId, clinic_id: clinicId, from, to })
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "Failed to generate");
      setMessage(`Successfully generated ${data.created_blocks} new 15-min slots for the next 30 days!`);
    } catch (e: any) {
      setMessage(`Error: ${e.message}`);
    }
    setLoading(false);
  };

  return (
    <div className="p-6 pb-20">
      <h1 className="text-2xl font-lora text-navy mb-6">Manage Availability</h1>
      
      <div className="flex gap-4 mb-6 max-w-2xl">
        <div className="flex-1">
          <label className="block text-sm mb-1 text-gray-600">Select Doctor</label>
          <select className="w-full border p-2 rounded" value={doctorId} onChange={e => setDoctorId(e.target.value)}>
            {doctors.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
          </select>
        </div>
        <div className="flex-1">
          <label className="block text-sm mb-1 text-gray-600">Select Clinic</label>
          <select className="w-full border p-2 rounded" value={clinicId} onChange={e => setClinicId(e.target.value)}>
            {clinics.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
        </div>
      </div>

      <div className="bg-white p-6 rounded border border-[#E4E2D9] max-w-4xl mb-6">
        <h2 className="text-lg font-medium text-navy mb-4">Weekly Schedule Template</h2>
        <p className="text-sm text-gray-600 mb-6">Set the default working hours and break time for this doctor at this clinic. Leave blank if not working.</p>
        
        <div className="space-y-4">
          <div className="grid grid-cols-[100px_1fr_1fr_1fr_1fr_80px] gap-4 font-medium text-xs text-gray-500 uppercase tracking-wider pb-2 border-b">
            <div>Day</div>
            <div>Start Time</div>
            <div>End Time</div>
            <div>Break Start</div>
            <div>Break End</div>
            <div></div>
          </div>
          
          {DAYS_OF_WEEK.map(day => {
            const blocks = schedule[day] || [];
            const b = blocks.length > 0 ? blocks[0] : null;
            return (
              <div key={day} className="grid grid-cols-[100px_1fr_1fr_1fr_1fr_80px] gap-4 items-center">
                <div className="capitalize text-sm font-medium text-navy">{day.substring(0,3)}</div>
                <input type="time" className="border px-2 py-1.5 rounded text-sm w-full" value={b?.start || ""} onChange={e => updateScheduleForDay(day, "start", e.target.value)} />
                <input type="time" className="border px-2 py-1.5 rounded text-sm w-full" value={b?.end || ""} onChange={e => updateScheduleForDay(day, "end", e.target.value)} />
                <input type="time" className="border px-2 py-1.5 rounded text-sm w-full" value={b?.break_start || ""} onChange={e => updateScheduleForDay(day, "break_start", e.target.value)} />
                <input type="time" className="border px-2 py-1.5 rounded text-sm w-full" value={b?.break_end || ""} onChange={e => updateScheduleForDay(day, "break_end", e.target.value)} />
                {b && (b.start || b.end) && (
                  <button onClick={() => clearDay(day)} className="text-xs text-red hover:underline">Clear</button>
                )}
              </div>
            );
          })}
        </div>
        
        <div className="mt-8 pt-4 border-t flex justify-end">
          <button 
            onClick={handleSaveSchedule} 
            disabled={loading}
            className="bg-navy text-white px-6 py-2 rounded flex items-center gap-2"
          >
            {loading && <Loader2 size={16} className="animate-spin" />}
            Save Schedule Template
          </button>
        </div>
      </div>

      <div className="bg-red-50 p-6 rounded border border-red-200 max-w-4xl">
        <h2 className="text-lg font-medium text-red-900 mb-2">Apply to Calendar</h2>
        <p className="text-sm text-red-800 mb-4">Once you save the template above, click here to automatically generate 15-minute booking slots in the system for the next 30 days based on these rules.</p>
        <button 
          onClick={handleBulkGenerate} 
          disabled={loading}
          className="bg-red text-white px-6 py-2 rounded flex items-center gap-2"
        >
          {loading && <Loader2 size={16} className="animate-spin" />}
          Generate Slots for Next 30 Days
        </button>
        
        {message && (
          <div className={`mt-4 p-3 rounded text-sm ${message.startsWith("Error") ? "bg-red-100 text-red-800" : "bg-green-100 text-green-800"}`}>
            {message}
          </div>
        )}
      </div>
    </div>
  );
}

export default function Admin() {
  const [isAuth, setIsAuth] = useState(!!localStorage.getItem("admin_token"));
  const navigate = useNavigate();

  if (!isAuth) {
    return <Login onLogin={() => setIsAuth(true)} />;
  }

  const handleLogout = () => {
    adminApi.logout();
    setIsAuth(false);
    navigate("/");
  };

  return (
    <div className="flex min-h-screen bg-paper">
      <aside className="w-64 bg-navy text-white flex flex-col">
        <div className="p-6 border-b border-[#1c2c54]">
          <h2 className="font-lora text-xl">Admin Panel</h2>
        </div>
        <nav className="flex-1 p-4 space-y-2">
          <Link to="/admin" className="flex items-center gap-3 px-4 py-2 hover:bg-[#1c2c54] rounded">
            <Calendar size={18} /> Dashboard
          </Link>
          <Link to="/admin/appointments" className="flex items-center gap-3 px-4 py-2 hover:bg-[#1c2c54] rounded">
            <Calendar size={18} /> Appointments
          </Link>
          <Link to="/admin/doctors" className="flex items-center gap-3 px-4 py-2 hover:bg-[#1c2c54] rounded">
            <Users size={18} /> Doctors
          </Link>
          <Link to="/admin/clinics" className="flex items-center gap-3 px-4 py-2 hover:bg-[#1c2c54] rounded">
            <MapPin size={18} /> Clinics
          </Link>
          <Link to="/admin/availability" className="flex items-center gap-3 px-4 py-2 hover:bg-[#1c2c54] rounded">
            <Clock size={18} /> Availability
          </Link>
        </nav>
        <div className="p-4 border-t border-[#1c2c54]">
          <button onClick={handleLogout} className="flex items-center gap-3 px-4 py-2 text-[#8FA0C6] hover:text-white w-full">
            <LogOut size={18} /> Logout
          </button>
        </div>
      </aside>
      <main className="flex-1 overflow-auto">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/appointments" element={<Appointments />} />
          <Route path="/doctors" element={<Doctors />} />
          <Route path="/clinics" element={<Clinics />} />
          <Route path="/availability" element={<AvailabilityAdmin />} />
        </Routes>
      </main>
    </div>
  );
}
