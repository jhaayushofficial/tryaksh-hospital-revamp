import { useState, useEffect } from "react";
import { Phone, MapPin, Clock, ChevronRight, ChevronLeft, Check, Loader2, Mail } from "lucide-react";
import { FaInstagram, FaFacebook, FaLinkedin } from "react-icons/fa";
import { BrowserRouter, Routes, Route, useNavigate, useSearchParams } from "react-router-dom";
import Carousel from 'react-bootstrap/Carousel';
import 'bootstrap/dist/css/bootstrap.min.css';
import * as api from "./api";
import "./index.css";
import { format, addDays } from "date-fns";
import { auth } from "./firebase";
import { RecaptchaVerifier, signInWithPhoneNumber, type ConfirmationResult } from "firebase/auth";

function Logo({ size = 64 }) {
  return (
    <img
      src="https://res.cloudinary.com/w5nagizy/image/upload/f_auto,q_auto/Tryaksh_1"
      alt="Tryaksh Hospital & Diagnostics Logo"
      width={size}
      height={size}
      className="shrink-0 object-contain"
    />
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

    // Clear downstream params to prevent stale state when navigating backwards
    const stepOrder = ["home", "doctor", "clinic", "date", "slot", "details", "confirmed"];
    const targetIdx = stepOrder.indexOf(newStep);
    if (targetIdx <= stepOrder.indexOf("doctor")) {
      newParams.delete("doctor");
      newParams.delete("clinic");
      newParams.delete("date");
      newParams.delete("slot");
      newParams.delete("slotEnd");
    } else if (targetIdx <= stepOrder.indexOf("clinic")) {
      newParams.delete("clinic");
      newParams.delete("date");
      newParams.delete("slot");
      newParams.delete("slotEnd");
    } else if (targetIdx <= stepOrder.indexOf("date")) {
      newParams.delete("date");
      newParams.delete("slot");
      newParams.delete("slotEnd");
    } else if (targetIdx <= stepOrder.indexOf("slot")) {
      newParams.delete("slot");
      newParams.delete("slotEnd");
    }

    // Re-apply the explicit params (they take priority)
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
        
        <section className="w-full py-8">
          <Carousel fade>
            <Carousel.Item>
              <img className="d-block w-100 object-cover h-[600px] md:h-[750px]" src="https://images.unsplash.com/photo-1519494026892-80bbd2d6fd0d?q=80&w=2000&auto=format&fit=crop" alt="Hospital Front" />
              <Carousel.Caption className="d-none d-md-block bg-black/60 rounded mb-4">
                <h5>Modern Facilities</h5>
                <p>Equipped with state-of-the-art medical technology.</p>
              </Carousel.Caption>
            </Carousel.Item>
            <Carousel.Item>
              <img className="d-block w-100 object-cover h-[600px] md:h-[750px]" src="https://images.unsplash.com/photo-1579684385127-1ef15d508118?q=80&w=2000&auto=format&fit=crop" alt="Laboratory" />
              <Carousel.Caption className="d-none d-md-block bg-black/60 rounded mb-4">
                <h5>Advanced Diagnostics</h5>
                <p>24/7 in-house laboratory and imaging services.</p>
              </Carousel.Caption>
            </Carousel.Item>
            <Carousel.Item>
              <img className="d-block w-100 object-cover h-[600px] md:h-[750px]" src="https://images.unsplash.com/photo-1581595220892-b0739db3ba8c?q=80&w=2000&auto=format&fit=crop" alt="Patient Room" />
              <Carousel.Caption className="d-none d-md-block bg-black/60 rounded mb-4">
                <h5>Comfortable Wards</h5>
                <p>Clean and spacious rooms for patient recovery.</p>
              </Carousel.Caption>
            </Carousel.Item>
            <Carousel.Item>
              <img className="d-block w-100 object-cover h-[600px] md:h-[750px]" src="https://res.cloudinary.com/w5nagizy/image/upload/v1788730017/Dr.MahimaMishra_Operating.jpg" alt="Dr. Mahima Mishra Operating" />
              <Carousel.Caption className="d-none d-md-block bg-black/60 rounded mb-4">
                <h5>Expert Surgical Care</h5>
                <p>Dr. Mahima Mishra performing a procedure in our modern OT.</p>
              </Carousel.Caption>
            </Carousel.Item>
            <Carousel.Item>
              <img className="d-block w-100 object-cover h-[600px] md:h-[750px]" src="https://res.cloudinary.com/w5nagizy/image/upload/v1788730017/Dr.Shankar_OPD.jpg" alt="Dr. Shankar Mishra OPD" />
              <Carousel.Caption className="d-none d-md-block bg-black/60 rounded mb-4">
                <h5>Dedicated Patient Consultations</h5>
                <p>Thorough diagnostic practice in our outpatient department.</p>
              </Carousel.Caption>
            </Carousel.Item>
            <Carousel.Item>
              <img className="d-block w-100 object-cover h-[600px] md:h-[750px]" src="https://res.cloudinary.com/w5nagizy/image/upload/v1788730018/Dr_Mahima_OPD.jpg" alt="Dr. Mahima Mishra OPD" />
              <Carousel.Caption className="d-none d-md-block bg-black/60 rounded mb-4">
                <h5>Compassionate Care</h5>
                <p>Personalized attention for every patient.</p>
              </Carousel.Caption>
            </Carousel.Item>
            <Carousel.Item>
              <img className="d-block w-100 object-cover h-[600px] md:h-[750px]" src="https://res.cloudinary.com/w5nagizy/image/upload/v1788730017/Dr.Mahima_Picwithpatient.jpg" alt="Trusted by Patients" />
              <Carousel.Caption className="d-none d-md-block bg-black/60 rounded mb-4">
                <h5>Trusted Relationships</h5>
                <p>Building lasting health relationships within the Darbhanga community.</p>
              </Carousel.Caption>
            </Carousel.Item>
          </Carousel>
        </section>

        <section className="max-w-5xl mx-auto px-6 py-12">
          <h2 className="font-lora text-navy text-2xl md:text-3xl text-center mb-10">Our Team</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {/* Doctor 1 */}
            <div className="bg-white border border-[#E4E2D9] rounded-lg overflow-hidden flex flex-col text-center shadow-sm">
              <div className="h-64 bg-[#F1F0EA] flex items-center justify-center">
                <img className="w-full h-full object-cover" src="https://res.cloudinary.com/w5nagizy/image/upload/v1788728660/Dr.PashupatiMishra.png" alt="डॉ. पशुपति मिश्रा" />
              </div>
              <div className="p-6 flex-1 flex flex-col">
                <h3 className="font-devanagari text-xl text-navy font-bold mb-1">डॉ. पशुपति मिश्रा</h3>
                <p className="text-red-deep font-medium mb-4">वरिष्ठ सलाहकार चिकित्सक</p>
                <div className="mt-auto text-sm text-[#4A4A45] leading-relaxed">
                  <p>एमबी बी एस (डीएमसीएच),</p>
                  <p>डी सी पी (डीएमसीएच)</p>
                </div>
              </div>
            </div>
            
            {/* Doctor 2 */}
            <div className="bg-white border border-[#E4E2D9] rounded-lg overflow-hidden flex flex-col text-center shadow-sm">
              <div className="h-64 bg-[#F1F0EA] flex items-center justify-center">
                <img className="w-full h-full object-cover" src="https://res.cloudinary.com/w5nagizy/image/upload/v1788728756/Dr.ShankarMishra.png" alt="डॉ. शंकर मिश्रा" />
              </div>
              <div className="p-6 flex-1 flex flex-col">
                <h3 className="font-devanagari text-xl text-navy font-bold mb-1">डॉ. शंकर मिश्रा, एम.डी .</h3>
                <p className="text-red-deep font-medium mb-4">जनरल फिजिशियन एवं पैथोलॉजिस्ट</p>
                <div className="mt-auto text-sm text-[#4A4A45] leading-relaxed">
                  <p>एमबीबीएस ( बीजेएमसी अहमदाबाद)</p>
                  <p>एम.डी . (पीएमसीएच पटना)</p>
                </div>
              </div>
            </div>

            {/* Doctor 3 */}
            <div className="bg-white border border-[#E4E2D9] rounded-lg overflow-hidden flex flex-col text-center shadow-sm">
              <div className="h-64 bg-[#F1F0EA] flex items-center justify-center">
                <img className="w-full h-full object-cover" src="https://res.cloudinary.com/w5nagizy/image/upload/v1788728744/Dr.MahimaMishra.png" alt="डॉ. (श्रीमती) महिमा मिश्रा" />
              </div>
              <div className="p-6 flex-1 flex flex-col">
                <h3 className="font-devanagari text-xl text-navy font-bold mb-1">डॉ. (श्रीमती) महिमा मिश्रा , एम.एस.</h3>
                <p className="text-red-deep font-medium mb-4">स्त्री एवं प्रसूति रोग विशेषज्ञ</p>
                <div className="mt-auto text-sm text-[#4A4A45] leading-relaxed">
                  <p>एमबीबीएस ( पीएमसीएच पटना )</p>
                  <p>एम.एस. (पीएमसीएच पटना)</p>
                </div>
              </div>
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
              <div className="flex items-center gap-4">
                {d.photo_url ? (
                  <img src={d.photo_url} alt={d.name} className="w-14 h-14 rounded-full object-cover border border-[#E4E2D9]" />
                ) : (
                  <div className="w-14 h-14 rounded-full bg-[#F1F0EA] flex items-center justify-center text-navy font-medium text-lg border border-[#E4E2D9]">
                    {d.name.replace(/^Dr\.\s*/i, '').substring(0, 1)}
                  </div>
                )}
                <div>
                  <div className="text-navy font-medium">{d.name}</div>
                  <div className="text-red-deep text-sm mt-1">{d.specialization}</div>
                </div>
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

  const [confirmationResult, setConfirmationResult] = useState<ConfirmationResult | null>(null);

  useEffect(() => {
    if (!(window as any).recaptchaVerifier) {
      (window as any).recaptchaVerifier = new RecaptchaVerifier(auth, 'recaptcha-container', {
        size: 'invisible'
      });
    }
  }, []);

  const handleRequestOTP = async () => {
    setError("");
    setLoading(true);
    try {
      const formattedPhone = phone.startsWith("+") ? phone : `+91${phone}`;
      const appVerifier = (window as any).recaptchaVerifier;
      const result = await signInWithPhoneNumber(auth, formattedPhone, appVerifier);
      setConfirmationResult(result);
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
      if (!confirmationResult) throw new Error("Session expired. Please try again.");
      
      // 1. Verify OTP with Firebase
      const result = await confirmationResult.confirm(code);
      const idToken = await result.user.getIdToken();
      
      // 2. Exchange Firebase token for our backend token
      const { token } = await api.firebaseLogin(idToken);
      
      // 3. Book
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
            type="tel"
            inputMode="numeric"
            pattern="\d*"
            value={phone}
            onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))}
            placeholder="10-digit mobile number"
            className="w-full px-3 py-2 rounded text-sm mb-5 border border-[#E4E2D9]"
          />
          {error && <p className="text-red text-xs mb-3">{error}</p>}
          <div id="recaptcha-container"></div>
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
            inputMode="numeric"
            pattern="\d*"
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
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

      <footer className="bg-navy-deep text-[#C6CEE2] py-12 px-6 mt-auto">
        <div className="max-w-5xl mx-auto grid grid-cols-1 md:grid-cols-2 gap-8">
          <div>
            <div className="flex items-center gap-3 mb-4">
              <Logo size={40} />
              <div>
                <div className="font-lora text-white text-lg leading-tight">Tryaksh Hospital</div>
                <div className="text-[#8FA0C6] text-xs">&amp; Diagnostics</div>
              </div>
            </div>
            <p className="text-sm max-w-sm text-[#8FA0C6]">
              Committed to providing world-class healthcare and advanced diagnostics services.
            </p>
          </div>
          
          <div className="flex flex-col md:items-end justify-center">
            <h4 className="text-white font-medium mb-4">Connect With Us</h4>
            <div className="space-y-3 text-sm flex flex-col md:items-end">
              <a href="tel:+919229333922" className="flex items-center gap-2 hover:text-white transition-colors">
                <Phone size={16} /> 922 9333 922
              </a>
              <a href="mailto:shankar.kr.mishra@gmail.com" className="flex items-center gap-2 hover:text-white transition-colors">
                <Mail size={16} /> shankar.kr.mishra@gmail.com
              </a>
              <div className="flex gap-4 mt-2">
                <a href="https://www.instagram.com/tryakshhospitalanddiagnostics/" target="_blank" rel="noreferrer" className="hover:text-white transition-colors">
                  <FaInstagram size={20} />
                </a>
                <a href="https://www.facebook.com/share/1BTSAq8Qoh/?mibextid=wwXIfr" target="_blank" rel="noreferrer" className="hover:text-white transition-colors">
                  <FaFacebook size={20} />
                </a>
                <a href="https://www.linkedin.com/in/dr-shankar-mishra-md-684771160?utm_source=share_via&utm_content=profile&utm_medium=member_ios" target="_blank" rel="noreferrer" className="hover:text-white transition-colors">
                  <FaLinkedin size={20} />
                </a>
              </div>
            </div>
          </div>
        </div>
        <div className="max-w-5xl mx-auto mt-10 pt-6 border-t border-[#314164] text-xs text-center text-[#8FA0C6]">
          &copy; {new Date().getFullYear()} Tryaksh Hospital &amp; Diagnostics. All rights reserved.
        </div>
      </footer>
    </div>
  );
}
