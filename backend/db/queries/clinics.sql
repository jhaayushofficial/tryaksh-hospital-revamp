-- name: GetClinicByID :one
SELECT * FROM clinics WHERE id = $1;

-- name: GetClinicBySlug :one
SELECT * FROM clinics WHERE slug = $1;

-- name: ListActiveClinics :many
SELECT * FROM clinics WHERE active = true ORDER BY name;

-- name: ListAllClinics :many
SELECT * FROM clinics ORDER BY name;

-- name: CreateClinic :one
INSERT INTO clinics (
    slug, name, address, directions, phone, maps_url, active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateClinic :one
UPDATE clinics SET
    slug = COALESCE(sqlc.narg(slug), slug),
    name = COALESCE(sqlc.narg(name), name),
    address = COALESCE(sqlc.narg(address), address),
    directions = COALESCE(sqlc.narg(directions), directions),
    phone = COALESCE(sqlc.narg(phone), phone),
    maps_url = COALESCE(sqlc.narg(maps_url), maps_url),
    active = COALESCE(sqlc.narg(active), active),
    updated_at = now()
WHERE id = $1
RETURNING *;
