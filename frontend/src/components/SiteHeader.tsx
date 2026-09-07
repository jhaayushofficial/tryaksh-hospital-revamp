import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Phone } from "lucide-react";
import { CLINIC } from "../lib/clinic";

export function Logo({ size = 44 }: { size?: number }) {
  return (
    <img
      src={CLINIC.logoUrl}
      alt=""
      width={size}
      height={size}
      className="shrink-0 object-contain"
    />
  );
}

/**
 * A live clock in the hospital's own timezone.
 *
 * This is the one piece of chrome that earns its place: the hospital's single
 * most important fact is that it never closes, and a running clock proves it
 * in a way the words "Open 24/7" cannot. It also tells a patient which
 * timezone the slot times on the next screen are in.
 */
function OpenNow() {
  const [now, setNow] = useState(() => new Date());

  useEffect(() => {
    const tick = setInterval(() => setNow(new Date()), 30_000);
    return () => clearInterval(tick);
  }, []);

  const time = new Intl.DateTimeFormat("en-IN", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone: CLINIC.timezone,
  }).format(now);

  return (
    <p className="flex items-center gap-2 text-xs text-navy-mist">
      <span className="relative flex h-1.5 w-1.5" aria-hidden>
        <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-jade opacity-70" />
        <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-jade" />
      </span>
      <span className="tabular text-white">{time}</span>
      <span className="text-navy-mist">Open now</span>
      <span className="font-deva text-[13px] text-white/70">· खुला है</span>
    </p>
  );
}

export function SiteHeader() {
  return (
    <header className="sticky top-0 z-30 border-b border-navy-soft bg-navy/95 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-5 py-3">
        <Link to="/" className="flex items-center gap-3 rounded" aria-label={`${CLINIC.nameFull} — home`}>
          <Logo />
          <span>
            <span className="block font-display text-lg leading-tight text-white">
              {CLINIC.name}
            </span>
            <span className="block text-xs text-navy-mist">
              &amp; Diagnostics · {CLINIC.city}
            </span>
          </span>
        </Link>

        <div className="flex items-center gap-5">
          <span className="hidden md:block">
            <OpenNow />
          </span>
          <a
            href={CLINIC.phoneHref}
            className="inline-flex items-center gap-2 rounded-lg bg-red px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-red-deep"
          >
            <Phone size={15} aria-hidden />
            <span className="hidden sm:inline">{CLINIC.phoneDisplay}</span>
            <span className="sm:hidden">Call</span>
          </a>
        </div>
      </div>
    </header>
  );
}
