package db

import (
	"context"

	"github.com/google/uuid"
)

type CreateClinicParams struct {
	Slug       string  `json:"slug"`
	Name       string  `json:"name"`
	Address    *string `json:"address,omitempty"`
	Directions *string `json:"directions,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	MapsURL    *string `json:"maps_url,omitempty"`
	Active     bool    `json:"active"`
}

const createClinic = `-- name: CreateClinic :one
INSERT INTO clinics (
    slug, name, address, directions, phone, maps_url, active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, slug, name, address, directions, phone, maps_url, active, created_at, updated_at
`

func (q *Queries) CreateClinic(ctx context.Context, arg CreateClinicParams) (Clinic, error) {
	row := q.db().QueryRow(ctx, createClinic,
		arg.Slug,
		arg.Name,
		arg.Address,
		arg.Directions,
		arg.Phone,
		arg.MapsURL,
		arg.Active,
	)
	var i Clinic
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Address,
		&i.Directions,
		&i.Phone,
		&i.MapsURL,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getClinicByID = `-- name: GetClinicByID :one
SELECT id, slug, name, address, directions, phone, maps_url, active, created_at, updated_at FROM clinics WHERE id = $1
`

func (q *Queries) GetClinicByID(ctx context.Context, id uuid.UUID) (Clinic, error) {
	row := q.db().QueryRow(ctx, getClinicByID, id)
	var i Clinic
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Address,
		&i.Directions,
		&i.Phone,
		&i.MapsURL,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getClinicBySlug = `-- name: GetClinicBySlug :one
SELECT id, slug, name, address, directions, phone, maps_url, active, created_at, updated_at FROM clinics WHERE slug = $1
`

func (q *Queries) GetClinicBySlug(ctx context.Context, slug string) (Clinic, error) {
	row := q.db().QueryRow(ctx, getClinicBySlug, slug)
	var i Clinic
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Address,
		&i.Directions,
		&i.Phone,
		&i.MapsURL,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listActiveClinics = `-- name: ListActiveClinics :many
SELECT id, slug, name, address, directions, phone, maps_url, active, created_at, updated_at FROM clinics WHERE active = true ORDER BY name
`

func (q *Queries) ListActiveClinics(ctx context.Context) ([]Clinic, error) {
	rows, err := q.db().Query(ctx, listActiveClinics)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Clinic
	for rows.Next() {
		var i Clinic
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.Name,
			&i.Address,
			&i.Directions,
			&i.Phone,
			&i.MapsURL,
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

const listAllClinics = `-- name: ListAllClinics :many
SELECT id, slug, name, address, directions, phone, maps_url, active, created_at, updated_at FROM clinics ORDER BY name
`

func (q *Queries) ListAllClinics(ctx context.Context) ([]Clinic, error) {
	rows, err := q.db().Query(ctx, listAllClinics)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Clinic
	for rows.Next() {
		var i Clinic
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.Name,
			&i.Address,
			&i.Directions,
			&i.Phone,
			&i.MapsURL,
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

type UpdateClinicParams struct {
	ID         uuid.UUID `json:"id"`
	Slug       *string   `json:"slug,omitempty"`
	Name       *string   `json:"name,omitempty"`
	Address    *string   `json:"address,omitempty"`
	Directions *string   `json:"directions,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	MapsURL    *string   `json:"maps_url,omitempty"`
	Active     *bool     `json:"active,omitempty"`
}

const updateClinic = `-- name: UpdateClinic :one
UPDATE clinics SET
    slug = COALESCE($2, slug),
    name = COALESCE($3, name),
    address = COALESCE($4, address),
    directions = COALESCE($5, directions),
    phone = COALESCE($6, phone),
    maps_url = COALESCE($7, maps_url),
    active = COALESCE($8, active),
    updated_at = now()
WHERE id = $1
RETURNING id, slug, name, address, directions, phone, maps_url, active, created_at, updated_at
`

func (q *Queries) UpdateClinic(ctx context.Context, arg UpdateClinicParams) (Clinic, error) {
	row := q.db().QueryRow(ctx, updateClinic,
		arg.ID,
		arg.Slug,
		arg.Name,
		arg.Address,
		arg.Directions,
		arg.Phone,
		arg.MapsURL,
		arg.Active,
	)
	var i Clinic
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Name,
		&i.Address,
		&i.Directions,
		&i.Phone,
		&i.MapsURL,
		&i.Active,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
