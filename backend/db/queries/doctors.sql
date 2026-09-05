-- name: GetDoctorByID :one
SELECT * FROM doctors WHERE id = $1;

-- name: GetDoctorBySlug :one
SELECT * FROM doctors WHERE slug = $1;

-- name: ListActiveDoctors :many
SELECT * FROM doctors WHERE active = true ORDER BY name;

-- name: ListAllDoctors :many
SELECT * FROM doctors ORDER BY name;

-- name: CreateDoctor :one
INSERT INTO doctors (
    slug, name, qualification, specialization, experience_years, bio, photo_url, active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateDoctor :one
UPDATE doctors SET
    slug = COALESCE(sqlc.narg(slug), slug),
    name = COALESCE(sqlc.narg(name), name),
    qualification = COALESCE(sqlc.narg(qualification), qualification),
    specialization = COALESCE(sqlc.narg(specialization), specialization),
    experience_years = COALESCE(sqlc.narg(experience_years), experience_years),
    bio = COALESCE(sqlc.narg(bio), bio),
    photo_url = COALESCE(sqlc.narg(photo_url), photo_url),
    active = COALESCE(sqlc.narg(active), active),
    updated_at = now()
WHERE id = $1
RETURNING *;
