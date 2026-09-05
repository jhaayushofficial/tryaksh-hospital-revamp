-- Drop all tables in reverse dependency order.

BEGIN;

DROP TABLE IF EXISTS phone_verifications;
DROP TABLE IF EXISTS appointments;
DROP TABLE IF EXISTS availability;
DROP TABLE IF EXISTS doctor_clinics;
DROP TABLE IF EXISTS clinics;
DROP TABLE IF EXISTS doctors;

COMMIT;
