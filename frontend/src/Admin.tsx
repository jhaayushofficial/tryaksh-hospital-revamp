import { useState, useEffect } from "react";
import { useNavigate, Routes, Route, Link } from "react-router-dom";
import * as adminApi from "./adminApi";
import { Loader2, Users, MapPin, Calendar, LogOut } from "lucide-react";

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
              <p className="text-sm text-gray-600">{a.appointment_date} at {a.start_time.substring(0,5)}</p>
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
        </Routes>
      </main>
    </div>
  );
}
