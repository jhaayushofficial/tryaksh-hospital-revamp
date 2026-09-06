package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/domain/auth"
	"github.com/tryaksh/clinic/backend/internal/http/response"
	"github.com/tryaksh/clinic/backend/internal/platform/notify"
	firebase "firebase.google.com/go/v4"
	"github.com/tryaksh/clinic/backend/internal/config"
	"google.golang.org/api/option"
)

type Auth struct {
	q        *db.Queries
	notifier notify.Notifier
	cfg      *config.Config
}

func NewAuth(pool *pgxpool.Pool, notifier notify.Notifier, cfg *config.Config) *Auth {
	return &Auth{
		q:        db.New(pool),
		notifier: notifier,
		cfg:      cfg,
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

	// Send OTP via the configured channel (SMS, WhatsApp, or console).
	if err := h.notifier.SendOTP(ctx, req.Phone, code); err != nil {
		slog.Error("failed to send OTP", "phone", req.Phone, "channel", req.Channel, "error", err)
		response.InternalServerError(w, r, err)
		return
	}

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
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(w, "Invalid or expired code")
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

func (h *Auth) FirebaseLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		IDToken string `json:"idToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	if req.IDToken == "" {
		response.BadRequest(w, "idToken is required")
		return
	}

	if h.cfg.FirebaseProjectID == "" {
		response.InternalServerError(w, r, errors.New("FIREBASE_PROJECT_ID is not configured on the backend"))
		return
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: h.cfg.FirebaseProjectID}, option.WithoutAuthentication())
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	token, err := authClient.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		slog.Error("firebase token verification failed", "error", err)
		response.BadRequest(w, "invalid firebase token")
		return
	}

	phoneVal, ok := token.Claims["phone_number"]
	if !ok {
		response.BadRequest(w, "firebase token does not contain a phone number")
		return
	}
	phone := phoneVal.(string)

	// Issue custom token
	customToken, err := auth.GenerateToken()
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	tokenHash := auth.HashToken(customToken)
	tokenExpiresAt := time.Now().Add(24 * 7 * time.Hour) // 7 days

	// Save to DB to maintain existing session architecture
	verif, err := h.q.CreatePhoneVerification(ctx, db.CreatePhoneVerificationParams{
		Phone:     phone,
		CodeHash:  "FIREBASE_VERIFIED",
		Channel:   "SMS",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	_, err = h.q.MarkPhoneVerified(ctx, verif.ID, tokenHash, tokenExpiresAt)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"token": customToken,
		"phone": phone,
	})
}
