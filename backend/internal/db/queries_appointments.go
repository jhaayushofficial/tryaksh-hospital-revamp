package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)



const getBookedTimes = `-- name: GetBookedTimes :many
SELECT start_time FROM appointments
WHERE doctor_id = $1 AND clinic_id = $2 AND appointment_date = $3 AND status <> 'CANCELLED'
`

// GetBookedTimes returns the start_time of all active appointments on a given date.
func (q *Queries) GetBookedTimes(ctx context.Context, doctorID uuid.UUID, clinicID uuid.UUID, appointmentDate time.Time) ([]time.Time, error) {
	rows, err := q.db().Query(ctx, getBookedTimes, doctorID, clinicID, appointmentDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []time.Time
	for rows.Next() {
		var start_time time.Time
		if err := rows.Scan(&start_time); err != nil {
			return nil, err
		}
		items = append(items, start_time)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type CreateAppointmentParams struct {
	Reference       string     `json:"reference"`
	DoctorID        uuid.UUID  `json:"doctor_id"`
	ClinicID        uuid.UUID  `json:"clinic_id"`
	AppointmentDate time.Time  `json:"appointment_date"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         time.Time  `json:"end_time"`
	PatientName     *string    `json:"patient_name,omitempty"`
	PatientPhone    *string    `json:"patient_phone,omitempty"`
	PatientEmail    *string    `json:"patient_email,omitempty"`
	PatientNote     *string    `json:"patient_note,omitempty"`
	IsBlock         bool       `json:"is_block"`
	Status          string     `json:"status"`
	IdempotencyKey  *string    `json:"idempotency_key,omitempty"`
}

const createAppointment = `-- name: CreateAppointment :one
INSERT INTO appointments (
    reference, doctor_id, clinic_id, appointment_date, start_time, end_time,
    patient_name, patient_phone, patient_email, patient_note, is_block, status, idempotency_key
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING id, reference, doctor_id, clinic_id, appointment_date, start_time, end_time, patient_name, patient_phone, patient_email, patient_note, is_block, status, cancelled_by, cancelled_reason, rescheduled_from, idempotency_key, actor_doctor_id, created_at, updated_at
`

func (q *Queries) CreateAppointment(ctx context.Context, arg CreateAppointmentParams) (Appointment, error) {
	row := q.db().QueryRow(ctx, createAppointment,
		arg.Reference,
		arg.DoctorID,
		arg.ClinicID,
		arg.AppointmentDate,
		arg.StartTime,
		arg.EndTime,
		arg.PatientName,
		arg.PatientPhone,
		arg.PatientEmail,
		arg.PatientNote,
		arg.IsBlock,
		arg.Status,
		arg.IdempotencyKey,
	)
	var i Appointment
	err := row.Scan(
		&i.ID,
		&i.Reference,
		&i.DoctorID,
		&i.ClinicID,
		&i.AppointmentDate,
		&i.StartTime,
		&i.EndTime,
		&i.PatientName,
		&i.PatientPhone,
		&i.PatientEmail,
		&i.PatientNote,
		&i.IsBlock,
		&i.Status,
		&i.CancelledBy,
		&i.CancelledReason,
		&i.RescheduledFrom,
		&i.IdempotencyKey,
		&i.ActorDoctorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAppointmentByReference = `-- name: GetAppointmentByReference :one
SELECT id, reference, doctor_id, clinic_id, appointment_date, start_time, end_time, patient_name, patient_phone, patient_email, patient_note, is_block, status, cancelled_by, cancelled_reason, rescheduled_from, idempotency_key, actor_doctor_id, created_at, updated_at FROM appointments
WHERE reference = $1
`

func (q *Queries) GetAppointmentByReference(ctx context.Context, reference string) (Appointment, error) {
	row := q.db().QueryRow(ctx, getAppointmentByReference, reference)
	var i Appointment
	err := row.Scan(
		&i.ID,
		&i.Reference,
		&i.DoctorID,
		&i.ClinicID,
		&i.AppointmentDate,
		&i.StartTime,
		&i.EndTime,
		&i.PatientName,
		&i.PatientPhone,
		&i.PatientEmail,
		&i.PatientNote,
		&i.IsBlock,
		&i.Status,
		&i.CancelledBy,
		&i.CancelledReason,
		&i.RescheduledFrom,
		&i.IdempotencyKey,
		&i.ActorDoctorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateAppointmentStatus = `-- name: UpdateAppointmentStatus :one
UPDATE appointments
SET status = $2, cancelled_by = $3, cancelled_reason = $4, updated_at = now()
WHERE id = $1
RETURNING id, reference, doctor_id, clinic_id, appointment_date, start_time, end_time, patient_name, patient_phone, patient_email, patient_note, is_block, status, cancelled_by, cancelled_reason, rescheduled_from, idempotency_key, actor_doctor_id, created_at, updated_at
`

func (q *Queries) UpdateAppointmentStatus(ctx context.Context, id uuid.UUID, status string, cancelledBy *string, cancelledReason *string) (Appointment, error) {
	row := q.db().QueryRow(ctx, updateAppointmentStatus, id, status, cancelledBy, cancelledReason)
	var i Appointment
	err := row.Scan(
		&i.ID,
		&i.Reference,
		&i.DoctorID,
		&i.ClinicID,
		&i.AppointmentDate,
		&i.StartTime,
		&i.EndTime,
		&i.PatientName,
		&i.PatientPhone,
		&i.PatientEmail,
		&i.PatientNote,
		&i.IsBlock,
		&i.Status,
		&i.CancelledBy,
		&i.CancelledReason,
		&i.RescheduledFrom,
		&i.IdempotencyKey,
		&i.ActorDoctorID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
