import { useState, useEffect } from "react";
import { Phone, MapPin, Clock, ChevronRight, ChevronLeft, Check, Loader2 } from "lucide-react";
import { BrowserRouter, Routes, Route, useNavigate, useSearchParams } from "react-router-dom";
import * as api from "./api";
import "./index.css";
import { format, addDays } from "date-fns";

function Logo({ size = 64 }) {
  return (
    <svg width={size} height={size} viewBox="0 0 200 200" className="shrink-0">
      <circle cx="100" cy="100" r="94" className="fill-navy stroke-red" strokeWidth="7" />
      <g fill="none" stroke="#FFFFFF" strokeWidth="13" strokeLinecap="round">
        <path d="M55 68 Q100 60 150 72" />
        <path d="M50 96 Q100 88 152 100" />
        <path d="M52 124 Q98 118 148 128" />
      </g>
      <path d="M100 34 L112 78 Q112 118 100 132 Q88 118 88 78 Z" className="fill-red" />
      <text x="100" y="168" textAnchor="middle" className="font-devanagari text-[34px] font-bold fill-red">
        त्र्यक्ष
      </text>
    </svg>
  );
}

function BookingShell({ title, onBack, doctor, children }: { title: string; onBack: () => void; doctor?: any; children: React.ReactNode }) {
  return (
    <div className="px-6 py-8 max-w-md mx-auto">
      <button onClick={onBack} className="flex items-center gap-1 text-xs mb-4 text-[#8A8880]">
        <ChevronLeft size={14} /> Back
      </button>
      {doctor && <p className="text-xs mb-1 text-red">{doctor.name}</p>}
      <h2 className="font-lora text-navy text-lg mb-5">{title}</h2>
      {children}
    </div>
  );
}

