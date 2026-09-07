import type { Doctor } from "../../lib/types";

/** Initial of the doctor's name, ignoring the "Dr." prefix. */
function initial(name: string): string {
  return name.replace(/^Dr\.?\s*/i, "").charAt(0).toUpperCase();
}

export function DoctorAvatar({ doctor, size = 44 }: { doctor: Doctor; size?: number }) {
  const dimensions = { width: size, height: size };

  if (doctor.photo_url) {
    return (
      <img
        src={doctor.photo_url}
        alt=""
        style={dimensions}
        className="shrink-0 rounded-full border border-line object-cover"
      />
    );
  }

  return (
    <span
      style={dimensions}
      aria-hidden
      className="flex shrink-0 items-center justify-center rounded-full border border-line bg-paper-sunk font-display text-lg text-navy"
    >
      {initial(doctor.name)}
    </span>
  );
}
