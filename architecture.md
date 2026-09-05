# Clinic Appointment System — Architecture

Go backend · React frontend · PostgreSQL · 3 tables

This is the final plan. It describes what to build, not how to write every line. There is no code here.

---

## 1. What we are building

A website where a patient can book a doctor's appointment in about a minute, and a dashboard where the two doctors manage their own schedules.

**The patient:** picks a doctor, picks a clinic, picks a date, sees the free 15-minute slots for that day, picks one, types their name and phone number, receives a 6-digit code on WhatsApp, types it in, and confirms. They get a reference number like `AR-7K4M92`. They pay ₹500 in cash or UPI at the clinic on the day.

**The doctor:** logs in, and opens a date at a clinic — for example "Adyar, 9 September, 10:00 to 13:00". As soon as that is saved, patients can book those slots. The doctor can also close a day, block a single slot, and see the day's list of patients.

**Fixed facts about this system:**

- 2 doctors, 3 clinics. This will not grow.
- Every appointment is 15 minutes.
- All clinics are in Chennai, so all times are `Asia/Kolkata`.
- No online payment.
- Patients never create an account.

**Not included, on purpose:** prescriptions, medical records, patient accounts, a queue system, or anything that stores clinical information. Storing health data would bring in a completely different set of legal duties. This system books time slots and nothing else.

---

## 2. How it works, in one page

Think of the system as answering two questions.

**Question 1: "When is the doctor free?"**

Nobody types out a list of free slots. The computer works it out each time somebody asks. It takes the hours the doctor opened for that day, chops them into 15-minute pieces, then removes the pieces that are during the lunch break, already booked, blocked, or in the past. What is left is what the patient sees.

This is important because it means changing the schedule is instant. The doctor changes the hours, and the very next patient sees different slots. There is no list to rebuild.

**Question 2: "Can two people book the same slot?"**

No, and the reason is worth understanding because it is the heart of the whole system.

We give the database one rule: *no two active appointments may share the same doctor, clinic, date and start time*. The database enforces this itself. If two patients press Confirm at the exact same moment, the database accepts one and rejects the other with an error. Our code catches that error and tells the second patient "sorry, just taken, pick another time".

We do not write any code to check whether a slot is free before saving. Checking first and saving after is exactly where this kind of bug lives — between the check and the save, someone else can slip in. Letting the database refuse is both simpler and correct.

Everything else in this document is ordinary work around those two ideas.

---

## 3. Technology choices

### Backend — Go

| What | Choice | Why |
|---|---|---|
| Web routing | `chi` | Small, built on Go's standard library, nothing new to learn |
| Database access | `pgx` + `sqlc` | You write plain SQL, `sqlc` turns it into Go functions with real types. Errors show up when you compile instead of when a patient books. |
| Migrations | `golang-migrate` | Numbered `.sql` files that create and change tables |
| Config | `godotenv` + `caarlos0/env` | Reads `.env` into a Go struct |
| Passwords | `argon2id` | The current recommended way to store passwords |
| Logging | `log/slog` | Built into Go |
| Testing | Go's `testing` + `testcontainers-go` | Runs a real PostgreSQL in Docker during tests. Necessary, because the double-booking rule lives in the database — a fake database cannot test it. |

An **ORM** (a library that hides SQL behind Go objects) is deliberately not used. The two or three queries that matter here are the ones ORMs make awkward, and you would end up fighting it.

### Frontend — React

| What | Choice | Why |
|---|---|---|
| Build tool | Vite | Fast, simple, produces plain files you can host anywhere |
| Language | TypeScript | Catches typos and wrong data shapes before you run the code |
| Routing | React Router | The standard choice |
| Server data | TanStack Query | Handles loading states, caching and refetching. Saves a lot of hand-written code. |
| Styling | Tailwind CSS | Styles written next to the markup, no separate CSS files to keep in sync |
| Components | Radix UI primitives | Unstyled building blocks (dropdowns, dialogs) that already handle keyboard and screen-reader behaviour |
| Forms | React Hook Form + Zod | Validation without much code |
| Dates | `date-fns` | Small and predictable |

**Not Next.js.** Next.js includes its own backend. You already have a Go backend, so half of it would sit unused. Vite gives you a folder of static files that Cloudflare hosts for free.

The one thing Next.js would have given you is search-engine visibility on the public pages. You get that by adding a prerender step at build time, which turns the five public pages into real HTML files. That is a small plugin, not a change of architecture.

**One React app, two areas.** The public site lives at `/`, the dashboard at `/admin`. The dashboard code loads only when someone visits `/admin`, so patients never download it.

### Database — PostgreSQL

Used because of one feature: the rule described in §2 is a single line of SQL in PostgreSQL. MySQL cannot express it directly. PostgreSQL also handles dates, times and JSON well, and every hosting provider offers it.

---

## 4. Configuration — `.env` for secrets, the database for the clinic itself

Nothing about how the app *runs* is written into the Go or React source code — that part is still true. What changed is where the clinic's actual content lives: the doctors, the clinics, and who-visits-where are now database rows the doctors can edit themselves, not lines in a config file. `.env` is left holding only what a config file is genuinely for: secrets and machine-level settings.

### 4.1 What is in `.env` now

