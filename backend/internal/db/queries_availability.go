package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const getAvailabilityBlocks = `-- name: GetAvailabilityBlocks :many
SELECT id, doctor_id, clinic_id, available_date, start_time, end_time, break_start, break_end, is_open, note, created_at, updated_at FROM availability
WHERE doctor_id = $1 AND clinic_id = $2 AND available_date = $3 AND is_open = true
ORDER BY start_time
`

func (q *Queries) GetAvailabilityBlocks(ctx context.Context, doctorID uuid.UUID, clinicID uuid.UUID, availableDate time.Time) ([]Availability, error) {
	rows, err := q.db().Query(ctx, getAvailabilityBlocks, doctorID, clinicID, availableDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Availability
	for rows.Next() {
		var i Availability
		if err := rows.Scan(
			&i.ID,
			&i.DoctorID,
			&i.ClinicID,
			&i.AvailableDate,
			&i.StartTime,
			&i.EndTime,
			&i.BreakStart,
			&i.BreakEnd,
			&i.IsOpen,
			&i.Note,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const listAvailabilityRange = `-- name: ListAvailabilityRange :many
SELECT id, doctor_id, clinic_id, available_date, start_time, end_time, break_start, break_end, is_open, note, created_at, updated_at FROM availability
WHERE doctor_id = $1 AND clinic_id = $2 
  AND available_date >= $3 AND available_date <= $4
ORDER BY available_date, start_time
`

func (q *Queries) ListAvailabilityRange(ctx context.Context, doctorID uuid.UUID, clinicID uuid.UUID, availableDate time.Time, availableDate2 time.Time) ([]Availability, error) {
	rows, err := q.db().Query(ctx, listAvailabilityRange, doctorID, clinicID, availableDate, availableDate2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Availability
	for rows.Next() {
		var i Availability
		if err := rows.Scan(
			&i.ID,
			&i.DoctorID,
			&i.ClinicID,
			&i.AvailableDate,
			&i.StartTime,
			&i.EndTime,
			&i.BreakStart,
			&i.BreakEnd,
			&i.IsOpen,
			&i.Note,
			&i.CreatedAt,
			&i.UpdatedAt,
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

type CreateAvailabilityParams struct {
	DoctorID      uuid.UUID  `json:"doctor_id"`
	ClinicID      uuid.UUID  `json:"clinic_id"`
	AvailableDate time.Time  `json:"available_date"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       time.Time  `json:"end_time"`
	BreakStart    *time.Time `json:"break_start,omitempty"`
	BreakEnd      *time.Time `json:"break_end,omitempty"`
	IsOpen        bool       `json:"is_open"`
	Note          *string    `json:"note,omitempty"`
}

const createAvailability = `-- name: CreateAvailability :one
INSERT INTO availability (
    doctor_id, clinic_id, available_date, start_time, end_time, break_start, break_end, is_open, note
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING id, doctor_id, clinic_id, available_date, start_time, end_time, break_start, break_end, is_open, note, created_at, updated_at
`

func (q *Queries) CreateAvailability(ctx context.Context, arg CreateAvailabilityParams) (Availability, error) {
	row := q.db().QueryRow(ctx, createAvailability,
		arg.DoctorID,
		arg.ClinicID,
		arg.AvailableDate,
		arg.StartTime,
		arg.EndTime,
		arg.BreakStart,
		arg.BreakEnd,
		arg.IsOpen,
		arg.Note,
	)
	var i Availability
	err := row.Scan(
		&i.ID,
		&i.DoctorID,
		&i.ClinicID,
		&i.AvailableDate,
		&i.StartTime,
		&i.EndTime,
		&i.BreakStart,
		&i.BreakEnd,
		&i.IsOpen,
		&i.Note,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

type UpdateAvailabilityParams struct {
	ID         uuid.UUID  `json:"id"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	BreakStart *time.Time `json:"break_start,omitempty"`
	BreakEnd   *time.Time `json:"break_end,omitempty"`
	IsOpen     *bool      `json:"is_open,omitempty"`
	Note       *string    `json:"note,omitempty"`
}

const updateAvailability = `-- name: UpdateAvailability :one
UPDATE availability SET
    start_time = COALESCE($2, start_time),
    end_time = COALESCE($3, end_time),
    break_start = COALESCE($4, break_start),
    break_end = COALESCE($5, break_end),
    is_open = COALESCE($6, is_open),
    note = COALESCE($7, note),
    updated_at = now()
WHERE id = $1
RETURNING id, doctor_id, clinic_id, available_date, start_time, end_time, break_start, break_end, is_open, note, created_at, updated_at
`

func (q *Queries) UpdateAvailability(ctx context.Context, arg UpdateAvailabilityParams) (Availability, error) {
	row := q.db().QueryRow(ctx, updateAvailability,
		arg.ID,
		arg.StartTime,
		arg.EndTime,
		arg.BreakStart,
		arg.BreakEnd,
		arg.IsOpen,
		arg.Note,
	)
	var i Availability
	err := row.Scan(
		&i.ID,
		&i.DoctorID,
		&i.ClinicID,
		&i.AvailableDate,
		&i.StartTime,
		&i.EndTime,
		&i.BreakStart,
		&i.BreakEnd,
		&i.IsOpen,
		&i.Note,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteAvailability = `-- name: DeleteAvailability :exec
DELETE FROM availability WHERE id = $1
`

func (q *Queries) DeleteAvailability(ctx context.Context, id uuid.UUID) error {
	_, err := q.db().Exec(ctx, deleteAvailability, id)
	return err
}

const deleteAvailabilityByDate = `-- name: DeleteAvailabilityByDate :exec
DELETE FROM availability
WHERE doctor_id = $1 AND clinic_id = $2 AND available_date = $3
`

func (q *Queries) DeleteAvailabilityByDate(ctx context.Context, doctorID uuid.UUID, clinicID uuid.UUID, availableDate time.Time) error {
	_, err := q.db().Exec(ctx, deleteAvailabilityByDate, doctorID, clinicID, availableDate)
	return err
}
