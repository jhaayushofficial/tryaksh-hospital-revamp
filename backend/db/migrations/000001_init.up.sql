-- Tryaksh Clinic Appointment System — Initial Schema
-- 6 tables: doctors, clinics, doctor_clinics, availability, appointments, phone_verifications

BEGIN;

-- =============================================================================
-- 1. doctors
-- =============================================================================
CREATE TABLE doctors (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug             TEXT        UNIQUE NOT NULL,
    name             TEXT        NOT NULL,
    qualification    TEXT        NOT NULL,
    specialization   TEXT        NOT NULL,
    experience_years INT         NOT NULL,
    bio              TEXT        NOT NULL,
    photo_url        TEXT,
    active           BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =============================================================================
-- 2. clinics
-- =============================================================================
CREATE TABLE clinics (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       TEXT        UNIQUE NOT NULL,
    name       TEXT        NOT NULL,
    address    TEXT        NOT NULL,
    directions TEXT,
    phone      TEXT        NOT NULL,
    maps_url   TEXT,
    active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =============================================================================
-- 3. doctor_clinics — which doctor visits which clinic
-- =============================================================================
CREATE TABLE doctor_clinics (
    doctor_id     UUID        NOT NULL REFERENCES doctors(id) ON DELETE RESTRICT,
    clinic_id     UUID        NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    default_hours JSONB,      -- weekly pattern for bulk-open, e.g. {"monday":[{"start":"10:00","end":"13:00"}]}
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (doctor_id, clinic_id)
);

-- =============================================================================
-- 4. availability — what the doctor opens
--    NO unique constraint on (doctor_id, clinic_id, available_date):
--    a doctor can have multiple blocks per day (morning + evening).
-- =============================================================================
CREATE TABLE availability (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_id      UUID        NOT NULL REFERENCES doctors(id) ON DELETE RESTRICT,
    clinic_id      UUID        NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    available_date DATE        NOT NULL,
    start_time     TIME        NOT NULL,
    end_time       TIME        NOT NULL,
    break_start    TIME,
    break_end      TIME,
    is_open        BOOLEAN     NOT NULL DEFAULT TRUE,
    note           TEXT,       -- e.g. 'Diwali' — shown to patients when closed
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Times must be valid
    CONSTRAINT chk_availability_times CHECK (start_time < end_time),

    -- Break fields must be both null or both non-null
    CONSTRAINT chk_availability_break_pair CHECK (
        (break_start IS NULL) = (break_end IS NULL)
    ),

    -- If break exists, it must be within the block and valid
    CONSTRAINT chk_availability_break_range CHECK (
        break_start IS NULL OR (
            break_start >= start_time
            AND break_end <= end_time
            AND break_start < break_end
        )
    )
);

-- Fast lookups for a doctor's availability on a date
CREATE INDEX idx_availability_lookup
    ON availability (doctor_id, clinic_id, available_date);

-- =============================================================================
-- 5. appointments — the bookings
-- =============================================================================
CREATE TABLE appointments (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    reference        TEXT        UNIQUE NOT NULL,   -- 'BC-7K4M92'
    doctor_id        UUID        NOT NULL REFERENCES doctors(id) ON DELETE RESTRICT,
    clinic_id        UUID        NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    appointment_date DATE        NOT NULL,
    start_time       TIME        NOT NULL,
    end_time         TIME        NOT NULL,

    -- Patient info (null for blocked slots)
    patient_name     TEXT,
    patient_phone    TEXT,
    patient_email    TEXT,
    patient_note     TEXT,

    is_block         BOOLEAN     NOT NULL DEFAULT FALSE,
    status           TEXT        NOT NULL DEFAULT 'BOOKED',
    cancelled_by     TEXT,       -- PATIENT or DOCTOR
    cancelled_reason TEXT,
    rescheduled_from UUID        REFERENCES appointments(id),
    idempotency_key  TEXT        UNIQUE,
    actor_doctor_id  UUID        REFERENCES doctors(id),  -- which doctor performed this action

    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_appointment_status CHECK (
        status IN ('BOOKED', 'COMPLETED', 'CANCELLED', 'NO_SHOW')
    ),

    CONSTRAINT chk_appointment_cancelled_by CHECK (
        cancelled_by IS NULL OR cancelled_by IN ('PATIENT', 'DOCTOR')
    )
);

-- THE CRITICAL CONSTRAINT: prevents double-booking.
-- Uses <> 'CANCELLED' so that BOOKED, COMPLETED, and NO_SHOW all occupy the slot.
-- Only CANCELLED appointments free the slot.
CREATE UNIQUE INDEX uniq_active_slot
    ON appointments (doctor_id, clinic_id, appointment_date, start_time)
    WHERE status <> 'CANCELLED';

-- Fast lookup for booked appointments on a date (used by slot calculation)
CREATE INDEX idx_appointments_lookup
    ON appointments (doctor_id, clinic_id, appointment_date)
    WHERE status = 'BOOKED';

-- Lookup by reference code
CREATE INDEX idx_appointments_reference
    ON appointments (reference);

-- Lookup by patient phone (for daily limits and history)
CREATE INDEX idx_appointments_phone
    ON appointments (patient_phone, created_at);

-- =============================================================================
-- 6. phone_verifications — the 6-digit codes
-- =============================================================================
CREATE TABLE phone_verifications (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    phone            TEXT        NOT NULL,
    code_hash        TEXT        NOT NULL,
    channel          TEXT        NOT NULL,
    attempts         INT         NOT NULL DEFAULT 0,
    expires_at       TIMESTAMPTZ NOT NULL,
    verified_at      TIMESTAMPTZ,
    token_hash       TEXT        UNIQUE,
    token_expires_at TIMESTAMPTZ,
    ip               INET,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_phone_verif_channel CHECK (
        channel IN ('WHATSAPP', 'SMS', 'CONSOLE')
    )
);

-- For the hourly cleanup goroutine: find expired unverified codes
CREATE INDEX idx_phone_verif_expires
    ON phone_verifications (expires_at)
    WHERE verified_at IS NULL;

-- Rate limiting: count codes sent to a phone in the last hour
CREATE INDEX idx_phone_verif_phone
    ON phone_verifications (phone, created_at);

-- Token lookup for returning patients
CREATE INDEX idx_phone_verif_token
    ON phone_verifications (token_hash)
    WHERE token_hash IS NOT NULL;

COMMIT;