- Database address, session secret, OTP pepper, WhatsApp/SMS credentials
- Booking rules: slot length (15 minutes), how far ahead patients may book, how close to the appointment booking stops, the cancellation window, the daily limit per phone number, rate limits
- A few clinic-wide display strings that are genuinely global, not per-doctor or per-clinic: clinic name, tagline, contact WhatsApp number, and the fee text shown to patients

### What moved out, and where it lives now

- The doctors — name, qualification, specialisation, experience, bio, photo, active/inactive → the `doctors` table (§6.1)
- The clinics — name, address, directions, phone, Maps link, active/inactive → the `clinics` table (§6.2)
- Which clinics each doctor visits → the `doctor_clinics` table (§6.3)
- The usual weekly hours per doctor per clinic → still a starting-point pattern, now stored as a small JSON column on `doctor_clinics` (or a `weekly_hours` table if it ever needs its own history) instead of a `SCHEDULE_*` line in `.env` — it feeds the same "open the next 4 weeks as usual" button described in §10

The React app still asks the backend for the global settings at load time through `GET /api/v1/config`, and separately for the doctor/clinic directory through `GET /api/v1/doctors` and `GET /api/v1/locations` — so none of it is hard-coded in the frontend, it's just coming from two different sources now instead of one.

### 4.2 Doctors and clinics are database tables, not `.env` entries

**This is the one part of the plan that changed.** The earlier draft kept the doctors and clinics in `.env` and treated them like fixed, unchanging text. In practice a clinic's doctor list and clinic list are not fixed — a doctor goes on leave, a new one joins, a clinic moves address — and every one of those became a redeploy under the old plan. So doctors, clinics, and which doctor visits which clinic now live in three real tables (`doctors`, `clinics`, `doctor_clinics` — see §6.4–§6.6), and `.env` goes back to being what `.env` is normally for: secrets and machine-level configuration.

`.env` now holds only:

- `DATABASE_URL`, `SESSION_SECRET`, `OTP_PEPPER` and the WhatsApp/SMS credentials
- Booking configuration — slot length, booking window, lead time, cancellation window, rate limits
- A handful of clinic-wide display strings that genuinely never change per-doctor or per-location: `CLINIC_NAME`, `CLINIC_TAGLINE`, `CONSULTATION_FEE_DISPLAY`

Everything that describes an actual doctor or an actual clinic — name, qualification, bio, photo, address, phone, Maps link, active/inactive — is a row in a table, editable from the dashboard, no redeploy needed.

**What this buys you:** adding Dr. Mishra's third clinic, or marking a doctor inactive while she's on leave, is a form submission, not a code change. It also means the database can finally enforce that an appointment points at a real doctor and a real clinic — a foreign key, not a startup script guessing at slugs.

**What it costs:** a doctor's public page now needs a join across `appointments → doctors` (or `doctor_clinics`) instead of one table read. That's a normal, indexed join, not a concern at this scale — see §6.11.

### 4.3 IDs are still permanent, they're just real primary keys now

The old rule ("never rename or delete a slug") still holds, it's just enforced by the database instead of by discipline and a startup check.

- `doctors.id` and `clinics.id` are UUIDs, generated once and never reused. `appointments.doctor_id` and `appointments.location_id` are foreign keys pointing at them, so the database physically cannot save an appointment against a doctor or clinic that doesn't exist.
- To stop a doctor practising, set `doctors.active = false` from the dashboard. The row — and every appointment that points at it — stays exactly as it was, so old bookings still show her name.
- Deleting a doctor or clinic row is blocked at the database level as long as any appointment references it (`ON DELETE RESTRICT`). If a doctor genuinely needs to be removed with no trace, that has to be a deliberate, separate decision — not something a stray click in the dashboard can do by accident.

The startup slug-check from the old plan (§4.3 previously) is no longer needed at all — that's the whole point of a foreign key.

---

## 5. The shape of the system

```
   Patient (phone)                       Doctor (laptop or phone)
        │                                        │
        ▼                                        ▼
┌──────────────────────────────────────────────────────────┐
│  React app — plain files on Cloudflare Pages             │
│    public: home · doctors · clinics · book · check       │
│    /admin: loaded only when a doctor visits it           │
└──────────────────────────┬───────────────────────────────┘
                           │  HTTPS, JSON
┌──────────────────────────▼───────────────────────────────┐
│  Go server — one program                                 │
│                                                          │
│   config       secrets + booking rules, read from .env   │
│   middleware   request id · logging · CORS · rate limit  │
│                · login check                              │
│                                                          │
│   directory     doctors, clinics, who visits where        │
│   availability  works out the free slots                  │
│   booking       create · cancel · reschedule · look up    │
│   schedule      the doctor's opening hours per date       │
│   verify        sends and checks the 6-digit code         │
│   admin         login and appointment management          │
│   notify        optional confirmation messages            │
│                                                          │
│   database layer (sqlc)                                  │
└──────────────────────────┬───────────────────────────────┘
                           ▼
      ┌────────────────────────────────────┐      ┌──────────────┐
      │   PostgreSQL — 6 tables            │      │  WhatsApp    │
      │   doctors · clinics · doctor_clinics│      │  (codes)     │
      │   availability · appointments       │      └──────────────┘
      │   phone_verifications              │
      └────────────────────────────────────┘
                           ▲
                           │
      ┌────────────────────────────────────┐
      │  .env — secrets & machine config    │
      │  DATABASE_URL · SESSION_SECRET      │
      │  WhatsApp credentials               │
      │  booking configuration              │
      └────────────────────────────────────┘
```

