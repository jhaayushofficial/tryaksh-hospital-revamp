import { useState } from "react";
import { Phone, MapPin, Clock, ChevronRight, ChevronLeft, Check } from "lucide-react";
import "./index.css";

const doctors = [
  {
    id: "pashupati",
    name: "Dr. Pashupati Mishra",
    nameHi: "डॉ. पशुपति मिश्रा",
    role: "Senior Consultant Physician",
    roleHi: "वरिष्ठ सलाहकार चिकित्सक",
    qual: "MBBS (DMCH), DCP (DMCH)",
  },
  {
    id: "shankar",
    name: "Dr. Shankar Mishra",
    nameHi: "डॉ. शंकर मिश्रा, एम.डी.",
    role: "General Physician & Pathologist",
    roleHi: "जनरल फिजिशियन एवं पैथोलॉजिस्ट",
    qual: "MBBS (BJMC Ahmedabad), MD (PMCH Patna)",
  },
  {
    id: "mahima",
    name: "Dr. Mahima Mishra",
    nameHi: "डॉ. (श्रीमती) महिमा मिश्रा, एम.एस.",
    role: "Obstetrics & Gynaecology",
    roleHi: "स्त्री एवं प्रसूति रोग विशेषज्ञ",
    qual: "MBBS (PMCH Patna), MS (PMCH Patna)",
  },
];

const exampleDates = ["Mon 9", "Tue 10", "Wed 11", "Thu 12", "Fri 13"];
const exampleSlots = [
  { t: "10:00 AM", status: "open" },
  { t: "10:15 AM", status: "open" },
  { t: "10:30 AM", status: "booked" },
  { t: "10:45 AM", status: "open" },
  { t: "11:00 AM", status: "open" },
  { t: "11:15 AM", status: "booked" },
  { t: "11:30 AM", status: "open" },
  { t: "11:45 AM", status: "open" },
];

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
      <text
        x="100"
        y="168"
        textAnchor="middle"
        className="font-devanagari text-[34px] font-bold fill-red"
      >
        त्र्यक्ष
      </text>
    </svg>
  );
}

function PlaceholderNote({ children }: { children: React.ReactNode }) {
  return (
    <div className="border-l-3 border-red bg-[#FDECEC] text-xs px-3 py-2 rounded-r">
      <span className="text-red-deep font-medium">For review: </span>
      <span className="text-[#5B1416]">{children}</span>
    </div>
  );
}

function BookingShell({
  title,
  onBack,
  doctor,
  children,
}: {
  title: string;
  onBack: () => void;
  doctor: any;
  children: React.ReactNode;
}) {
  return (
    <div className="px-6 py-8 max-w-md mx-auto">
      <button onClick={onBack} className="flex items-center gap-1 text-xs mb-4 text-[#8A8880]">
        <ChevronLeft size={14} /> Back
      </button>
      <p className="text-xs mb-1 text-red">{doctor.name}</p>
      <h2 className="font-lora text-navy text-lg mb-5">{title}</h2>
      {children}
    </div>
  );
}

