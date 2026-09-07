package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/domain/auth"
	"github.com/tryaksh/clinic/backend/internal/http/response"
	"github.com/tryaksh/clinic/backend/internal/platform/logging"
)

const (
	PatientPhoneKey contextKey = "patient_phone"
)

// PatientPhone returns the phone number proven by the bearer token, and false
// when the request did not pass through RequirePatientAuth.
func PatientPhone(ctx context.Context) (string, bool) {
	phone, ok := ctx.Value(PatientPhoneKey).(string)
	return phone, ok && phone != ""
}

func RequirePatientAuth(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	queries := db.New(pool)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.Unauthorized(w, "invalid authorization format")
				return
			}

			token := parts[1]
			tokenHash := auth.HashToken(token)

			verif, err := queries.GetVerificationByToken(r.Context(), tokenHash)
			if err != nil {
				// Either no such token or it has expired — both are a 401, and
				// the distinction is deliberately not leaked to the caller.
				if !errors.Is(err, pgx.ErrNoRows) {
					slog.ErrorContext(r.Context(), "failed to look up session token", slog.Any("error", err))
				}
				response.Unauthorized(w, "invalid or expired token")
				return
			}

			// Carry the verified phone for handlers, and a masked copy for logs.
			ctx := context.WithValue(r.Context(), PatientPhoneKey, verif.Phone)
			ctx = logging.With(ctx, logging.Phone("patient_phone", verif.Phone))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
