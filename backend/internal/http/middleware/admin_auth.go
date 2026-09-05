package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tryaksh/clinic/backend/internal/config"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

func RequireAdminAuth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			// 1. Try to get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenString = parts[1]
				}
			}

			// 2. If no header, fallback to Cookie
			if tokenString == "" {
				cookie, err := r.Cookie("admin_token")
				if err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				response.Unauthorized(w, "missing admin authorization")
				return
			}

			// Parse and validate JWT
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return []byte(cfg.SessionSecret), nil
			})

			if err != nil || !token.Valid {
				response.Unauthorized(w, "invalid or expired admin token")
				return
			}

			// Token is valid, proceed
			next.ServeHTTP(w, r)
		})
	}
}
