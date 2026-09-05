-- name: CreateAppointment :one
INSERT INTO appointments (
    reference, doctor_id, clinic_id, appointment_date, start_time, end_time,
    patient_name, patient_phone, patient_email, patient_note,
    is_block, status, idempotency_key, actor_doctor_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: GetAppointmentByID :one
SELECT * FROM appointments WHERE id = $1;

-- name: GetAppointmentByReference :one
SELECT * FROM appointments WHERE reference = $1;

-- name: ListAppointmentsByDoctorDate :many
SELECT * FROM appointments
WHERE doctor_id = $1 AND clinic_id = $2 AND appointment_date = $3
ORDER BY start_time;

-- name: ListAppointmentsByPhone :many
SELECT * FROM appointments
WHERE patient_phone = $1
ORDER BY appointment_date DESC, start_time DESC;

-- name: UpdateAppointmentStatus :one
UPDATE appointments SET
    status = $2,
    cancelled_by = $3,
    cancelled_reason = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: RescheduleAppointment :one
UPDATE appointments SET
    rescheduled_from = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CountAppointmentsByPhoneAndDate :one
SELECT count(*) FROM appointments
WHERE patient_phone = $1 AND appointment_date = $2 AND status <> 'CANCELLED' AND is_block = false;

-- name: BookedSlotsForDate :many
SELECT start_time FROM appointments
WHERE doctor_id = $1 AND clinic_id = $2 AND appointment_date = $3 AND status <> 'CANCELLED';

-- name: SearchAppointments :many
SELECT * FROM appointments
WHERE (sqlc.narg(doctor_id)::uuid IS NULL OR doctor_id = sqlc.narg(doctor_id)::uuid)
  AND (sqlc.narg(clinic_id)::uuid IS NULL OR clinic_id = sqlc.narg(clinic_id)::uuid)
  AND (sqlc.narg(start_date)::date IS NULL OR appointment_date >= sqlc.narg(start_date)::date)
  AND (sqlc.narg(end_date)::date IS NULL OR appointment_date <= sqlc.narg(end_date)::date)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
ORDER BY appointment_date DESC, start_time DESC;
