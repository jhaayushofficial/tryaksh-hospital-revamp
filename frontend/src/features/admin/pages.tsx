import { useCallback, useEffect, useState } from "react";
import { Calendar, MapPin, Users } from "lucide-react";
import type { Appointment, AppointmentStatus, Clinic, Doctor, WeeklySchedule } from "../../lib/types";
import { formatDateShort, formatTime, toDateOnly } from "../../lib/format";
import { report } from "../../lib/logger";
import { cx } from "../../lib/cx";
import { Alert, Button, Card, EmptyState, Field, Spinner } from "../../components/ui";
import * as adminApi from "./api";

function PageHeading({ title, subtitle }: { title: string; subtitle?: string }) {
  return (
    <header className="mb-6">
      <h1 className="text-2xl">{title}</h1>
      {subtitle && <p className="mt-1 text-sm text-ink-mute">{subtitle}</p>}
    </header>
  );
}

export function Dashboard() {
  const [counts, setCounts] = useState({ doctors: 0, clinics: 0, upcoming: 0 });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      adminApi.fetchAllDoctors(),
      adminApi.fetchAllClinics(),
      adminApi.fetchAppointments({ status: "BOOKED" }),
    ])
      .then(([doctors, clinics, appointments]) =>
        setCounts({
          doctors: doctors?.length ?? 0,
          clinics: clinics?.length ?? 0,
          upcoming: appointments?.length ?? 0,
        }),
      )
      .catch((cause) => setError(report("failed to load dashboard counts", cause)))
      .finally(() => setLoading(false));
  }, []);

  const tiles = [
    { label: "Doctors", value: counts.doctors, Icon: Users },
    { label: "Clinics", value: counts.clinics, Icon: MapPin },
    { label: "Booked appointments", value: counts.upcoming, Icon: Calendar },
  ];

  return (
    <div className="p-6 md:p-8">
      <PageHeading title="Overview" subtitle="What is currently set up in the system." />
      {error && <Alert>{error}</Alert>}
      {loading ? (
        <Spinner />
      ) : (
        <ul className="grid gap-4 sm:grid-cols-3">
          {tiles.map(({ label, value, Icon }) => (
            <Card as="li" key={label} className="p-5">
              <p className="flex items-center gap-2 text-sm text-ink-soft">
                <Icon size={16} className="text-ink-mute" aria-hidden /> {label}
              </p>
              <p className="tabular mt-2 text-3xl font-medium text-navy">{value}</p>
            </Card>
          ))}
        </ul>
      )}
    </div>
  );
}

/** Shared shape for the two near-identical "list plus add form" screens. */
interface DirectoryConfig<T> {
  title: string;
  subtitle: string;
  fields: { key: string; label: string; placeholder?: string; type?: string }[];
  load: () => Promise<T[]>;
  create: (values: Record<string, string>) => Promise<unknown>;
  primary: (item: T) => string;
  secondary: (item: T) => string;
}

function Directory<T extends { id: string }>({ config }: { config: DirectoryConfig<T> }) {
  const [items, setItems] = useState<T[]>([]);
  const [values, setValues] = useState<Record<string, string>>({});
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const [loading, setLoading] = useState(true);

  const load = useCallback(() => {
    return config
      .load()
      .then((loaded) => setItems(loaded ?? []))
      .catch((cause) => setError(report(`failed to load ${config.title.toLowerCase()}`, cause)))
      .finally(() => setLoading(false));
  }, [config]);

  useEffect(() => {
    load();
  }, [load]);

  const submit = async () => {
    setError("");
    setSaving(true);
    try {
      await config.create(values);
      setValues({});
      await load();
    } catch (cause) {
      setError(report(`failed to create ${config.title.toLowerCase()}`, cause));
    } finally {
      setSaving(false);
    }
  };

  const required = config.fields.map((field) => field.key);
  const complete = required.every((key) => values[key]?.trim());

  return (
    <div className="p-6 md:p-8">
      <PageHeading title={config.title} subtitle={config.subtitle} />
      {error && <Alert>{error}</Alert>}

      <Card className="mb-8 p-5">
        <h2 className="mb-4 text-base font-medium text-navy">Add {config.title.toLowerCase()}</h2>
        <div className="grid gap-x-4 sm:grid-cols-2">
          {config.fields.map((field) => (
            <Field
              key={field.key}
              label={field.label}
              type={field.type}
              placeholder={field.placeholder}
              value={values[field.key] ?? ""}
              onChange={(event) =>
                setValues((current) => ({ ...current, [field.key]: event.target.value }))
              }
            />
          ))}
        </div>
        <Button variant="secondary" loading={saving} disabled={!complete} onClick={submit}>
          Add
        </Button>
      </Card>

      {loading ? (
        <Spinner />
      ) : items.length === 0 ? (
        <EmptyState title={`No ${config.title.toLowerCase()} yet`} hint="Add the first one above." />
      ) : (
        <ul className="space-y-2.5">
          {items.map((item) => (
            <Card as="li" key={item.id} className="px-4 py-3.5">
              <p className="text-sm font-medium text-navy">{config.primary(item)}</p>
              <p className="mt-0.5 text-xs text-ink-mute">{config.secondary(item)}</p>
            </Card>
          ))}
        </ul>
      )}
    </div>
  );
}

