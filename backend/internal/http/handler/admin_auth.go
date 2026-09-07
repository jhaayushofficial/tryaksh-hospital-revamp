package handler

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/tryaksh/clinic/backend/internal/config"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

// adminSessionTTL bounds how long a single admin login stays valid.
const adminSessionTTL = 24 * time.Hour

type AdminAuth struct {
	cfg *config.Config
}

func NewAdminAuth(cfg *config.Config) *AdminAuth {
	return &AdminAuth{cfg: cfg}
}

func (h *AdminAuth) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	// Constant-time comparison: a plain != leaks the password length and
	// prefix through response timing.
	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(h.cfg.AdminLoginPassword)) != 1 {
		slog.WarnContext(r.Context(), "admin login failed", slog.String("remote_ip", r.RemoteAddr))
		response.Unauthorized(w, "invalid password")
		return
	}

	expiresAt := time.Now().Add(adminSessionTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "admin",
		"exp":  expiresAt.Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.cfg.SessionSecret))
	if err != nil {
		response.InternalServerError(w, r, fmt.Errorf("signing admin token: %w", err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    tokenString,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   h.cfg.AppEnv == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	slog.InfoContext(r.Context(), "admin login succeeded", slog.Time("expires_at", expiresAt))

	response.JSONCtx(r.Context(), w, http.StatusOK, map[string]string{
		"message": "logged in successfully",
		"token":   tokenString, // Also returned in the body for non-cookie clients.
	})
}

func (h *AdminAuth) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   h.cfg.AppEnv == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}
