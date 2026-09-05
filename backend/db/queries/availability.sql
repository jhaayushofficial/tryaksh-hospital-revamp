-- name: GetAvailabilityBlocks :many
SELECT * FROM availability
WHERE doctor_id = $1 AND clinic_id = $2 AND available_date = $3 AND is_open = true
ORDER BY start_time;

-- name: ListAvailabilityRange :many
SELECT * FROM availability
WHERE doctor_id = $1 AND clinic_id = $2 
  AND available_date >= $3 AND available_date <= $4
ORDER BY available_date, start_time;

-- name: CreateAvailability :one
INSERT INTO availability (
    doctor_id, clinic_id, available_date, start_time, end_time, break_start, break_end, is_open, note
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: UpdateAvailability :one
UPDATE availability SET
    start_time = COALESCE(sqlc.narg(start_time), start_time),
    end_time = COALESCE(sqlc.narg(end_time), end_time),
    break_start = COALESCE(sqlc.narg(break_start), break_start),
    break_end = COALESCE(sqlc.narg(break_end), break_end),
    is_open = COALESCE(sqlc.narg(is_open), is_open),
    note = COALESCE(sqlc.narg(note), note),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteAvailability :exec
DELETE FROM availability WHERE id = $1;

-- name: DeleteAvailabilityByDate :exec
DELETE FROM availability
WHERE doctor_id = $1 AND clinic_id = $2 AND available_date = $3;
