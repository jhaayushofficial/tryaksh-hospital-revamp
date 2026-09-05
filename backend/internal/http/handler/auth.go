package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/domain/auth"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

type Auth struct {
	q *db.Queries
}

func NewAuth(pool *pgxpool.Pool) *Auth {
	return &Auth{
		q: db.New(pool),
	}
}

func (h *Auth) RequestCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Phone   string `json:"phone"`
		Channel string `json:"channel"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	if req.Phone == "" {
		response.BadRequest(w, "phone is required")
		return
	}

	if req.Channel == "" {
		req.Channel = "SMS" // default
	}

	if req.Channel != "SMS" && req.Channel != "WHATSAPP" && req.Channel != "CONSOLE" {
		response.BadRequest(w, "invalid channel (must be SMS, WHATSAPP, or CONSOLE)")
		return
	}

	// Rate Limiting check should be done here if implemented.

	code, err := auth.GenerateOTP()
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	hash, err := auth.HashOTP(code)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	expiresAt := time.Now().Add(10 * time.Minute)

	_, err = h.q.CreatePhoneVerification(ctx, db.CreatePhoneVerificationParams{
		Phone:     req.Phone,
		CodeHash:  hash,
		Channel:   req.Channel,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// MOCKED SENDING OTP
	slog.Info("MOCK SEND OTP", "phone", req.Phone, "code", code, "channel", req.Channel)

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "OTP requested successfully",
	})
}

func (h *Auth) VerifyCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	if req.Phone == "" || req.Code == "" {
		response.BadRequest(w, "phone and code are required")
		return
	}

	verif, err := h.q.GetLatestUnverifiedByPhone(ctx, req.Phone)
	if err != nil {
		if err.Error() == "no rows in result set" {
			response.BadRequest(w, "no active verification found for this phone")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	if verif.Attempts >= 5 {
		response.BadRequest(w, "too many failed attempts. please request a new code")
		return
	}

	if !auth.CheckOTP(req.Code, verif.CodeHash) {
		err := h.q.IncrementVerificationAttempts(ctx, verif.ID)
		if err != nil {
			slog.Error("failed to increment attempts", "error", err)
		}
		response.BadRequest(w, "invalid code")
		return
	}

	token, err := auth.GenerateToken()
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	tokenHash := auth.HashToken(token)
	tokenExpiresAt := time.Now().Add(24 * 7 * time.Hour) // 7 days

	_, err = h.q.MarkPhoneVerified(ctx, verif.ID, tokenHash, tokenExpiresAt)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}
