package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreatePhoneVerificationParams struct {
	Phone     string    `json:"phone"`
	CodeHash  string    `json:"code_hash"`
	Channel   string    `json:"channel"`
	ExpiresAt time.Time `json:"expires_at"`
	Ip        *string   `json:"ip,omitempty"`
}

const createPhoneVerification = `-- name: CreatePhoneVerification :one
INSERT INTO phone_verifications (
    phone, code_hash, channel, expires_at, ip
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, phone, code_hash, channel, attempts, expires_at, verified_at, token_hash, token_expires_at, ip, created_at
`

func (q *Queries) CreatePhoneVerification(ctx context.Context, arg CreatePhoneVerificationParams) (PhoneVerification, error) {
	row := q.db().QueryRow(ctx, createPhoneVerification,
		arg.Phone,
		arg.CodeHash,
		arg.Channel,
		arg.ExpiresAt,
		arg.Ip,
	)
	var i PhoneVerification
	err := row.Scan(
		&i.ID,
		&i.Phone,
		&i.CodeHash,
		&i.Channel,
		&i.Attempts,
		&i.ExpiresAt,
		&i.VerifiedAt,
		&i.TokenHash,
		&i.TokenExpiresAt,
		&i.Ip,
		&i.CreatedAt,
	)
	return i, err
}

const getLatestUnverifiedByPhone = `-- name: GetLatestUnverifiedByPhone :one
SELECT id, phone, code_hash, channel, attempts, expires_at, verified_at, token_hash, token_expires_at, ip, created_at FROM phone_verifications
WHERE phone = $1 AND verified_at IS NULL AND expires_at > now()
ORDER BY created_at DESC
LIMIT 1
`

func (q *Queries) GetLatestUnverifiedByPhone(ctx context.Context, phone string) (PhoneVerification, error) {
	row := q.db().QueryRow(ctx, getLatestUnverifiedByPhone, phone)
	var i PhoneVerification
	err := row.Scan(
		&i.ID,
		&i.Phone,
		&i.CodeHash,
		&i.Channel,
		&i.Attempts,
		&i.ExpiresAt,
		&i.VerifiedAt,
		&i.TokenHash,
		&i.TokenExpiresAt,
		&i.Ip,
		&i.CreatedAt,
	)
	return i, err
}

const incrementVerificationAttempts = `-- name: IncrementVerificationAttempts :exec
UPDATE phone_verifications
SET attempts = attempts + 1
WHERE id = $1
`

func (q *Queries) IncrementVerificationAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := q.db().Exec(ctx, incrementVerificationAttempts, id)
	return err
}

const markPhoneVerified = `-- name: MarkPhoneVerified :one
UPDATE phone_verifications
SET verified_at = now(), token_hash = $2, token_expires_at = $3
WHERE id = $1
RETURNING id, phone, code_hash, channel, attempts, expires_at, verified_at, token_hash, token_expires_at, ip, created_at
`

func (q *Queries) MarkPhoneVerified(ctx context.Context, id uuid.UUID, tokenHash string, tokenExpiresAt time.Time) (PhoneVerification, error) {
	row := q.db().QueryRow(ctx, markPhoneVerified, id, tokenHash, tokenExpiresAt)
	var i PhoneVerification
	err := row.Scan(
		&i.ID,
		&i.Phone,
		&i.CodeHash,
		&i.Channel,
		&i.Attempts,
		&i.ExpiresAt,
		&i.VerifiedAt,
		&i.TokenHash,
		&i.TokenExpiresAt,
		&i.Ip,
		&i.CreatedAt,
	)
	return i, err
}

const getVerificationByToken = `-- name: GetVerificationByToken :one
SELECT id, phone, code_hash, channel, attempts, expires_at, verified_at, token_hash, token_expires_at, ip, created_at FROM phone_verifications
WHERE token_hash = $1 AND token_expires_at > now()
`

func (q *Queries) GetVerificationByToken(ctx context.Context, tokenHash string) (PhoneVerification, error) {
	row := q.db().QueryRow(ctx, getVerificationByToken, tokenHash)
	var i PhoneVerification
	err := row.Scan(
		&i.ID,
		&i.Phone,
		&i.CodeHash,
		&i.Channel,
		&i.Attempts,
		&i.ExpiresAt,
		&i.VerifiedAt,
		&i.TokenHash,
		&i.TokenExpiresAt,
		&i.Ip,
		&i.CreatedAt,
	)
	return i, err
}