One program, one database. Doctors and clinics are now rows the doctors can edit from the dashboard, not lines in a config file — `.env` is left holding only what genuinely belongs in an environment file: secrets and machine-level settings. Nothing runs in the background except an optional message sender.

---

## 6. The database — six tables

### 6.1 `doctors`

```sql
doctors
  id                 uuid    primary key
  slug               text    unique not null   -- 'anitha' — used in URLs, never reused
  name               text    not null           -- 'Dr. Anitha Raghunathan'
  qualification      text    not null
  specialization     text    not null
  experience_years    int    not null
  bio                text    not null
  photo_url          text    null
  active             boolean not null default true
  created_at         timestamptz
  updated_at         timestamptz
```

`slug` replaces what used to be the whole identity in `.env` — it's still a short, permanent, URL-friendly handle (`/doctors/anitha`), but now it's a column with a `unique` constraint instead of a naming convention you had to maintain by hand. Editing a doctor's bio, qualification or photo is a `PUT /admin/doctors/{id}` from the dashboard; no redeploy.

`active = false` is how a doctor is retired without breaking history — same idea as before, now a real boolean the dashboard can flip.

### 6.2 `clinics`

```sql
clinics
  id            uuid    primary key
  slug          text    unique not null   -- 'adyar'
  name          text    not null           -- 'Adyar'
  address       text    not null
  directions    text    null
  phone         text    not null
  maps_url      text    null
  active        boolean not null default true
  created_at    timestamptz
  updated_at    timestamptz
```

Same idea as `doctors`. Moving a clinic's address or adding a fourth location is a dashboard form, not a config edit and a restart.

### 6.3 `doctor_clinics` — which doctor visits which clinic

```sql
doctor_clinics
  doctor_id   uuid   not null references doctors(id)  on delete restrict
  clinic_id   uuid   not null references clinics(id)  on delete restrict
  created_at  timestamptz

  primary key (doctor_id, clinic_id)
```

This is the one join table in the system, and the one place the old "no JOINs anywhere" claim no longer holds. It answers "which clinics does Dr. Anitha visit?" and "which doctors visit Adyar?" — both simple, indexed reads off a two-column primary key. Adding or removing a clinic for a doctor is one row inserted or deleted here; it never touches `doctors`, `clinics`, or any appointment.

### 6.4 `availability` — what the doctor opens

One row means "this doctor is at this clinic on this date, during these hours".

```sql
availability
  id             uuid   primary key
  doctor_id      uuid   not null references doctors(id) on delete restrict
  clinic_id      uuid   not null references clinics(id) on delete restrict
  available_date date   not null       -- 2026-09-09
  start_time     time   not null       -- 10:00
  end_time       time   not null       -- 13:00
  break_start    time   null           -- 11:30
  break_end      time   null           -- 11:45
  is_open        boolean not null default true
  note           text   null           -- 'Diwali' — shown to patients when closed
  created_at     timestamptz
  updated_at     timestamptz

  unique (doctor_id, clinic_id, available_date)
```

The `unique` line means one doctor can have only one entry per clinic per day. That prevents two contradictory sets of hours for the same day. The two foreign keys are new: the database itself now refuses to save availability against a doctor or clinic that doesn't exist, which used to be enforced only by discipline around slugs (§4.3).

`is_open = false` means the doctor explicitly closed that day. That is different from having no row at all, which means the day was never opened. The patient sees a different message for each.

There is no slot length column. Every appointment is 15 minutes, and that number lives in `.env`.

### 6.5 `appointments` — the bookings

```sql
appointments
  id                uuid   primary key
  reference         text   unique not null   -- 'AR-7K4M92', given to the patient
  doctor_id         uuid   not null references doctors(id) on delete restrict
  clinic_id         uuid   not null references clinics(id) on delete restrict
  appointment_date  date   not null
  start_time        time   not null
  end_time          time   not null

  patient_name      text   null      -- empty only for a blocked slot
  patient_phone     text   null      -- +919840055512
  patient_email     text   null
  patient_note      text   null      -- optional 'reason for visit'

  is_block          boolean not null default false
  status            text   not null default 'BOOKED'
  cancelled_by      text   null      -- PATIENT or DOCTOR
  cancelled_reason  text   null
  rescheduled_from  uuid   null      -- points at the appointment this replaced
  idempotency_key   text   unique null
  actor_doctor_id   uuid   null references doctors(id)  -- which doctor was "acting" (see §6.6) for any admin-side change

  created_at        timestamptz
  updated_at        timestamptz
```

**Status** is one of four words: `BOOKED`, `COMPLETED`, `CANCELLED`, `NO_SHOW`. Rows are never deleted. Cancelling changes the status, which frees the slot (see §6.8) and keeps the history.

