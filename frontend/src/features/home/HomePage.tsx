import { ChevronRight, Clock, MapPin, ShieldCheck } from "lucide-react";
import type { Doctor } from "../../lib/types";
import { CLINIC, RESIDENT_DOCTORS } from "../../lib/clinic";
import { Button, Card, EmptyState, Eyebrow, Spinner } from "../../components/ui";
import { PhotoRail } from "./PhotoRail";
import { DoctorAvatar } from "../booking/DoctorAvatar";

interface HomePageProps {
  doctors: Doctor[];
  loading: boolean;
  onPickDoctor: (doctorId: string) => void;
}

const FACTS = [
  { Icon: Clock, text: "Open 24 hours, every day" },
  { Icon: MapPin, text: CLINIC.address },
  { Icon: ShieldCheck, text: "In-house laboratory and imaging" },
];

export function HomePage({ doctors, loading, onPickDoctor }: HomePageProps) {
  return (
    <main>
      {/*
        The hero is the first step of the booking flow, not a picture of one.
        The single job of this page is to get a patient a time slot, so the
        doctor list is on screen immediately instead of behind a call to action.
      */}
      {/* pb-24 clears the sticky mobile action bar so the last doctor card
          in the picker is never hidden behind it; md:pb-20 drops that
          reservation once the bar itself disappears at the md breakpoint. */}
      <section className="bg-navy px-5 pt-14 pb-24 md:py-20">
        <div className="mx-auto grid max-w-6xl items-start gap-10 lg:grid-cols-[1.1fr_1fr]">
          <div>
            <p className="eyebrow text-navy-mist">
              Appointments
              <span className="deva text-white/70">अपॉइंटमेंट</span>
            </p>

            <h1 className="mt-4 font-display text-4xl leading-[1.15] text-white md:text-5xl">
              See a doctor today,
              <br />
              without the wait.
            </h1>

            <p className="mt-5 max-w-md text-[15px] leading-relaxed text-[#c6cee2]">
              Pick a doctor, pick a time, confirm with a code sent to your phone. No account,
              no queue, no calling to ask what time is free.
            </p>

            <ul className="mt-8 space-y-3">
              {FACTS.map(({ Icon, text }) => (
                <li key={text} className="flex items-center gap-3 text-sm text-[#c6cee2]">
                  <Icon size={16} className="shrink-0 text-red" aria-hidden />
                  {text}
                </li>
              ))}
            </ul>
          </div>

          <Card className="p-5 shadow-lift md:p-6">
            <Eyebrow en="Step 1 of 4" hi="पहला चरण" />
            <h2 className="mt-2 mb-5 text-xl">Choose a doctor</h2>

            {loading ? (
              <Spinner label="Loading doctors" />
            ) : doctors.length === 0 ? (
              <EmptyState
                title="No doctors are listed right now"
                hint={`Call ${CLINIC.phoneDisplay} and the front desk will book you in.`}
              />
            ) : (
              <ul className="space-y-2.5">
                {doctors.map((doctor) => (
                  <li key={doctor.id}>
                    <button
                      onClick={() => onPickDoctor(doctor.id)}
                      className="group flex w-full items-center gap-3.5 rounded-lg border border-line bg-white px-3.5 py-3 text-left transition-colors hover:border-navy hover:bg-paper"
                    >
                      <DoctorAvatar doctor={doctor} />
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-sm font-medium text-navy">
                          {doctor.name}
                        </span>
                        <span className="block truncate text-xs text-red-deep">
                          {doctor.specialization ?? "General consultation"}
                        </span>
                      </span>
                      <ChevronRight
                        size={16}
                        className="shrink-0 text-ink-mute transition-transform group-hover:translate-x-0.5 group-hover:text-navy"
                        aria-hidden
                      />
                    </button>
                  </li>
                ))}
              </ul>
            )}

            <p className="mt-5 text-xs leading-relaxed text-ink-mute">
              Prefer to speak to someone?{" "}
              <a href={CLINIC.phoneHref} className="font-medium text-red-deep underline underline-offset-2">
                Call {CLINIC.phoneDisplay}
              </a>
              .
            </p>
          </Card>
        </div>
      </section>

      <PhotoRail />

      <section className="mx-auto max-w-6xl px-5 py-16">
        <Eyebrow en="Our doctors" hi="हमारे चिकित्सक" />
        <h2 className="mt-2 mb-10 text-2xl md:text-3xl">Who you will see</h2>

        <ul className="grid gap-6 md:grid-cols-3">
          {RESIDENT_DOCTORS.map((doctor) => (
            <Card as="li" key={doctor.nameLatin} className="flex flex-col overflow-hidden">
              <img
                src={doctor.photo}
                alt={doctor.nameLatin}
                loading="lazy"
                className="aspect-[4/5] w-full bg-paper-sunk object-cover"
              />
              <div className="flex flex-1 flex-col p-5">
                <h3 className="font-deva text-lg leading-snug font-bold text-navy">
                  {doctor.name}
                </h3>
                <p className="mt-0.5 text-xs text-ink-mute">{doctor.nameLatin}</p>

                <p className="mt-3 font-deva text-sm font-medium text-red-deep">{doctor.role}</p>
                <p className="text-xs text-ink-mute">{doctor.roleLatin}</p>

                <ul className="mt-auto space-y-0.5 pt-4 font-deva text-sm text-ink-soft">
                  {doctor.credentials.map((credential) => (
                    <li key={credential}>{credential}</li>
                  ))}
                </ul>
              </div>
            </Card>
          ))}
        </ul>
      </section>

      {/* On a phone the two things a patient came for stay in reach. */}
      <div className="sticky bottom-0 z-20 flex gap-3 border-t border-line bg-white/95 px-5 py-3 backdrop-blur md:hidden">
        <Button variant="ghost" className="flex-1" onClick={() => window.location.assign(CLINIC.phoneHref)}>
          Call the hospital
        </Button>
        <Button
          className="flex-1"
          onClick={() => document.getElementById("root")?.scrollIntoView({ behavior: "smooth" })}
        >
          Book a time
        </Button>
      </div>
    </main>
  );
}
