-- name: ListClinicsByDoctor :many
SELECT c.*, dc.default_hours
FROM clinics c
JOIN doctor_clinics dc ON c.id = dc.clinic_id
WHERE dc.doctor_id = $1
ORDER BY c.name;

-- name: ListDoctorsByClinic :many
SELECT d.*, dc.default_hours
FROM doctors d
JOIN doctor_clinics dc ON d.id = dc.doctor_id
WHERE dc.clinic_id = $1
ORDER BY d.name;

-- name: GetDoctorClinic :one
SELECT * FROM doctor_clinics
WHERE doctor_id = $1 AND clinic_id = $2;

-- name: LinkDoctorClinic :exec
INSERT INTO doctor_clinics (
    doctor_id, clinic_id, default_hours
) VALUES (
    $1, $2, $3
) ON CONFLICT (doctor_id, clinic_id) DO NOTHING;

-- name: UpdateDoctorClinicHours :one
UPDATE doctor_clinics SET
    default_hours = $3
WHERE doctor_id = $1 AND clinic_id = $2
RETURNING *;

-- name: UnlinkDoctorClinic :exec
DELETE FROM doctor_clinics
WHERE doctor_id = $1 AND clinic_id = $2;