**Blocking a slot** does not need its own table. When the doctor blocks 11:15 for a phone call, we save an ordinary appointment row with `is_block = true` and no patient name. It takes up the slot in exactly the same way, shows greyed out in the dashboard, and is hidden from patients. One extra true/false column replaces a whole table and a second set of code.

**`idempotency_key`** is a random ID the browser creates once, when the patient opens the details form. It is sent with the booking request. If the same request arrives twice — the patient double-taps Confirm, or the phone loses signal and retries — the database rejects the second one because the key is already used, and the server returns the first booking instead of making two.

**Why store `end_time`** when it is always start plus fifteen minutes? So that old appointments still display correctly if you ever change the slot length. One column, and one whole class of confusing bug avoided.

### 6.6 The admin login — no table at all

The password is fixed and shared by both doctors, and changing it is rare enough that redeploying for it is fine. That means there's nothing about the login that needs to survive a restart, which means it doesn't need a table.

`ADMIN_LOGIN_PASSWORD` in `.env` is the password, in the same way `SESSION_SECRET` and `WHATSAPP_TOKEN` already sit in `.env` as plaintext secrets — protected by "never commit `.env`," not by hashing at rest. At startup, the server hashes it once with argon2id and keeps the hash in memory; the plaintext is never written anywhere, including logs. To change the password: edit `.env`, redeploy. Every existing session is invalidated the moment the new process starts (see below), so there's no separate "sign out everywhere" step to remember.

Everything else about a login that used to live in `admin_login` — failed-attempt count, lockout, last-login time — is small, short-lived, and fine to lose on restart, so it lives in a single mutex-guarded struct in the Go process instead of a database row:

```go
var loginState struct {
    mu               sync.Mutex
    failedAttempts   int
    lockedUntil      time.Time
    lastLoginAt      time.Time
    sessionsValidFrom time.Time // set to time.Now() at process start
}
```

`sessionsValidFrom` starts at the moment the server boots, so a redeploy — for a password change or anything else — already signs out every device for both doctors, as a free side effect, without a dedicated "sign out everywhere" flow. `failedAttempts` and `lockedUntil` reset on restart too, which is a fine trade-off: brute-force protection restarting occasionally is a minor loss, not a real weakening, at the volume of login attempts one clinic's admin page will ever see.

There are no roles and no permission levels. Both doctors see everything and can do everything — that hasn't changed.

**This relies on there being exactly one server process**, which §14 already assumes (one small Lightsail/DigitalOcean box). If that ever changes — a second instance for redundancy — the lockout counters and `sessionsValidFrom` would need to move to somewhere shared (Redis, or back into Postgres). Not a concern at this scale, worth remembering if it stops being true.

**Attribution, without an account to attribute to.** With no per-doctor login, the system still can't tell Anitha's actions from Vikram's by *who's logged in* — same as the shared-login version, just more so. So it asks the same lighter question: after signing in, the dashboard asks "Which doctor are you?" and stores that as a claim in the signed cookie (§6.9), switchable any time without re-entering the password. That claim is written to `actor_doctor_id` (§6.5) on any admin action, so "who cancelled this" still has an answer even though there's no login to trace it to.

### 6.7 `phone_verifications` — the 6-digit codes

```sql
phone_verifications
  id                uuid   primary key
  phone             text   not null
  code_hash         text   not null      -- sha256 of the code plus a secret
  channel           text   not null      -- WHATSAPP or SMS
  attempts          int    not null default 0
  expires_at        timestamptz not null -- 5 minutes after sending
  verified_at       timestamptz null
  token_hash        text   unique null
  token_expires_at  timestamptz null
  ip                inet
  created_at        timestamptz
```

One row does two jobs, one after the other. When a code is sent, the row is created with `code_hash` filled in. When the patient types the right code, the same row gets `verified_at` and a `token_hash` — a long random value that the browser keeps for 30 days so the patient does not need a code next time.

The actual 6-digit code is **never stored**. Only a hash of it. When the patient types `492817`, we hash what they typed and compare hashes.

This table connects to nothing else. It answers one question at booking time: has this phone number proved it is real?

### 6.8 The rule that prevents double booking

```sql
CREATE UNIQUE INDEX uniq_active_slot
  ON appointments (doctor_id, clinic_id, appointment_date, start_time)
  WHERE status IN ('BOOKED', 'COMPLETED');
```

In plain words: *no two rows may share the same doctor, clinic, date and start time — but only count the rows whose status is `BOOKED` or `COMPLETED`.*

That `WHERE` clause is what makes this work so neatly. Cancelled appointments do not match it, so they stop taking up the slot the instant they are cancelled. You never delete anything, and the slot frees itself.

When two people confirm the same slot at the same moment, PostgreSQL handles them one after another. The first is saved. The second fails with error code `23505` — "unique violation". Your code catches `23505` and returns HTTP status `409` to the browser, which shows "someone just booked that time".

Four more indexes make the common queries fast:

```sql
appointments (doctor_id, clinic_id, appointment_date) WHERE status = 'BOOKED'
appointments (reference)
appointments (patient_phone, created_at)
availability  (doctor_id, clinic_id, available_date)
```

### 6.9 Logging in without a sessions table

When a doctor logs in with the shared password, the server sends back a cookie containing the time it was issued, which doctor is currently "acting" (§6.6), and a signature made with a secret key. The signature is what stops anyone editing the cookie — including changing the acting-doctor claim by hand.

