import type { ReactNode } from "react";
import { ChevronLeft } from "lucide-react";
import type { Clinic, Doctor } from "../../lib/types";
import { formatDateShort } from "../../lib/format";
import { cx } from "../../lib/cx";
import { STEPS, type Step } from "./constants";

interface BookingShellProps {
  step: Step;
  title: string;
  onBack: () => void;
  doctor?: Doctor;
  clinic?: Clinic;
  date?: string | null;
  children: ReactNode;
}

/**
 * The frame every booking step shares: how far along you are, what you have
 * already chosen, and a way back. The chosen values stay visible so nobody has
 * to remember what they picked two screens ago.
 */
export function BookingShell({
  step,
  title,
  onBack,
  doctor,
  clinic,
  date,
  children,
}: BookingShellProps) {
  const currentIndex = STEPS.indexOf(step);

  const chosen = [
    doctor?.name,
    clinic?.name,
    date ? formatDateShort(date) : undefined,
  ].filter(Boolean) as string[];

  return (
    <main className="mx-auto max-w-lg px-5 py-8">
      <button
        onClick={onBack}
        className="mb-6 inline-flex items-center gap-1.5 rounded text-xs font-medium text-ink-mute transition-colors hover:text-navy"
      >
        <ChevronLeft size={14} aria-hidden /> Back
      </button>

      {/* Progress is a row of rules, not numbered badges: the count is small
          enough to read at a glance and the steps have names, not numbers. */}
      <ol className="mb-5 flex gap-1.5" aria-label={`Step ${currentIndex + 1} of ${STEPS.length}`}>
        {STEPS.map((name, index) => (
          <li
            key={name}
            aria-current={index === currentIndex ? "step" : undefined}
            className={cx(
              "h-1 flex-1 rounded-full transition-colors",
              index < currentIndex && "bg-navy",
              index === currentIndex && "bg-red",
              index > currentIndex && "bg-line",
            )}
          >
            <span className="sr-only">{name}</span>
          </li>
        ))}
      </ol>

      {chosen.length > 0 && (
        <p className="mb-1 text-xs text-ink-mute">{chosen.join(" · ")}</p>
      )}
      <h1 className="mb-6 text-2xl">{title}</h1>

      <div className="animate-rise">{children}</div>
    </main>
  );
}
