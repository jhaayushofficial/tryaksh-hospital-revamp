package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tryaksh/clinic/backend/internal/config"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

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

	if req.Password != h.cfg.AdminLoginPassword {
		response.Unauthorized(w, "invalid password")
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "admin",
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.cfg.SessionSecret))
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Set HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    tokenString,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   h.cfg.AppEnv == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "logged in successfully",
		"token":   tokenString, // Also return in body for mobile apps
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
