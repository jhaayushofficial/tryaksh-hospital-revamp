import type { ButtonHTMLAttributes, InputHTMLAttributes, ReactNode } from "react";
import { AlertCircle, CheckCircle2, Loader2 } from "lucide-react";
import { cx } from "../lib/cx";

/**
 * The shared primitives. Everything visual that appears more than twice lives
 * here, so a button in the admin panel and a button in the booking flow cannot
 * drift apart.
 */

type ButtonVariant = "primary" | "secondary" | "ghost" | "danger";

const BUTTON_VARIANTS: Record<ButtonVariant, string> = {
  primary: "bg-red text-white hover:bg-red-deep shadow-raise",
  secondary: "bg-navy text-white hover:bg-navy-soft shadow-raise",
  ghost: "bg-white text-navy border border-line hover:border-navy hover:bg-paper",
  danger: "bg-red-tint text-red-deep border border-red/25 hover:bg-red hover:text-white",
};

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  loading?: boolean;
  block?: boolean;
}

export function Button({
  variant = "primary",
  loading = false,
  block = false,
  disabled,
  className,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      {...props}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      className={cx(
        "inline-flex items-center justify-center gap-2 rounded-lg px-5 py-2.5",
        "text-sm font-medium transition-colors duration-150",
        "disabled:cursor-not-allowed disabled:opacity-45 disabled:shadow-none",
        block && "w-full",
        BUTTON_VARIANTS[variant],
        className,
      )}
    >
      {loading && <Loader2 size={16} className="animate-spin" aria-hidden />}
      {children}
    </button>
  );
}

export function Card({
  children,
  className,
  as: Tag = "div",
}: {
  children: ReactNode;
  className?: string;
  as?: "div" | "li" | "section" | "article";
}) {
  return (
    <Tag className={cx("rounded-card border border-line bg-white", className)}>{children}</Tag>
  );
}

interface FieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  hint?: string;
}

export function Field({ label, hint, id, className, ...props }: FieldProps) {
  const inputId = id ?? `field-${label.toLowerCase().replace(/\s+/g, "-")}`;
  const hintId = hint ? `${inputId}-hint` : undefined;

  return (
    <div className="mb-4">
      <label
        htmlFor={inputId}
        className="mb-1.5 block text-xs font-medium tracking-wide text-ink-soft"
      >
        {label}
      </label>
      <input
        {...props}
        id={inputId}
        aria-describedby={hintId}
        className={cx(
          "w-full rounded-lg border border-line bg-white px-3.5 py-2.5 text-sm text-ink",
          "placeholder:text-ink-mute transition-colors",
          "hover:border-ink-mute focus:border-navy focus:outline-none",
          className,
        )}
      />
      {hint && (
        <p id={hintId} className="mt-1.5 text-xs text-ink-mute">
          {hint}
        </p>
      )}
    </div>
  );
}

/**
 * Errors are announced, not just coloured: a screen reader user gets the same
 * information a sighted user gets from the red panel.
 */
export function Alert({
  tone = "error",
  children,
}: {
  tone?: "error" | "success";
  children: ReactNode;
}) {
  const isError = tone === "error";
  const Icon = isError ? AlertCircle : CheckCircle2;

  return (
    <div
      role={isError ? "alert" : "status"}
      className={cx(
        "mb-4 flex items-start gap-2.5 rounded-lg border px-3.5 py-3 text-sm",
        isError
          ? "border-red/20 bg-red-tint text-red-deep"
          : "border-jade/20 bg-jade-tint text-jade",
      )}
    >
      <Icon size={16} className="mt-0.5 shrink-0" aria-hidden />
      <span>{children}</span>
    </div>
  );
}

export function Spinner({ label = "Loading" }: { label?: string }) {
  return (
    <div className="flex items-center justify-center gap-2 p-10 text-sm text-ink-mute" role="status">
      <Loader2 size={18} className="animate-spin" aria-hidden />
      {label}
    </div>
  );
}

/** Shown wherever a list can legitimately be empty, instead of blank space. */
export function EmptyState({ title, hint }: { title: string; hint?: string }) {
  return (
    <div className="rounded-card border border-dashed border-line bg-paper-sunk/50 px-6 py-10 text-center">
      <p className="text-sm font-medium text-navy">{title}</p>
      {hint && <p className="mt-1.5 text-xs text-ink-mute">{hint}</p>}
    </div>
  );
}

/** The bilingual section label used throughout the public site. */
export function Eyebrow({ en, hi }: { en: string; hi?: string }) {
  return (
    <p className="eyebrow">
      {en}
      {hi && <span className="deva">{hi}</span>}
    </p>
  );
}