On each request the server checks the signature, then checks that the issue time is **after** `loginState.sessionsValidFrom` — the in-memory timestamp from §6.6, set once when the process starts. Because that value resets to "now" on every restart, redeploying the server (for a password change or anything else) already signs out every device for both doctors — there's no separate button needed, and nothing to keep in sync with the database.

Switching identity (Anitha ↔ Vikram) re-signs the cookie with the new claim but doesn't touch the password or the session at all — it's a `POST /admin/auth/switch-doctor` that just changes who gets recorded on the next action.

Cookies expire on their own after 24 hours. No sessions table, and no database lookup for the session at all — not even the single-row kind.

### 6.10 Time zones

Every clinic is in Chennai. So an appointment stores a plain date and a plain time — `2026-09-09` and `11:45` — and nothing else. No UTC conversion, no offsets.

`TIMEZONE=Asia/Kolkata` in `.env` is used in exactly two places: deciding what "now" means when hiding slots that have already passed, and working out when to send a reminder.

Keep `time.Now()` inside one small `clock` package. Tests can then pretend it is any date they like, which is the only sane way to test "hide slots in the past".

### 6.11 The whole picture

```
doctors · clinics · doctor_clinics   (the directory — see §6.1–§6.3)
availability     (doctor_id, clinic_id → foreign keys)
appointments     (doctor_id, clinic_id → foreign keys)
phone_verifications

The admin login is not a table at all — see §6.6.
```

Six tables, and now a real web of foreign keys: `doctor_clinics` joins doctors to clinics, and `availability`/`appointments` each point at a real doctor and a real clinic. The admin login is still the one thing that isn't a table at all — it's cheap enough, and rare enough to change, that an environment variable remains the right home for it (§6.6).

---

## 7. Working out the free slots

```
The patient asks:  GET /api/v1/slots?doctor=anitha&location=adyar&date=2026-09-09

 1. Find the availability row for that doctor, clinic and date.
 2. No row?             → "Dr. Anitha is not at Adyar on this date"
 3. Row with is_open=false → "Closed — Diwali"
 4. Chop 10:00–13:00 into 15-minute pieces:
       10:00, 10:15, 10:30, 10:45, 11:00, 11:15, 11:30, 11:45, 12:00, 12:15, 12:30, 12:45
 5. Remove anything inside the break (11:30 and 11:45 go if the break is 11:30–12:00).
 6. Ask the database which of these already have a BOOKED appointment. Mark those "booked".
 7. If the date is today, mark anything earlier than now + 1 hour as "past".
 8. Send back every slot with a label: available, booked, or past.
```

Steps 4, 5 and 7 are pure calculation — no database, no clock passed in from outside. That makes them very easy to test. Write those tests first, before any screen exists. Odd cases to cover: a break that does not line up with the 15-minute grid, an end time that is not a multiple of 15, a day with no break at all, and a day that is entirely in the past.

Steps 1 and 6 are two quick indexed queries. The URL still says `doctor=anitha&location=adyar` — slugs stay in the public API and in the address bar because they're readable and shareable. The server's first step on every request is one indexed lookup, `doctors WHERE slug = 'anitha'`, to get the real `id` it needs for everything after that.

**Booked slots are shown, greyed out, not hidden.** A patient who can see that 11:15 exists and is taken trusts the page more than one watching a list mysteriously shrink.

There is a second endpoint for the calendar, `GET /available-dates`, which returns each date in a range with a count of free slots. That is what greys out the closed days before the patient taps anything.

---

## 8. How a booking happens

```
Patient                 React                  Go server            Database    WhatsApp
   │ picks doctor        │                        │                     │          │
   │ picks clinic        │  GET /available-dates ►│────────────────────►│          │
   │ picks date          │  GET /slots ──────────►│────────────────────►│          │
   │ picks 11:45         │◄─── free / booked ─────┤                     │          │
   │ types name + phone  │                        │                     │          │
   │                     │                        │                     │          │
   │ ─── if this phone was already verified on this device, skip to Confirm ───    │
   │                     │                        │                     │          │
   │ taps "Send code"    │  POST /otp/send ──────►│ check limits        │          │
   │                     │                        │ save hashed code ──►│          │
   │                     │                        │ send code ──────────┼─────────►│
   │◄──────────── 6-digit code arrives ───────────┼─────────────────────┼──────────┤
   │ types the code      │  POST /otp/verify ────►│ compare hashes      │          │
   │                     │◄── verification token ─┤ save token ────────►│          │
   │ taps Confirm        │  POST /appointments ──►│ token matches phone?│          │
   │                     │   + idempotency key    │ INSERT ────────────►│          │
   │                     │                        │      ⚡ 23505 → 409  │          │
   │◄─ AR-7K4M92 ────────┤◄──── 201 created ──────┤                     │          │
   │                     │                        │                     │          │
   │ "₹500 at the clinic"│                        │                     │          │
```

The booking itself is one request and one row inserted. Everything before it is preparation.

**One trade-off to know about.** Typing a code takes 30 to 60 seconds. During that time somebody else could take the slot. We are not reserving the slot during those seconds, because holding a slot means storing an expiry time and running a background job to clean up expired holds — a large amount of machinery for a rare event. Two patients would have to want the same 15 minutes within the same minute at a small clinic.