const DOCTORS_CONFIG: DirectoryConfig<Doctor> = {
  title: "Doctors",
  subtitle: "Doctors patients can book. Adding one here puts them on the website immediately.",
  fields: [
    { key: "name", label: "Name", placeholder: "Dr. Shankar Mishra" },
    { key: "slug", label: "URL slug", placeholder: "dr-shankar-mishra" },
    { key: "qualification", label: "Qualification", placeholder: "MBBS, MD" },
    { key: "specialization", label: "Specialisation", placeholder: "General physician" },
    { key: "experience_years", label: "Years of experience", type: "number", placeholder: "12" },
    { key: "bio", label: "Short bio", placeholder: "One line shown under the name" },
  ],
  load: adminApi.fetchAllDoctors,
  create: (values) =>
    adminApi.createDoctor({
      slug: values.slug,
      name: values.name,
      qualification: values.qualification,
      specialization: values.specialization,
      experience_years: Number(values.experience_years) || 0,
      bio: values.bio,
      active: true,
    }),
  primary: (doctor) => doctor.name,
  secondary: (doctor) => [doctor.specialization, doctor.slug].filter(Boolean).join(" · "),
};

const CLINICS_CONFIG: DirectoryConfig<Clinic> = {
  title: "Clinics",
  subtitle: "Locations a patient can choose when booking.",
  fields: [
    { key: "name", label: "Name", placeholder: "Tryaksh Laxmisagar" },
    { key: "slug", label: "URL slug", placeholder: "laxmisagar" },
    { key: "address", label: "Address", placeholder: "Road No 3, Laxmisagar" },
    { key: "phone", label: "Phone", placeholder: "+91 92293 33922" },
  ],
  load: adminApi.fetchAllClinics,
  create: (values) =>
    adminApi.createClinic({
      slug: values.slug,
      name: values.name,
      address: values.address,
      phone: values.phone,
      active: true,
    }),
  primary: (clinic) => clinic.name,
  secondary: (clinic) => clinic.address ?? clinic.slug,
};

export const Doctors = () => <Directory config={DOCTORS_CONFIG} />;
export const Clinics = () => <Directory config={CLINICS_CONFIG} />;

const STATUS_STYLES: Record<AppointmentStatus, string> = {
  BOOKED: "bg-navy/5 text-navy",
  COMPLETED: "bg-jade-tint text-jade",
  CANCELLED: "bg-red-tint text-red-deep",
  NO_SHOW: "bg-paper-sunk text-ink-soft",
};

