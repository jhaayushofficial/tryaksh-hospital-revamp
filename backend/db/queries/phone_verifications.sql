-- name: CreatePhoneVerification :one
INSERT INTO phone_verifications (
    phone, code_hash, channel, expires_at, ip
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetPhoneVerificationByPhone :one
SELECT * FROM phone_verifications
WHERE phone = $1 AND verified_at IS NULL AND expires_at > now()
ORDER BY created_at DESC LIMIT 1;

-- name: GetPhoneVerificationByID :one
SELECT * FROM phone_verifications WHERE id = $1;

-- name: IncrementPhoneVerificationAttempts :one
UPDATE phone_verifications SET
    attempts = attempts + 1
WHERE id = $1
RETURNING *;

-- name: MarkPhoneVerified :one
UPDATE phone_verifications SET
    verified_at = now(),
    token_hash = $2,
    token_expires_at = $3
WHERE id = $1
RETURNING *;

-- name: GetPhoneVerificationByToken :one
SELECT * FROM phone_verifications
WHERE token_hash = $1 AND token_expires_at > now();

-- name: CleanupExpiredVerifications :exec
DELETE FROM phone_verifications
WHERE verified_at IS NULL AND expires_at < now() - interval '1 day';

-- name: CleanupExpiredTokens :exec
DELETE FROM phone_verifications
WHERE verified_at IS NOT NULL AND token_expires_at IS NOT NULL AND token_expires_at < now();

-- name: CleanupOldVerified :exec
DELETE FROM phone_verifications
WHERE verified_at IS NOT NULL AND token_hash IS NULL AND created_at < now() - interval '7 days';

-- name: CountVerificationsByPhoneHour :one
SELECT count(*) FROM phone_verifications
WHERE phone = $1 AND created_at > now() - interval '1 hour';

-- name: CountVerificationsByIPHour :one
SELECT count(*) FROM phone_verifications
WHERE ip = $1 AND created_at > now() - interval '1 hour';