If it does happen, the patient sees a clear screen: their details are kept, their number is already verified, and picking another time is one tap. If your logs ever show this happening regularly, adding a reservation later is a contained change — one more status value and one cleanup job.

**Cancelling** sets the status to `CANCELLED`. The slot frees itself because of the `WHERE` clause in §6.8. If the patient is cancelling from a device that doesn't already hold a valid 30-day token — a different phone, a cleared browser — `/appointments/{reference}/cancel` asks for the phone number and runs the same OTP send/verify pair as booking before it will cancel anything. This is the same rule (`OTP_REQUIRED_FOR_CANCEL=true`) applied to a fresh device, not a separate code path.

**Rescheduling** is one database transaction: insert the new appointment (the unique rule protects the new slot), then mark the old one cancelled and point `rescheduled_from` at it. If the new slot has gone, the whole transaction is undone and the patient keeps their original appointment. They are never left with neither.

---

## 9. Phone verification

Without payment, nothing stops someone filling a whole day with fake names. A 6-digit code is what replaces the ₹500 as proof that a patient is real.

**How it behaves**

- 6 digits, valid for 5 minutes
- 5 wrong attempts and that code is dead; the patient asks for a new one
- Resend allowed after 30 seconds, 3 codes per number per hour, 10 per IP address per hour
- On success the browser gets a token valid 30 days, so returning patients never see a code
- The doctor booking for a phone caller skips verification completely — they are already talking to the person

**Sent over WhatsApp, not SMS.** They cost roughly the same per message (about ₹0.13). But sending SMS in India requires DLT registration with the telecom operators, which costs around ₹5,900 for the entity and again for the sender name, and takes four to six weeks. WhatsApp costs nothing to set up and takes days to approve. For a clinic this size that decides it.

The patient without WhatsApp calls the clinic and the doctor books them by hand. That path already exists.

**Start the WhatsApp business verification and template approval on day one.** It is the only outside dependency that can hold up your launch. Build against a "console" driver that just prints the code to your terminal while you wait.

---

## 10. The API

Base path `/api/v1`. Everything is JSON. Every error looks the same:

```json
{ "error": { "code": "SLOT_TAKEN", "message": "That time was just booked. Please choose another." },
  "request_id": "01J..." }
```

Error codes: `VALIDATION_FAILED`, `SLOT_TAKEN`, `NOT_AVAILABLE`, `OUTSIDE_BOOKING_WINDOW`, `TOO_MANY_BOOKINGS`, `OTP_INVALID`, `OTP_EXPIRED`, `RATE_LIMITED`, `UNAUTHENTICATED`, `NOT_FOUND`.

### Public — anyone can call these

| Method | Path | What it does |
|---|---|---|
| GET | `/config` | Clinic name, tagline, fee text and booking rules — the handful of things still in `.env` |
| GET | `/doctors` | Active doctors, read from the `doctors` table, each with its `doctor_clinics` joined in |
| GET | `/locations` | Active clinics, read from the `clinics` table |
| GET | `/available-dates?doctor=&location=&from=&to=` | Which dates are open, with a count of free slots |
| GET | `/slots?doctor=&location=&date=` | The slots for one day, each labelled available, booked or past |
| POST | `/otp/send` | Sends a code to a phone number |
| POST | `/otp/verify` | Checks a code, returns a 30-day token |
| POST | `/appointments` | Makes the booking. Needs the token and an idempotency key. |
| POST | `/appointments/lookup` | Reference number plus phone → the appointment |
| POST | `/appointments/{reference}/cancel` | Cancels, if still inside the cancellation window |

### Dashboard — needs a logged-in doctor

| Method | Path | What it does |
|---|---|---|
| POST | `/admin/auth/login` · `/logout` | Logging in and out with the shared password |
| POST | `/admin/auth/switch-doctor` | Sets which doctor is "acting" (§6.6) — no password involved, just re-signs the cookie |
| GET | `/admin/me` | Confirms the session is valid and returns the current acting doctor |
| GET | `/admin/dashboard` | Today's list, next 7 days, warnings |
| GET | `/admin/appointments?doctor=&location=&from=&to=&status=&q=` | Search and filter |
| GET | `/admin/appointments/{id}` | One appointment in full |
| POST | `/admin/appointments` | Book for someone who phoned. No code needed. |
| POST | `/admin/appointments/{id}/cancel` · `/complete` · `/no-show` · `/reschedule` | Change an appointment |
| GET | `/admin/availability?doctor=&location=&from=&to=` | Calendar data |
| PUT | `/admin/availability` | **The main one.** Set the hours for one doctor, one clinic, one date. |
| DELETE | `/admin/availability/{id}` | Close a day |
| POST | `/admin/availability/bulk` | Open many dates at once |
| POST | `/admin/blocks` | Block one slot |
| GET | `/admin/doctors` · `/admin/clinics` | Full list, including inactive ones |
| POST | `/admin/doctors` · `/admin/clinics` | Add a new doctor or clinic |
| PUT | `/admin/doctors/{id}` · `/admin/clinics/{id}` | Edit name, bio, address, photo, etc. — the fields that used to only be editable in `.env` |
| POST | `/admin/doctor-clinics` · DELETE `/admin/doctor-clinics/{doctor_id}/{clinic_id}` | Add or remove a clinic from a doctor's list |