export default function App() {
  const [step, setStep] = useState("home");
  const [doctor, setDoctor] = useState<any>(null);
  const [date, setDate] = useState<string | null>(null);
  const [slot, setSlot] = useState<string | null>(null);
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const reference = "TH-4M92KX";

  const goTo = (s: string) => {
    window.scrollTo({ top: 0, behavior: "smooth" });
    setStep(s);
  };

  return (
    <div className="font-ibm bg-paper min-h-screen">
      <header className="bg-navy px-5 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Logo size={44} />
          <div>
            <div className="font-lora text-white text-lg leading-tight">
              Tryaksh Hospital
            </div>
            <div className="text-[#B9C2D9] text-xs">
              &amp; Diagnostics · Darbhanga
            </div>
          </div>
        </div>
        <a
          href="tel:+919229333922"
          className="bg-red text-white flex items-center gap-1.5 text-sm px-3 py-1.5 rounded"
        >
          <Phone size={14} /> 922 9333 922
        </a>
      </header>

      {step === "home" && (
        <main>
          <section className="bg-navy-deep px-6 py-14 text-center">
            <p className="text-[#8FA0C6] text-sm mb-3">
              Road No 3, Laxmisagar, Darbhanga · Open 24/7
            </p>
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

          <section className="px-6 py-10 max-w-3xl mx-auto">
            <h2 className="font-lora text-navy text-xl mb-1">Our doctors</h2>
            <p className="text-sm mb-6 text-[#6B6B64]">हमारे डॉक्टर</p>
            <div className="space-y-3">
              {doctors.map((d) => (
                <div
                  key={d.id}
                  className="flex items-center justify-between gap-4 p-4 rounded bg-white border border-[#E4E2D9]"
                >
                  <div>
                    <div className="text-navy font-medium">{d.name}</div>
                    <div className="text-[#8A8880] text-xs mt-0.5">{d.nameHi}</div>
                    <div className="text-red-deep text-sm mt-1.5">{d.role}</div>
                    <div className="text-[#8A8880] text-xs">{d.qual}</div>
                  </div>
                  <button
                    onClick={() => {
                      setDoctor(d);
                      goTo("date");
                    }}
                    className="border border-navy text-navy text-sm px-3 py-1.5 rounded whitespace-nowrap"
                  >
                    Book
                  </button>
                </div>
              ))}
            </div>
            <div className="mt-6">
              <PlaceholderNote>
                Doctor photos, full bios and exact OPD services are pulled from the public site text — send
                us photos and any corrections and we'll drop them straight into the <code>doctors</code>{" "}
                table.
              </PlaceholderNote>
            </div>
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
      )}

      {step === "date" && doctor && (
        <BookingShell title="Choose a date" onBack={() => goTo("home")} doctor={doctor}>
          <div className="grid grid-cols-5 gap-2">
            {exampleDates.map((d) => (
              <button
                key={d}
                onClick={() => {
                  setDate(d);
                  goTo("slot");
                }}
                className={`py-3 rounded text-sm text-navy border ${
                  date === d ? "border-red" : "border-[#E4E2D9]"
                }`}
              >
                {d}
              </button>
            ))}
          </div>
          <div className="mt-5">
            <PlaceholderNote>
              These five weekday tiles are a placeholder. Real dates will come from each doctor's saved
              availability.
            </PlaceholderNote>
          </div>
        </BookingShell>
      )}

      {step === "slot" && doctor && (
        <BookingShell title={`Available times · ${date}`} onBack={() => goTo("date")} doctor={doctor}>
          <div className="grid grid-cols-2 gap-2">
            {exampleSlots.map((s) => (
              <button
                key={s.t}
                disabled={s.status === "booked"}
                onClick={() => {
                  setSlot(s.t);
                  goTo("details");
                }}
                className={`py-2.5 rounded text-sm border ${
                  slot === s.t ? "border-red" : "border-[#E4E2D9]"
                } ${
                  s.status === "booked"
                    ? "text-[#B7B5AC] bg-[#F1F0EA]"
                    : "text-navy bg-white"
                }`}
              >
                {s.t} {s.status === "booked" && "· booked"}
              </button>
            ))}
          </div>
        </BookingShell>
      )}

      {step === "details" && doctor && (
        <BookingShell title="Your details" onBack={() => goTo("slot")} doctor={doctor}>
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
          <button
            disabled={!name || phone.length < 10}
            onClick={() => goTo("confirmed")}
            className={`w-full py-2.5 rounded text-sm font-medium text-white ${
              !name || phone.length < 10 ? "bg-[#D7D5CC]" : "bg-red"
            }`}
          >
            Send code &amp; confirm
          </button>
          <div className="mt-4">
            <PlaceholderNote>
              The real flow sends a 6-digit WhatsApp code here before confirming. Skipped in this prototype so reviewers can see the full path quickly.
            </PlaceholderNote>
          </div>
        </BookingShell>
      )}

      {step === "confirmed" && doctor && (
        <div className="px-6 py-14 max-w-md mx-auto text-center">
          <div className="bg-[#E8F3EC] text-[#1F6B3E] w-14 h-14 rounded-full flex items-center justify-center mx-auto mb-5">
            <Check size={26} />
          </div>
          <h2 className="font-lora text-navy text-xl mb-2">Appointment confirmed</h2>
          <p className="text-sm mb-1 text-[#6B6B64]">
            {doctor.name} · {date} · {slot}
          </p>
          <div className="mt-5 inline-block px-5 py-2 rounded font-medium tracking-wide text-sm border border-dashed border-red text-red-deep">
            {reference}
          </div>
          <p className="text-xs mt-5 text-[#8A8880]">
            Fee payable at the clinic. Please arrive 10 minutes early.
          </p>
          <button
            onClick={() => {
              setDoctor(null);
              setDate(null);
              setSlot(null);
              setName("");
              setPhone("");
              goTo("home");
            }}
            className="mt-8 text-sm px-4 py-2 rounded text-navy border border-[#E4E2D9]"
          >
            Back to home
          </button>
        </div>
      )}
    </div>
  );
}
