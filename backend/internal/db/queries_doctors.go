package db

import (
	"context"

	"github.com/google/uuid"
)

type CreateDoctorParams struct {
	Slug            string  `json:"slug"`
	Name            string  `json:"name"`
	Qualification   *string `json:"qualification,omitempty"`
	Specialization  *string `json:"specialization,omitempty"`
	ExperienceYears *int32  `json:"experience_years,omitempty"`
	Bio             *string `json:"bio,omitempty"`
	PhotoURL        *string `json:"photo_url,omitempty"`
	Active          bool    `json:"active"`
}

const createDoctor = `-- name: CreateDoctor :one
INSERT INTO doctors (
    slug, name, qualification, specialization, experience_years, bio, photo_url, active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id, slug, name, qualification, specialization, experience_years, bio, photo_url, active, created_at, updated_at
`

func (q *Queries) CreateDoctor(ctx context.Context, arg CreateDoctorParams) (Doctor, error) {
	row := q.db().QueryRow(ctx, createDoctor,
		arg.Slug,
		arg.Name,
		arg.Qualification,
		arg.Specialization,
		arg.ExperienceYears,
		arg.Bio,
		arg.PhotoURL,
		arg.Active,
	)
	var i Doctor
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Qualification,
		&i.Specialization,
		&i.ExperienceYears,
		&i.Bio,
		&i.PhotoURL,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getDoctorByID = `-- name: GetDoctorByID :one
SELECT id, slug, name, qualification, specialization, experience_years, bio, photo_url, active, created_at, updated_at FROM doctors WHERE id = $1
`

func (q *Queries) GetDoctorByID(ctx context.Context, id uuid.UUID) (Doctor, error) {
	row := q.db().QueryRow(ctx, getDoctorByID, id)
	var i Doctor
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Qualification,
		&i.Specialization,
		&i.ExperienceYears,
		&i.Bio,
		&i.PhotoURL,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getDoctorBySlug = `-- name: GetDoctorBySlug :one
SELECT id, slug, name, qualification, specialization, experience_years, bio, photo_url, active, created_at, updated_at FROM doctors WHERE slug = $1
`

func (q *Queries) GetDoctorBySlug(ctx context.Context, slug string) (Doctor, error) {
	row := q.db().QueryRow(ctx, getDoctorBySlug, slug)
	var i Doctor
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Qualification,
		&i.Specialization,
		&i.ExperienceYears,
		&i.Bio,
		&i.PhotoURL,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listActiveDoctors = `-- name: ListActiveDoctors :many
SELECT id, slug, name, qualification, specialization, experience_years, bio, photo_url, active, created_at, updated_at FROM doctors WHERE active = true ORDER BY name
`

func (q *Queries) ListActiveDoctors(ctx context.Context) ([]Doctor, error) {
	rows, err := q.db().Query(ctx, listActiveDoctors)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Doctor
	for rows.Next() {
		var i Doctor
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.Name,
			&i.Qualification,
			&i.Specialization,
			&i.ExperienceYears,
			&i.Bio,
			&i.PhotoURL,
			&i.Active,
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

const listAllDoctors = `-- name: ListAllDoctors :many
SELECT id, slug, name, qualification, specialization, experience_years, bio, photo_url, active, created_at, updated_at FROM doctors ORDER BY name
`

func (q *Queries) ListAllDoctors(ctx context.Context) ([]Doctor, error) {
	rows, err := q.db().Query(ctx, listAllDoctors)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Doctor
	for rows.Next() {
		var i Doctor
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.Name,
			&i.Qualification,
			&i.Specialization,
			&i.ExperienceYears,
			&i.Bio,
			&i.PhotoURL,
			&i.Active,
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

type UpdateDoctorParams struct {
	ID              uuid.UUID `json:"id"`
	Slug            *string   `json:"slug,omitempty"`
	Name            *string   `json:"name,omitempty"`
	Qualification   *string   `json:"qualification,omitempty"`
	Specialization  *string   `json:"specialization,omitempty"`
	ExperienceYears *int32    `json:"experience_years,omitempty"`
	Bio             *string   `json:"bio,omitempty"`
	PhotoURL        *string   `json:"photo_url,omitempty"`
	Active          *bool     `json:"active,omitempty"`
}

const updateDoctor = `-- name: UpdateDoctor :one
UPDATE doctors SET
    slug = COALESCE($2, slug),
    name = COALESCE($3, name),
    qualification = COALESCE($4, qualification),
    specialization = COALESCE($5, specialization),
    experience_years = COALESCE($6, experience_years),
    bio = COALESCE($7, bio),
    photo_url = COALESCE($8, photo_url),
    active = COALESCE($9, active),
    updated_at = now()
WHERE id = $1
RETURNING id, slug, name, qualification, specialization, experience_years, bio, photo_url, active, created_at, updated_at
`

func (q *Queries) UpdateDoctor(ctx context.Context, arg UpdateDoctorParams) (Doctor, error) {
	row := q.db().QueryRow(ctx, updateDoctor,
		arg.ID,
		arg.Slug,
		arg.Name,
		arg.Qualification,
		arg.Specialization,
		arg.ExperienceYears,
		arg.Bio,
		arg.PhotoURL,
		arg.Active,
	)
	var i Doctor
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Qualification,
		&i.Specialization,
		&i.ExperienceYears,
		&i.Bio,
		&i.PhotoURL,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