**About the bulk endpoint.** Nobody is going to fill in 365 dates by hand. This takes a doctor, a clinic, a date range, which weekdays, and the hours — then creates one `availability` row per matching date. It arrives pre-filled from the usual weekly pattern in `.env`, so the normal case is one button: *"Open the next 4 weeks as usual."* Individual days are then edited or closed as needed.

Without this the most likely real-world failure is a doctor forgetting to open dates, and patients seeing an empty calendar. The dashboard also shows a warning when nothing is open in the coming week.

---

## 11. What the doctors can do

Both doctors have the same abilities. There is no difference between their accounts.

**Every day:** see today's list of patients, mark someone as seen or as a no-show, book for a phone caller, block a slot.

**Every few weeks:** open the next four weeks with one button, then adjust individual dates — a shorter Friday, a closed Tuesday, a public holiday.

**When something changes:** cancel or move an appointment, close a whole day.

**Closing a day that already has bookings never happens silently.** The screen lists every affected patient by name and time, and offers either moving them one by one, or cancelling all of them together with a message sent to each. The system will not quietly delete somebody's appointment.

The same protection applies to shortening a day's hours and to blocking a slot that is already booked.

---

## 12. Security

| Area | What we do |
|---|---|
| Connection | HTTPS everywhere, with Cloudflare in front |
| Passwords | The one shared password is hashed with argon2id in memory at startup (§6.6). The real password sits in `.env` only, never in the database or in logs. |
| Shared password | One password for both doctors means one leak affects both, and one leak means both need to know to change it. Changing it means editing `.env` and redeploying — which also signs out every device immediately, as a side effect of the restart. Worth a password manager and a line in the runbook about who to tell. |
| Login cookie | Signed so it cannot be edited, `HttpOnly` so JavaScript cannot read it, `Secure`, `SameSite=Strict`, expires after 24 hours |
| Brute force | 5 failed logins locks the shared login for 15 minutes |
| CSRF | A double-submit token on every dashboard action that changes data. Public endpoints do not use cookies, so this does not apply to them. |
| Fake bookings | The phone code is the main defence. Backed up by a daily limit per phone number and 5 bookings per hour per IP address. |
| SQL injection | Every query uses parameters through `sqlc`. Never build SQL by joining strings — not even for a search box. |
| XSS | React escapes text automatically. Ban `dangerouslySetInnerHTML` with a lint rule and it stays safe. |
| Slot tampering | The server always recalculates whether a slot is real and free. It never trusts a time sent by the browser. |
| Patient data | Name, phone, optional email, optional note. Nothing medical. |
| Secrets | `.env` is never committed to Git. `.env.example` is, with the values blanked out. |
| Backups | A `pg_dump` every night, uploaded to object storage. **Practise restoring it once before you launch** — a backup you have never restored is not a backup. |

**One legal note.** India's DPDP Act 2023 applies. In practice for v1: a privacy page explaining what you store and why, a consent checkbox at booking, and a stated retention period. Write these before going live.

---

## 13. Things that can go wrong

| Situation | What happens |
|---|---|
| Two patients confirm the same slot | The database rejects the second. They see a friendly screen, their details are kept, they pick another time in one tap. |
| Patient books from an old, stale page | Same as above. The server always rechecks at the moment of saving. |
| Patient double-taps Confirm | The idempotency key means the second request returns the first booking. Never two. |
| Signal drops mid-booking | The browser retries with the same key. At most one appointment exists. |
| The code never arrives | Resend after 30 seconds, up to 3 times. Then the screen shows the clinic's phone number, and the doctor books by hand. |
| Patient types the wrong code | 5 tries, then that code dies and they request a new one. |
| Slot taken while the patient was typing the code | The 409 screen. Their number is already verified, so re-picking is one tap. |
| Doctor closes a day that has bookings | Blocked until resolved. The affected patients are listed with options. |
| Doctor shortens the hours | Same, listing only the appointments now outside the new hours. |
| Doctor blocks a slot that is booked | Refused, showing the patient's name. Cancel or move that appointment first. |
| Patient cancels too late | Refused politely, with the clinic's number shown. |
| Patient cancels from a new device | Treated like booking: phone number, OTP, then the cancel goes through. |
| Patient does not turn up | The doctor marks it `NO_SHOW`. Repeats from one number are visible in the list. |
| Someone tries to book 20 slots | The daily limit per phone number stops it, and the code stops them using a number they do not own. |
| Patient refreshes the booking page | The choices are in the web address (`/book?doctor=anitha&location=adyar&date=2026-09-09`), so refresh, the back button, and sharing the link all work. |
| The doctor forgets to open any dates | The dashboard warns, and one button opens the next four weeks. |
| Someone tries to delete a doctor or clinic that has appointments | Blocked by the foreign key (`ON DELETE RESTRICT`, §4.3). Retire with `active = false` instead. |

---

## 14. What it costs to run

Prices as of late 2026, in rupees, assuming about 500 bookings a month. Add 18% GST to Indian services.

