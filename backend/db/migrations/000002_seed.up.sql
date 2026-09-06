-- Seed data: real doctors, clinics, and links for Tryaksh Hospital, Darbhanga.
-- Uses fixed UUIDs so the seed is idempotent.

BEGIN;

-- ── Doctors ────────────────────────────────────────────────────────────────────

INSERT INTO doctors (id, slug, name, qualification, specialization, experience_years, bio, photo_url)
VALUES
    (
        'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
        'shankar',
        'Dr. Shankar Mishra',
        'MBBS, MD',
        'General Medicine',
        15,
        'Dr. Shankar Mishra is an experienced general physician providing trusted family healthcare in Darbhanga. He specialises in preventive care, chronic disease management, and patient-centred treatment.',
        'https://res.cloudinary.com/w5nagizy/image/upload/v1788728756/Dr.ShankarMishra.png'
    ),
    (
        'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e',
        'mahima',
        'Dr. Mahima Mishra',
        'MBBS, MD',
        'General Medicine',
        10,
        'Dr. Mahima Mishra is a dedicated physician with a focus on holistic patient care. She is known for her compassionate approach and thorough diagnostic practice in Darbhanga.',
        'https://res.cloudinary.com/w5nagizy/image/upload/v1788728744/Dr.MahimaMishra.png'
    );

-- ── Clinics ────────────────────────────────────────────────────────────────────

INSERT INTO clinics (id, slug, name, address, phone, maps_url)
VALUES
    (
        'c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f',
        'bhandar-chowk',
        'Bhandar Chowk',
        'Bhandar Chowk, Darbhanga, Bihar',
        '+919229333922',
        'https://maps.app.goo.gl/5jQuTncJpuTnUsYb7?g_st=iw'
    ),
    (
        'd4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a',
        'kamalpur',
        'Kamalpur',
        'Kamalpur, Darbhanga, Bihar',
        '+919229333922',
        'https://maps.app.goo.gl/eG9q1QAR8tQUToKKA?g_st=iw'
    ),
    (
        'e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b',
        'laxmisagar',
        'Laxmisagar',
        'Laxmisagar, Darbhanga, Bihar',
        '+919229333922',
        'https://maps.app.goo.gl/QCv9a2qbuoNNHff97?g_st=iw'
    );

-- ── Doctor-Clinic links ────────────────────────────────────────────────────────
-- Both doctors visit all 3 clinics.
-- default_hours: a reasonable weekly pattern (editable from the dashboard).

INSERT INTO doctor_clinics (doctor_id, clinic_id, default_hours)
VALUES
    -- Dr. Shankar at all 3 clinics
    (
        'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
        'c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f',
        '{
            "monday":    [{"start": "10:00", "end": "13:00", "break_start": "11:30", "break_end": "11:45"}],
            "tuesday":   [{"start": "10:00", "end": "13:00"}],
            "wednesday": [{"start": "10:00", "end": "13:00"}],
            "thursday":  [{"start": "10:00", "end": "13:00"}],
            "friday":    [{"start": "10:00", "end": "13:00"}],
            "saturday":  [{"start": "10:00", "end": "12:00"}],
            "sunday":    []
        }'::jsonb
    ),
    (
        'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
        'd4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a',
        '{
            "monday":    [{"start": "15:00", "end": "17:00"}],
            "wednesday": [{"start": "15:00", "end": "17:00"}],
            "friday":    [{"start": "15:00", "end": "17:00"}]
        }'::jsonb
    ),
    (
        'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
        'e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b',
        '{
            "tuesday":  [{"start": "15:00", "end": "17:00"}],
            "thursday": [{"start": "15:00", "end": "17:00"}],
            "saturday": [{"start": "14:00", "end": "16:00"}]
        }'::jsonb
    ),

    -- Dr. Mahima at all 3 clinics
    (
        'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e',
        'c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f',
        '{
            "monday":    [{"start": "15:00", "end": "18:00"}],
            "tuesday":   [{"start": "15:00", "end": "18:00"}],
            "wednesday": [{"start": "15:00", "end": "18:00"}],
            "thursday":  [{"start": "15:00", "end": "18:00"}],
            "friday":    [{"start": "15:00", "end": "18:00"}],
            "saturday":  [],
            "sunday":    []
        }'::jsonb
    ),
    (
        'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e',
        'd4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a',
        '{
            "tuesday":  [{"start": "10:00", "end": "13:00"}],
            "thursday": [{"start": "10:00", "end": "13:00"}],
            "saturday": [{"start": "10:00", "end": "12:00"}]
        }'::jsonb
    ),
    (
        'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e',
        'e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b',
        '{
            "monday":   [{"start": "10:00", "end": "13:00"}],
            "wednesday":[{"start": "10:00", "end": "13:00"}],
            "friday":   [{"start": "10:00", "end": "13:00"}]
        }'::jsonb
    );

COMMIT;