export function Appointments() {
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<string | null>(null);

  const load = useCallback(() => {
    return adminApi
      .fetchAppointments()
      .then((loaded) => setAppointments(loaded ?? []))
      .catch((cause) => setError(report("failed to load appointments", cause)))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const setStatus = async (id: string, status: AppointmentStatus) => {
    setError("");
    setUpdating(id);
    try {
      await adminApi.updateAppointmentStatus(id, status);
      await load();
    } catch (cause) {
      setError(report("failed to update appointment status", cause, { id, status }));
    } finally {
      setUpdating(null);
    }
  };

  return (
    <div className="p-6 md:p-8">
      <PageHeading title="Appointments" subtitle="Everything booked, newest first." />
      {error && <Alert>{error}</Alert>}

      {loading ? (
        <Spinner />
      ) : appointments.length === 0 ? (
        <EmptyState title="No appointments yet" hint="Bookings made on the website appear here." />
      ) : (
        <ul className="space-y-2.5">
          {appointments.map((appointment) => (
            <Card
              as="li"
              key={appointment.id}
              className="flex flex-wrap items-center justify-between gap-4 px-4 py-3.5"
            >
              <div className="min-w-0">
                <p className="text-sm font-medium text-navy">
                  {appointment.patient_name ?? "Blocked slot"}{" "}
                  {appointment.patient_phone && (
                    <span className="tabular text-xs font-normal text-ink-mute">
                      {appointment.patient_phone}
                    </span>
                  )}
                </p>
                <p className="tabular mt-0.5 text-xs text-ink-soft">
                  {formatDateShort(toDateOnly(appointment.appointment_date))} ·{" "}
                  {formatTime(appointment.start_time)} · {appointment.reference}
                </p>
                <span
                  className={cx(
                    "mt-2 inline-block rounded px-2 py-0.5 text-[11px] font-medium",
                    STATUS_STYLES[appointment.status],
                  )}
                >
                  {appointment.status}
                </span>
              </div>

              {appointment.status === "BOOKED" && (
                <div className="flex gap-2">
                  <Button
                    variant="ghost"
                    loading={updating === appointment.id}
                    onClick={() => setStatus(appointment.id, "COMPLETED")}
                  >
                    Mark completed
                  </Button>
                  <Button
                    variant="danger"
                    loading={updating === appointment.id}
                    onClick={() => setStatus(appointment.id, "CANCELLED")}
                  >
                    Cancel
                  </Button>
                </div>
              )}
            </Card>
          ))}
        </ul>
      )}
    </div>
  );
}

const DAYS = ["monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"];
const HOUR_FIELDS = [
  { key: "start", label: "Opens" },
  { key: "end", label: "Closes" },
  { key: "break_start", label: "Break from" },
  { key: "break_end", label: "Break to" },
] as const;

/** How far ahead "generate slots" reaches. */
const GENERATE_DAYS = 30;

export function Availability() {
  const [doctors, setDoctors] = useState<Doctor[]>([]);
  const [clinics, setClinics] = useState<Clinic[]>([]);
  const [doctorId, setDoctorId] = useState("");
  const [clinicId, setClinicId] = useState("");
  const [schedule, setSchedule] = useState<WeeklySchedule>({});
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    Promise.all([adminApi.fetchAllDoctors(), adminApi.fetchAllClinics()])
      .then(([loadedDoctors, loadedClinics]) => {
        setDoctors(loadedDoctors ?? []);
        setClinics(loadedClinics ?? []);
        if (loadedDoctors?.length) setDoctorId(loadedDoctors[0].id);
        if (loadedClinics?.length) setClinicId(loadedClinics[0].id);
      })
      .catch((cause) => setError(report("failed to load doctors and clinics", cause)));
  }, []);

  useEffect(() => {
    if (!doctorId || !clinicId) return;

    adminApi
      .fetchDoctorClinic(doctorId, clinicId)
      .then((link) => setSchedule(link?.default_hours ?? {}))
      .catch((cause) => {
        // A doctor who has never been linked to this clinic is the normal
        // starting state, not a failure worth showing.
        const notLinked =
          typeof cause === "object" && cause !== null && "status" in cause && cause.status === 404;
        if (!notLinked) setError(report("failed to load the schedule", cause));
        setSchedule({});
      });
  }, [doctorId, clinicId]);

  const setDayField = (day: string, field: string, value: string) => {
    setSchedule((current) => {
      const block = { ...(current[day]?.[0] ?? {}) } as Record<string, string>;

      if (value) block[field] = value;
      else delete block[field];

      return { ...current, [day]: Object.keys(block).length ? [block] : [] };
    });
  };

  const clearDay = (day: string) => setSchedule((current) => ({ ...current, [day]: [] }));

  const run = async (action: () => Promise<string>) => {
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      setSuccess(await action());
    } catch (cause) {
      setError(report("availability update failed", cause, { doctorId, clinicId }));
    } finally {
      setBusy(false);
    }
  };

  const saveTemplate = () =>
    run(async () => {
      await adminApi.updateDoctorClinicHours(doctorId, clinicId, schedule);
      return "Weekly hours saved. Generate slots below to open them for booking.";
    });

  const generateSlots = () =>
    run(async () => {
      const today = new Date();
      const from = today.toISOString().split("T")[0];
      const to = new Date(today.getTime() + GENERATE_DAYS * 86_400_000)
        .toISOString()
        .split("T")[0];

      const { created_blocks } = await adminApi.bulkGenerateAvailability(
        doctorId,
        clinicId,
        from,
        to,
      );
      return `Opened ${created_blocks} new blocks for the next ${GENERATE_DAYS} days.`;
    });

  return (
    <div className="p-6 pb-20 md:p-8">
      <PageHeading
        title="Availability"
        subtitle="Set the weekly hours once, then generate the bookable slots from them."
      />

      {error && <Alert>{error}</Alert>}
      {success && <Alert tone="success">{success}</Alert>}

      <div className="mb-6 grid max-w-2xl gap-4 sm:grid-cols-2">
        <label className="block">
          <span className="mb-1.5 block text-xs font-medium text-ink-soft">Doctor</span>
          <select
            className="w-full rounded-lg border border-line bg-white px-3 py-2.5 text-sm"
            value={doctorId}
            onChange={(event) => setDoctorId(event.target.value)}
          >
            {doctors.map((doctor) => (
              <option key={doctor.id} value={doctor.id}>
                {doctor.name}
              </option>
            ))}
          </select>
        </label>
        <label className="block">
          <span className="mb-1.5 block text-xs font-medium text-ink-soft">Clinic</span>
          <select
            className="w-full rounded-lg border border-line bg-white px-3 py-2.5 text-sm"
            value={clinicId}
            onChange={(event) => setClinicId(event.target.value)}
          >
            {clinics.map((clinic) => (
              <option key={clinic.id} value={clinic.id}>
                {clinic.name}
              </option>
            ))}
          </select>
        </label>
      </div>

      <Card className="mb-6 max-w-4xl p-5 md:p-6">
        <h2 className="text-base font-medium text-navy">Weekly hours</h2>
        <p className="mt-1 mb-5 text-sm text-ink-mute">
          Leave a day blank when this doctor does not sit at this clinic.
        </p>

        <div className="space-y-3">
          {DAYS.map((day) => {
            const block = (schedule[day]?.[0] ?? {}) as Record<string, string>;
            const hasHours = Boolean(block.start || block.end);

            return (
              <div
                key={day}
                className="grid items-end gap-3 border-b border-line pb-3 last:border-0 sm:grid-cols-[80px_repeat(4,1fr)_70px]"
              >
                <span className="text-sm font-medium text-navy capitalize">{day.slice(0, 3)}</span>
                {HOUR_FIELDS.map((field) => (
                  <label key={field.key} className="block">
                    <span className="mb-1 block text-[11px] text-ink-mute sm:sr-only">
                      {field.label}
                    </span>
                    <input
                      type="time"
                      aria-label={`${day} ${field.label}`}
                      className="tabular w-full rounded-lg border border-line bg-white px-2 py-1.5 text-sm"
                      value={block[field.key] ?? ""}
                      onChange={(event) => setDayField(day, field.key, event.target.value)}
                    />
                  </label>
                ))}
                {hasHours && (
                  <button
                    onClick={() => clearDay(day)}
                    className="rounded text-xs text-red-deep hover:underline"
                  >
                    Clear
                  </button>
                )}
              </div>
            );
          })}
        </div>

        <div className="mt-6 flex justify-end">
          <Button variant="secondary" loading={busy} disabled={!doctorId || !clinicId} onClick={saveTemplate}>
            Save weekly hours
          </Button>
        </div>
      </Card>

      <Card className="max-w-4xl border-red/20 bg-red-tint p-5 md:p-6">
        <h2 className="text-base font-medium text-red-deep">Open slots for booking</h2>
        <p className="mt-1 mb-4 text-sm text-red-deep/80">
          Turns the hours above into bookable slots for the next {GENERATE_DAYS} days. Existing
          bookings are left alone.
        </p>
        <Button loading={busy} disabled={!doctorId || !clinicId} onClick={generateSlots}>
          Generate slots
        </Button>
      </Card>
    </div>
  );
}
