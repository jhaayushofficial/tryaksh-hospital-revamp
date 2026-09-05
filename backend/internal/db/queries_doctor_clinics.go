package db

import (
	"context"

	"github.com/google/uuid"
)

const getDoctorClinic = `-- name: GetDoctorClinic :one
SELECT doctor_id, clinic_id, default_hours, created_at FROM doctor_clinics
WHERE doctor_id = $1 AND clinic_id = $2
`

func (q *Queries) GetDoctorClinic(ctx context.Context, doctorID uuid.UUID, clinicID uuid.UUID) (DoctorClinic, error) {
	row := q.db().QueryRow(ctx, getDoctorClinic, doctorID, clinicID)
	var i DoctorClinic
	err := row.Scan(
		&i.DoctorID,
		&i.ClinicID,
		&i.DefaultHours,
		&i.CreatedAt,
	)
	return i, err
}

type LinkDoctorClinicParams struct {
	DoctorID     uuid.UUID `json:"doctor_id"`
	ClinicID     uuid.UUID `json:"clinic_id"`
	DefaultHours []byte    `json:"default_hours,omitempty"`
}

const linkDoctorClinic = `-- name: LinkDoctorClinic :exec
INSERT INTO doctor_clinics (
    doctor_id, clinic_id, default_hours
) VALUES (
    $1, $2, $3
) ON CONFLICT (doctor_id, clinic_id) DO NOTHING
`

func (q *Queries) LinkDoctorClinic(ctx context.Context, arg LinkDoctorClinicParams) error {
	_, err := q.db().Exec(ctx, linkDoctorClinic, arg.DoctorID, arg.ClinicID, arg.DefaultHours)
	return err
}

type ListClinicsByDoctorRow struct {
	Clinic       Clinic `json:"clinic"`
	DefaultHours []byte `json:"default_hours,omitempty"`
}

const listClinicsByDoctor = `-- name: ListClinicsByDoctor :many
SELECT c.id, c.slug, c.name, c.address, c.directions, c.phone, c.maps_url, c.active, c.created_at, c.updated_at, dc.default_hours
FROM clinics c
JOIN doctor_clinics dc ON c.id = dc.clinic_id
WHERE dc.doctor_id = $1
ORDER BY c.name
`

func (q *Queries) ListClinicsByDoctor(ctx context.Context, doctorID uuid.UUID) ([]ListClinicsByDoctorRow, error) {
	rows, err := q.db().Query(ctx, listClinicsByDoctor, doctorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListClinicsByDoctorRow
	for rows.Next() {
		var i ListClinicsByDoctorRow
		if err := rows.Scan(
			&i.Clinic.ID,
			&i.Clinic.Slug,
			&i.Clinic.Name,
			&i.Clinic.Address,
			&i.Clinic.Directions,
			&i.Clinic.Phone,
			&i.Clinic.MapsURL,
			&i.Clinic.Active,
			&i.Clinic.CreatedAt,
			&i.Clinic.UpdatedAt,
			&i.DefaultHours,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type ListDoctorsByClinicRow struct {
	Doctor       Doctor `json:"doctor"`
	DefaultHours []byte `json:"default_hours,omitempty"`
}

const listDoctorsByClinic = `-- name: ListDoctorsByClinic :many
SELECT d.id, d.slug, d.name, d.qualification, d.specialization, d.experience_years, d.bio, d.photo_url, d.active, d.created_at, d.updated_at, dc.default_hours
FROM doctors d
JOIN doctor_clinics dc ON d.id = dc.doctor_id
WHERE dc.clinic_id = $1
ORDER BY d.name
`

func (q *Queries) ListDoctorsByClinic(ctx context.Context, clinicID uuid.UUID) ([]ListDoctorsByClinicRow, error) {
	rows, err := q.db().Query(ctx, listDoctorsByClinic, clinicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListDoctorsByClinicRow
	for rows.Next() {
		var i ListDoctorsByClinicRow
		if err := rows.Scan(
			&i.Doctor.ID,
			&i.Doctor.Slug,
			&i.Doctor.Name,
			&i.Doctor.Qualification,
			&i.Doctor.Specialization,
			&i.Doctor.ExperienceYears,
			&i.Doctor.Bio,
			&i.Doctor.PhotoURL,
			&i.Doctor.Active,
			&i.Doctor.CreatedAt,
			&i.Doctor.UpdatedAt,
			&i.DefaultHours,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const unlinkDoctorClinic = `-- name: UnlinkDoctorClinic :exec
DELETE FROM doctor_clinics
WHERE doctor_id = $1 AND clinic_id = $2
`

func (q *Queries) UnlinkDoctorClinic(ctx context.Context, doctorID uuid.UUID, clinicID uuid.UUID) error {
	_, err := q.db().Exec(ctx, unlinkDoctorClinic, doctorID, clinicID)
	return err
}

type UpdateDoctorClinicHoursParams struct {
	DoctorID     uuid.UUID `json:"doctor_id"`
	ClinicID     uuid.UUID `json:"clinic_id"`
	DefaultHours []byte    `json:"default_hours,omitempty"`
}

const updateDoctorClinicHours = `-- name: UpdateDoctorClinicHours :one
UPDATE doctor_clinics SET
    default_hours = $3
WHERE doctor_id = $1 AND clinic_id = $2
RETURNING doctor_id, clinic_id, default_hours, created_at
`

func (q *Queries) UpdateDoctorClinicHours(ctx context.Context, arg UpdateDoctorClinicHoursParams) (DoctorClinic, error) {
	row := q.db().QueryRow(ctx, updateDoctorClinicHours, arg.DoctorID, arg.ClinicID, arg.DefaultHours)
	var i DoctorClinic
	err := row.Scan(
		&i.DoctorID,
		&i.ClinicID,
		&i.DefaultHours,
		&i.CreatedAt,
	)
	return i, err
}
