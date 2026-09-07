import { useState } from "react";
import { NavLink, Route, Routes, useNavigate } from "react-router-dom";
import { Calendar, ClipboardList, Clock, LogOut, MapPin, Users } from "lucide-react";
import { CLINIC } from "../../lib/clinic";
import { report } from "../../lib/logger";
import { cx } from "../../lib/cx";
import { Alert, Button, Card, Field } from "../../components/ui";
import * as adminApi from "./api";
import { Appointments, Availability, Clinics, Dashboard, Doctors } from "./pages";

const NAV = [
  { to: "/admin", end: true, label: "Overview", Icon: ClipboardList },
  { to: "/admin/appointments", label: "Appointments", Icon: Calendar },
  { to: "/admin/doctors", label: "Doctors", Icon: Users },
  { to: "/admin/clinics", label: "Clinics", Icon: MapPin },
  { to: "/admin/availability", label: "Availability", Icon: Clock },
];

function Login({ onSignedIn }: { onSignedIn: () => void }) {
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    setError("");
    setBusy(true);
    try {
      await adminApi.login(password);
      onSignedIn();
    } catch (cause) {
      setError(report("admin login failed", cause));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-navy px-5">
      <Card className="w-full max-w-sm p-7 shadow-lift">
        <p className="eyebrow">Staff only</p>
        <h1 className="mt-2 mb-6 text-xl">Sign in to {CLINIC.name}</h1>

        {error && <Alert>{error}</Alert>}

        <Field
          label="Password"
          type="password"
          value={password}
          autoComplete="current-password"
          onChange={(event) => setPassword(event.target.value)}
          onKeyDown={(event) => event.key === "Enter" && password && submit()}
        />

        <Button block loading={busy} disabled={!password} onClick={submit}>
          Sign in
        </Button>
      </Card>
    </div>
  );
}

export default function Admin() {
  const [signedIn, setSignedIn] = useState(() => Boolean(adminApi.getToken()));
  const navigate = useNavigate();

  if (!signedIn) return <Login onSignedIn={() => setSignedIn(true)} />;

  const signOut = () => {
    adminApi.logout();
    setSignedIn(false);
    navigate("/");
  };

  return (
    <div className="flex min-h-screen flex-col bg-paper md:flex-row">
      <aside className="flex shrink-0 flex-col bg-navy text-white md:w-60">
        <div className="border-b border-navy-soft px-5 py-5">
          <p className="font-display text-lg leading-tight">{CLINIC.name}</p>
          <p className="text-xs text-navy-mist">Staff dashboard</p>
        </div>

        <nav className="flex gap-1 overflow-x-auto p-3 md:flex-1 md:flex-col md:overflow-visible">
          {NAV.map(({ to, end, label, Icon }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              className={({ isActive }) =>
                cx(
                  "flex items-center gap-3 rounded-lg px-3.5 py-2.5 text-sm whitespace-nowrap transition-colors",
                  isActive ? "bg-navy-soft text-white" : "text-navy-mist hover:bg-navy-soft/60 hover:text-white",
                )
              }
            >
              <Icon size={17} aria-hidden /> {label}
            </NavLink>
          ))}
        </nav>

        <div className="border-t border-navy-soft p-3">
          <button
            onClick={signOut}
            className="flex w-full items-center gap-3 rounded-lg px-3.5 py-2.5 text-sm text-navy-mist transition-colors hover:bg-navy-soft/60 hover:text-white"
          >
            <LogOut size={17} aria-hidden /> Sign out
          </button>
        </div>
      </aside>

      <main className="min-w-0 flex-1 overflow-auto">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="appointments" element={<Appointments />} />
          <Route path="doctors" element={<Doctors />} />
          <Route path="clinics" element={<Clinics />} />
          <Route path="availability" element={<Availability />} />
        </Routes>
      </main>
    </div>
  );
}