function BookingWizard() {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  const step = searchParams.get("step") || "home";
  const doctorId = searchParams.get("doctor");
  const clinicId = searchParams.get("clinic");
  const date = searchParams.get("date");
  const slotStart = searchParams.get("slot");
  const slotEnd = searchParams.get("slotEnd");

  const [doctors, setDoctors] = useState<any[]>([]);
  const [clinics, setClinics] = useState<any[]>([]);
  const [slots, setSlots] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  // Load doctors and clinics on mount
  useEffect(() => {
    Promise.all([api.fetchDoctors(), api.fetchClinics()])
      .then(([d, c]) => {
        setDoctors(d || []);
        setClinics(c || []);
      })
      .catch(console.error);
  }, []);

  // Load slots when doctor, clinic, and date are selected
  useEffect(() => {
    if (step === "slot" && doctorId && clinicId && date) {
      setLoading(true);
      api.fetchSlots(doctorId, clinicId, date)
        .then(data => setSlots(data.slots || []))
        .catch(console.error)
        .finally(() => setLoading(false));
    }
  }, [step, doctorId, clinicId, date]);

  const goTo = (newStep: string, params: Record<string, string> = {}) => {
    const newParams = new URLSearchParams(searchParams);
    newParams.set("step", newStep);
    Object.entries(params).forEach(([k, v]) => newParams.set(k, v));
    setSearchParams(newParams);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const selectedDoctor = doctors.find(d => d.id === doctorId);
  const selectedClinic = clinics.find(c => c.id === clinicId);

  // Derived dates for the next 7 days
  const availableDates = Array.from({ length: 7 }).map((_, i) => format(addDays(new Date(), i), 'yyyy-MM-dd'));

  if (step === "home") {
    return (
      <main>
        <section className="bg-navy-deep px-6 py-14 text-center">
          <p className="text-[#8FA0C6] text-sm mb-3">Road No 3, Laxmisagar, Darbhanga · Open 24/7</p>
          <h1 className="font-lora text-white text-3xl md:text-4xl leading-snug max-w-xl mx-auto">
            Book your appointment in under a minute
          </h1>
          <p className="text-[#C6CEE2] mt-4 max-w-md mx-auto text-sm">
            Pick a doctor, pick a time, and get a confirmation on WhatsApp. No account, no waiting on hold.
          </p>
          <button
            onClick={() => goTo("doctor")}
            className="bg-red text-white mt-7 px-6 py-3 rounded font-medium text-sm inline-flex items-center gap-2"
          >
            Book an appointment <ChevronRight size={16} />
          </button>
        </section>

        <section className="bg-white border-t border-[#E4E2D9] px-6 py-8">
          <div className="max-w-3xl mx-auto flex flex-wrap gap-6 text-sm text-[#4A4A45]">
            <div className="flex items-center gap-2">
              <MapPin size={16} className="text-red" /> Road No 3, Laxmisagar, Darbhanga, Bihar
            </div>
            <div className="flex items-center gap-2">
              <Clock size={16} className="text-red" /> Open 24 / 7
            </div>
            <div className="flex items-center gap-2">
              <Phone size={16} className="text-red" /> 922 9333 922 (call or WhatsApp)
            </div>
          </div>
        </section>
      </main>
    );
  }

  if (step === "doctor") {
    return (
      <BookingShell title="Select a doctor" onBack={() => navigate("/")}>
        <div className="space-y-3">
          {doctors.map((d) => (
            <div key={d.id} className="flex items-center justify-between gap-4 p-4 rounded bg-white border border-[#E4E2D9]">
              <div>
                <div className="text-navy font-medium">{d.name}</div>
                <div className="text-red-deep text-sm mt-1">{d.specialization}</div>
              </div>
              <button
                onClick={() => goTo("clinic", { doctor: d.id })}
                className="border border-navy text-navy text-sm px-3 py-1.5 rounded whitespace-nowrap"
              >
                Select
              </button>
            </div>
          ))}
          {doctors.length === 0 && <p className="text-sm text-[#8A8880]">No doctors found (Database is empty)</p>}
        </div>
      </BookingShell>
    );
  }

  if (step === "clinic" && selectedDoctor) {
    return (
      <BookingShell title="Select a clinic" onBack={() => goTo("doctor")} doctor={selectedDoctor}>
        <div className="space-y-3">
          {clinics.map((c) => (
            <div key={c.id} className="p-4 rounded bg-white border border-[#E4E2D9]">
              <div className="text-navy font-medium">{c.name}</div>
              <div className="text-[#8A8880] text-sm mt-1">{c.address}</div>
              <button
                onClick={() => goTo("date", { clinic: c.id })}
                className="mt-3 bg-navy text-white text-sm px-4 py-2 rounded"
              >
                Select this clinic
              </button>
            </div>
          ))}
          {clinics.length === 0 && <p className="text-sm text-[#8A8880]">No clinics found</p>}
        </div>
      </BookingShell>
    );
  }

  if (step === "date" && selectedDoctor && selectedClinic) {
    return (
      <BookingShell title="Choose a date" onBack={() => goTo("clinic")} doctor={selectedDoctor}>
        <div className="grid grid-cols-2 gap-2">
          {availableDates.map((d) => {
            const dateObj = new Date(d);
            const label = format(dateObj, "EEE, MMM d");
            return (
              <button
                key={d}
                onClick={() => goTo("slot", { date: d })}
                className={`py-3 rounded text-sm text-navy border ${
                  date === d ? "border-red" : "border-[#E4E2D9]"
                }`}
              >
                {label}
              </button>
            );
          })}
        </div>
      </BookingShell>
    );
  }

  if (step === "slot" && selectedDoctor && selectedClinic && date) {
    return (
      <BookingShell title={`Available times · ${date}`} onBack={() => goTo("date")} doctor={selectedDoctor}>
        {loading ? (
          <div className="flex justify-center p-8"><Loader2 className="animate-spin text-navy" /></div>
        ) : (
          <div className="grid grid-cols-2 gap-2">
            {slots.map((s) => (
              <button
                key={s.start}
                disabled={s.status !== "AVAILABLE"}
                onClick={() => goTo("details", { slot: s.start, slotEnd: s.end })}
                className={`py-2.5 rounded text-sm border ${
                  slotStart === s.start ? "border-red" : "border-[#E4E2D9]"
                } ${
                  s.status !== "AVAILABLE"
                    ? "text-[#B7B5AC] bg-[#F1F0EA]"
                    : "text-navy bg-white"
                }`}
              >
                {s.start.substring(0,5)} {s.status !== "AVAILABLE" && "· booked"}
              </button>
            ))}
            {slots.length === 0 && <p className="text-sm text-[#8A8880] col-span-2 text-center">No slots available</p>}
          </div>
        )}
      </BookingShell>
    );
  }

  if (step === "details" && selectedDoctor && selectedClinic && date && slotStart) {
    return <BookingDetails onBack={() => goTo("slot")} doctor={selectedDoctor} clinicId={selectedClinic.id} date={date} slotStart={slotStart} slotEnd={slotEnd!} />;
  }

  if (step === "confirmed") {
    return (
      <div className="px-6 py-14 max-w-md mx-auto text-center">
        <div className="bg-[#E8F3EC] text-[#1F6B3E] w-14 h-14 rounded-full flex items-center justify-center mx-auto mb-5">
          <Check size={26} />
        </div>
        <h2 className="font-lora text-navy text-xl mb-2">Appointment confirmed</h2>
        <p className="text-sm mb-1 text-[#6B6B64]">
          {selectedDoctor?.name} · {date} · {slotStart?.substring(0,5)}
        </p>
        <div className="mt-5 inline-block px-5 py-2 rounded font-medium tracking-wide text-sm border border-dashed border-red text-red-deep">
          {searchParams.get("ref")}
        </div>
        <p className="text-xs mt-5 text-[#8A8880]">Fee payable at the clinic. Please arrive 10 minutes early.</p>
        <button
          onClick={() => {
            navigate("/");
          }}
          className="mt-8 text-sm px-4 py-2 rounded text-navy border border-[#E4E2D9]"
        >
          Back to home
        </button>
      </div>
    );
  }

  // Fallback if step is broken
  return null;
}

function BookingDetails({ onBack, doctor, clinicId, date, slotStart, slotEnd }: any) {
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [otpStage, setOtpStage] = useState(false);
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [searchParams, setSearchParams] = useSearchParams();

  const handleRequestOTP = async () => {
    setError("");
    setLoading(true);
    try {
      await api.requestOTP(phone);
      setOtpStage(true);
    } catch (err: any) {
      setError(err.message);
    }
    setLoading(false);
  };

  const handleVerifyAndBook = async () => {
    setError("");
    setLoading(true);
    try {
      // 1. Verify OTP
      const { token } = await api.verifyOTP(phone, code);
      // 2. Book
      const booking = await api.bookAppointment({
        doctor_id: doctor.id,
        clinic_id: clinicId,
        appointment_date: date,
        start_time: slotStart,
        end_time: slotEnd,
        patient_name: name,
        patient_phone: phone,
        idempotency_key: `${phone}-${date}-${slotStart}`, // naive idempotency
      }, token);

      // Go to confirmed
      const newParams = new URLSearchParams(searchParams);
      newParams.set("step", "confirmed");
      newParams.set("ref", booking.reference);
      setSearchParams(newParams);
      window.scrollTo({ top: 0, behavior: "smooth" });
    } catch (err: any) {
      setError(err.message);
    }
    setLoading(false);
  };

  return (
    <BookingShell title="Your details" onBack={onBack} doctor={doctor}>
      {!otpStage ? (
        <>
          <label className="block text-xs mb-1 text-[#6B6B64]">Full name</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Ritu Kumari"
            className="w-full px-3 py-2 rounded text-sm mb-4 border border-[#E4E2D9]"
          />
          <label className="block text-xs mb-1 text-[#6B6B64]">Phone number</label>
          <input
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            placeholder="10-digit mobile number"
            className="w-full px-3 py-2 rounded text-sm mb-5 border border-[#E4E2D9]"
          />
          {error && <p className="text-red text-xs mb-3">{error}</p>}
          <button
            disabled={!name || phone.length < 10 || loading}
            onClick={handleRequestOTP}
            className={`w-full py-2.5 flex items-center justify-center gap-2 rounded text-sm font-medium text-white ${
              !name || phone.length < 10 || loading ? "bg-[#D7D5CC]" : "bg-red"
            }`}
          >
            {loading && <Loader2 size={16} className="animate-spin" />}
            Send code via WhatsApp
          </button>
        </>
      ) : (
        <>
          <p className="text-sm mb-4 text-navy">Enter the 6-digit code sent to +91 {phone}</p>
          <input
            value={code}
            onChange={(e) => setCode(e.target.value)}
            placeholder="000000"
            className="w-full px-3 py-2 rounded text-sm mb-5 border border-[#E4E2D9] text-center tracking-widest text-lg"
            maxLength={6}
          />
          {error && <p className="text-red text-xs mb-3">{error}</p>}
          <button
            disabled={code.length < 6 || loading}
            onClick={handleVerifyAndBook}
            className={`w-full py-2.5 flex items-center justify-center gap-2 rounded text-sm font-medium text-white ${
              code.length < 6 || loading ? "bg-[#D7D5CC]" : "bg-red"
            }`}
          >
            {loading && <Loader2 size={16} className="animate-spin" />}
            Confirm Appointment
          </button>
        </>
      )}
    </BookingShell>
  );
}

import Admin from "./Admin";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/admin/*" element={<Admin />} />
        <Route path="/*" element={<PublicLayout />} />
      </Routes>
    </BrowserRouter>
  );
}

function PublicLayout() {
  return (
    <div className="font-ibm bg-paper min-h-screen flex flex-col">
      <header className="bg-navy px-5 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Logo size={44} />
          <div>
            <div className="font-lora text-white text-lg leading-tight">Tryaksh Hospital</div>
            <div className="text-[#B9C2D9] text-xs">&amp; Diagnostics · Darbhanga</div>
          </div>
        </div>
        <a href="tel:+919229333922" className="bg-red text-white flex items-center gap-1.5 text-sm px-3 py-1.5 rounded">
          <Phone size={14} /> <span className="hidden sm:inline">922 9333 922</span>
        </a>
      </header>

      <div className="flex-1">
        <BookingWizard />
      </div>
    </div>
  );
}
