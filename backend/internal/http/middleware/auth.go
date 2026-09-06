package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/domain/auth"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

const (
	PatientPhoneKey contextKey = "patient_phone"
)

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
				// If no rows, it's invalid/expired
				response.Unauthorized(w, "invalid or expired token")
				return
			}

			// Add verified phone to context
			ctx := context.WithValue(r.Context(), PatientPhoneKey, verif.Phone)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
