package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"

	"github.com/tryaksh/clinic/backend/internal/config"
	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/domain/auth"
	"github.com/tryaksh/clinic/backend/internal/http/response"
	"github.com/tryaksh/clinic/backend/internal/platform/logging"
	"github.com/tryaksh/clinic/backend/internal/platform/notify"
)

// validOTPChannels are the channels a client may ask a code to be sent over.
var validOTPChannels = map[string]bool{
	"SMS":      true,
	"WHATSAPP": true,
	"CONSOLE":  true,
}

type Auth struct {
	q        *db.Queries
	notifier notify.Notifier
	cfg      *config.Config

	// The Firebase client fetches and caches Google's signing keys, so it is
	// built once and shared rather than per request.
	fbMu   sync.Mutex
	fbAuth *fbauth.Client
}

func NewAuth(pool *pgxpool.Pool, notifier notify.Notifier, cfg *config.Config) *Auth {
	return &Auth{
		q:        db.New(pool),
		notifier: notifier,
		cfg:      cfg,
	}
}

// otpTTL and verificationTTL read from config so the deployed values are the
// ones that actually apply.
func (h *Auth) otpTTL() time.Duration {
	return time.Duration(h.cfg.OTPTTLSeconds) * time.Second
}

func (h *Auth) verificationTTL() time.Duration {
	return time.Duration(h.cfg.OTPVerificationValidDays) * 24 * time.Hour
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
		req.Channel = "SMS"
	}
	if !validOTPChannels[req.Channel] {
		response.BadRequest(w, "invalid channel (must be SMS, WHATSAPP, or CONSOLE)")
		return
	}

	ctx = logging.With(ctx,
		logging.Phone("phone", req.Phone),
		slog.String("channel", req.Channel),
	)
	r = r.WithContext(ctx)

	code, err := auth.GenerateOTP()
	if err != nil {
		response.InternalServerError(w, r, fmt.Errorf("generating otp: %w", err))
		return
	}

	hash, err := auth.HashOTP(code)
	if err != nil {
		response.InternalServerError(w, r, fmt.Errorf("hashing otp: %w", err))
		return
	}

	if _, err := h.q.CreatePhoneVerification(ctx, db.CreatePhoneVerificationParams{
		Phone:     req.Phone,
		CodeHash:  hash,
		Channel:   req.Channel,
		ExpiresAt: time.Now().Add(h.otpTTL()),
	}); err != nil {
		response.InternalServerError(w, r, fmt.Errorf("storing phone verification: %w", err))
		return
	}

	if err := h.notifier.SendOTP(ctx, req.Phone, code); err != nil {
		response.InternalServerError(w, r, fmt.Errorf("sending otp: %w", err))
		return
	}

	slog.InfoContext(ctx, "otp requested", slog.Duration("ttl", h.otpTTL()))

	response.JSONCtx(ctx, w, http.StatusOK, map[string]string{
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

	ctx = logging.With(ctx, logging.Phone("phone", req.Phone))
	r = r.WithContext(ctx)

	verif, err := h.q.GetLatestUnverifiedByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(ctx, "otp verification rejected", slog.String("reason", "no pending code"))
			response.NotFound(w, "Invalid or expired code")
			return
		}
		response.InternalServerError(w, r, fmt.Errorf("loading phone verification: %w", err))
		return
	}

	if int(verif.Attempts) >= h.cfg.OTPMaxAttempts {
		slog.WarnContext(ctx, "otp verification rejected",
			slog.String("reason", "attempts exhausted"),
			slog.Int("attempts", int(verif.Attempts)),
		)
		response.BadRequest(w, "too many failed attempts. please request a new code")
		return
	}

	if !auth.CheckOTP(req.Code, verif.CodeHash) {
		if err := h.q.IncrementVerificationAttempts(ctx, verif.ID); err != nil {
			slog.ErrorContext(ctx, "failed to increment otp attempts", slog.Any("error", err))
		}
		slog.WarnContext(ctx, "otp verification rejected",
			slog.String("reason", "wrong code"),
			slog.Int("attempts", int(verif.Attempts)+1),
		)
		response.BadRequest(w, "invalid code")
		return
	}

	token, err := h.issueSessionToken(ctx, verif.ID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	slog.InfoContext(ctx, "otp verified")

	response.JSONCtx(ctx, w, http.StatusOK, map[string]string{"token": token})
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

	authClient, err := h.firebaseAuth(ctx)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	token, err := authClient.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		slog.WarnContext(ctx, "firebase token verification failed", slog.Any("error", err))
		response.BadRequest(w, "invalid firebase token")
		return
	}

	// The phone claim is only present when the user signed in with a phone
	// number, which is the only flow this endpoint supports.
	phone, ok := token.Claims["phone_number"].(string)
	if !ok || phone == "" {
		response.BadRequest(w, "firebase token does not contain a phone number")
		return
	}

	ctx = logging.With(ctx, logging.Phone("phone", phone))
	r = r.WithContext(ctx)

	// Firebase has already proven ownership of the number, so the row is
	// recorded as verified rather than carrying a real code.
	verif, err := h.q.CreatePhoneVerification(ctx, db.CreatePhoneVerificationParams{
		Phone:     phone,
		CodeHash:  "FIREBASE_VERIFIED",
		Channel:   "SMS",
		ExpiresAt: time.Now().Add(h.otpTTL()),
	})
	if err != nil {
		response.InternalServerError(w, r, fmt.Errorf("storing phone verification: %w", err))
		return
	}

	sessionToken, err := h.issueSessionToken(ctx, verif.ID)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	slog.InfoContext(ctx, "firebase login succeeded")

	response.JSONCtx(ctx, w, http.StatusOK, map[string]string{
		"token": sessionToken,
		"phone": phone,
	})
}

// issueSessionToken mints an opaque bearer token, stores only its hash against
// the verification row, and returns the clear-text token to the caller.
func (h *Auth) issueSessionToken(ctx context.Context, verificationID uuid.UUID) (string, error) {
	token, err := auth.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("generating session token: %w", err)
	}

	if _, err := h.q.MarkPhoneVerified(ctx, verificationID, auth.HashToken(token), time.Now().Add(h.verificationTTL())); err != nil {
		return "", fmt.Errorf("marking phone verified: %w", err)
	}

	return token, nil
}

// firebaseAuth lazily builds the Firebase Auth client and reuses it. Failures
// are not cached, so a transient startup problem can recover on a later call.
func (h *Auth) firebaseAuth(ctx context.Context) (*fbauth.Client, error) {
	h.fbMu.Lock()
	defer h.fbMu.Unlock()

	if h.fbAuth != nil {
		return h.fbAuth, nil
	}

	if h.cfg.FirebaseProjectID == "" {
		return nil, errors.New("FIREBASE_PROJECT_ID is not configured on the backend")
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: h.cfg.FirebaseProjectID}, option.WithoutAuthentication())
	if err != nil {
		return nil, fmt.Errorf("initialising firebase app: %w", err)
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialising firebase auth client: %w", err)
	}

	h.fbAuth = client
	return client, nil
}