| Item | Choice | Per month |
|---|---|---|
| Frontend | Cloudflare Pages | ₹0 |
| Backend and database | **One small server in Mumbai** running Go, PostgreSQL and Caddy together — AWS Lightsail (~$5) or DigitalOcean (~$6) | ₹450–550 |
| Backups | Nightly dump to Cloudflare R2 (free tier) | ₹0 |
| WhatsApp codes | About 650 messages | ₹110 |
| Domain name | Spread over the year | ₹75 |
| Error tracking | Sentry free tier | ₹0 |
| **Total** | | **₹700–900** |

**Putting PostgreSQL on the same server as the app** is the biggest saving and is the right choice here, not a shortcut. Managed databases earn their price when you need automatic failover and copies for reading. You need neither for a database that will be under 100 MB after five years.

The database size is worth stating plainly: about 33,000 appointments a year at a few hundred bytes each is roughly **10 MB a year**. The number of tables makes no difference to your bill at all.

**Things that would quietly multiply this**, all normal choices at a bigger scale and wrong here: a managed database (₹2,000–4,000 a month for 100 MB of data), a load balancer (₹1,800 a month for one program), Kubernetes, Redis for sessions (avoided — sessions are a cookie), and Twilio for codes (about three times the price of Indian providers).

There is a nearly free version using Oracle's always-free Mumbai server. It genuinely works, but free capacity in Indian regions is often unavailable and idle machines have been reclaimed before. For a clinic taking real bookings, ₹500 a month for a machine nobody can take away is worth it.

---

## 15. Putting it live

```
You push to GitHub
      ↓
GitHub Actions runs the linters and the tests
      ↓
Builds the Go program into a container, and the React app into plain files
      ↓
React files → Cloudflare Pages
Container   → your server, after running any new migrations
```

Three setups, each with its own `.env`: your laptop (Docker Compose), a staging copy, and production. Migrations run before the new version starts. Add `/healthz` and `/readyz` endpoints so you can tell whether it is alive, and send errors to Sentry.

---

## 16. Folder layout

```
clinic/
  backend/
    cmd/api/main.go              starts everything
    internal/
      config/                    reads and checks .env, holds doctors and clinics
      http/                      routes, middleware, request handlers
      domain/
        availability/            slot calculation (the pure functions)
        booking/                 create, cancel, reschedule
        schedule/                the doctor's opening hours
        verify/                  the 6-digit codes
        admin/                   login
      db/                        sqlc-generated code
      platform/                  clock, whatsapp, email
    db/migrations/               numbered .sql files
    db/queries/                  the SQL that sqlc reads
  frontend/
    src/
      features/site/             home, doctors, clinics
      features/booking/          the 5 steps
      features/status/           look up an appointment
      features/admin/            the dashboard
      components/                buttons, inputs, cards
  .env.example
  docs/
```

Two habits that keep this tidy as it grows:

1. The `domain/` folders should not import anything to do with HTTP or SQL. They take plain values and return plain values. That is what makes the slot logic easy to test.
2. Anything that talks to the outside world (WhatsApp, email) sits behind a small interface in `platform/`, with a fake version for tests. Your tests should never send a real message.

---

## 17. Build order

Build it in this order. Each step should work before you start the next.

| Step | What | Days |
|---|---|---|
| 1 | Project setup, database migrations, reading `.env`, health checks | 2 |
| 2 | Availability: save hours for a date, the bulk open, and **the slot calculation with thorough tests** | 4 |
| 3 | Booking: create, cancel, reschedule, look up. Then a test that fires 100 bookings at one slot at once and checks exactly one succeeds. | 3 |
| 4 | Phone codes: send, verify, limits, the 30-day token | 2 |
| 5 | Patient website: all the pages and the 5-step booking flow, working on a real phone | 6 |
| 6 | Dashboard: login, today's list, the availability calendar and editor, the conflict screen | 5 |
| 7 | Confirmation messages (optional) | 2 |
| 8 | Rate limits, security headers, restore a backup for practice, real `.env` for production | 2 |

**About 26 working days, so roughly 5 weeks** at full time.

Steps 2 and 3 hold all the difficulty in this project. Do them properly, with tests, before you write a single screen. The common way a project like this goes wrong is building the pretty dashboard first and discovering the slot logic is subtly wrong after real patients are using it.

---

## 18. What could go wrong with the project

| Risk | What to do about it |
|---|---|
| WhatsApp approval takes longer than expected | The only thing that can block your launch. Start it on day one and build against a console driver meanwhile. |
| The slot calculation has a subtle bug | Write those tests first, covering breaks that do not align, odd end times, and days in the past. This is the highest-value testing in the project. |
| The doctor forgets to open dates | The dashboard warning and the one-button bulk open. Watch this in the first month. |
| Patients do not show up, since nothing is paid | Reminders and no-show tracking. This is a business problem, not a technical one — decide the clinic's policy before launch. |
| A doctor or clinic slug gets reused or renamed | The `unique` constraint on `doctors.slug` / `clinics.slug` catches a collision immediately; renaming is a deliberate dashboard edit, not a silent `.env` change (§4.3). |
| Someone suggests adding medical records | The scope in §1 is the answer. Storing health information changes the legal picture completely and is a different project. |
