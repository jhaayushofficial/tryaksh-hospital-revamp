-- Remove seed data in reverse dependency order.

BEGIN;

DELETE FROM doctor_clinics
WHERE doctor_id IN (
    'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
    'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e'
);

DELETE FROM clinics
WHERE id IN (
    'c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f',
    'd4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a',
    'e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b'
);

DELETE FROM doctors
WHERE id IN (
    'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d',
    'b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e'
);

COMMIT;
