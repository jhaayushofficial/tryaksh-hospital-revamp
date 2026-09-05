// Package config reads and validates the application configuration from
// environment variables (loaded from .env via godotenv).
package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds every configurable value in the system.
// Values are read from environment variables using struct tags.
type Config struct {
	// ── App ─────────────────────────────────────────────────────────────────
	AppEnv             string `env:"APP_ENV"              envDefault:"local"`
	Port               int    `env:"PORT"                 envDefault:"8080"`
	BaseURL            string `env:"BASE_URL,required"`
	DatabaseURL        string `env:"DATABASE_URL,required"`
	SessionSecret      string `env:"SESSION_SECRET,required"`
	Timezone           string `env:"TIMEZONE"             envDefault:"Asia/Kolkata"`
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`

	// ── Booking rules ───────────────────────────────────────────────────────
	SlotDurationMinutes           int `env:"SLOT_DURATION_MINUTES"            envDefault:"15"`
	BookingWindowDays             int `env:"BOOKING_WINDOW_DAYS"              envDefault:"30"`
	MinLeadTimeMinutes            int `env:"MIN_LEAD_TIME_MINUTES"            envDefault:"60"`
	CancellationWindowHours       int `env:"CANCELLATION_WINDOW_HOURS"        envDefault:"6"`
	MaxBookingsPerPhonePerDay     int `env:"MAX_BOOKINGS_PER_PHONE_PER_DAY"   envDefault:"2"`
	RateLimitBookingsPerHourPerIP int `env:"RATE_LIMIT_BOOKINGS_PER_HOUR_PER_IP" envDefault:"5"`
	RateLimitLookupsPerMinPerIP   int `env:"RATE_LIMIT_LOOKUPS_PER_MIN_PER_IP"   envDefault:"5"`

	// ── OTP ─────────────────────────────────────────────────────────────────
	OTPChannel               string `env:"OTP_CHANNEL"                envDefault:"console"`
	OTPFallbackChannel       string `env:"OTP_FALLBACK_CHANNEL"`
	OTPLength                int    `env:"OTP_LENGTH"                 envDefault:"6"`
	OTPTTLSeconds            int    `env:"OTP_TTL_SECONDS"            envDefault:"300"`
	OTPMaxAttempts           int    `env:"OTP_MAX_ATTEMPTS"           envDefault:"5"`
	OTPResendCooldownSeconds int    `env:"OTP_RESEND_COOLDOWN_SECONDS" envDefault:"30"`
	OTPMaxPerPhonePerHour    int    `env:"OTP_MAX_PER_PHONE_PER_HOUR" envDefault:"3"`
	OTPMaxPerIPPerHour       int    `env:"OTP_MAX_PER_IP_PER_HOUR"    envDefault:"10"`
	OTPVerificationValidDays int    `env:"OTP_VERIFICATION_VALID_DAYS" envDefault:"30"`
	OTPPepper                string `env:"OTP_PEPPER,required"`
	OTPRequiredForCancel     bool   `env:"OTP_REQUIRED_FOR_CANCEL"    envDefault:"true"`

	// ── WhatsApp ────────────────────────────────────────────────────────────
	WhatsAppPhoneNumberID   string `env:"WHATSAPP_PHONE_NUMBER_ID"`
	WhatsAppToken           string `env:"WHATSAPP_TOKEN"`
	WhatsAppOTPTemplateName string `env:"WHATSAPP_OTP_TEMPLATE_NAME"`

	// ── SMS (optional) ──────────────────────────────────────────────────────
	MSG91AuthKey       string `env:"MSG91_AUTH_KEY"`
	MSG91SenderID      string `env:"MSG91_SENDER_ID"`
	MSG91DLTTemplateID string `env:"MSG91_DLT_TEMPLATE_ID"`

	// ── Clinic display ──────────────────────────────────────────────────────
	ClinicName             string `env:"CLINIC_NAME"              envDefault:"Tryaksh Hospital and Diagnostics"`
	ClinicTagline          string `env:"CLINIC_TAGLINE"           envDefault:"Trusted family healthcare in Darbhanga"`
	ClinicPhone            string `env:"CLINIC_PHONE"`
	ClinicWhatsApp         string `env:"CLINIC_WHATSAPP"`
	ConsultationFeeDisplay string `env:"CONSULTATION_FEE_DISPLAY" envDefault:"Payable at the clinic"`

	// ── Admin ───────────────────────────────────────────────────────────────
	AdminLoginPassword string `env:"ADMIN_LOGIN_PASSWORD,required"`

	// ── Notifications ───────────────────────────────────────────────────────
	NotifyWhatsAppEnabled bool   `env:"NOTIFY_WHATSAPP_ENABLED" envDefault:"false"`
	NotifyEmailEnabled    bool   `env:"NOTIFY_EMAIL_ENABLED"    envDefault:"false"`
	ResendAPIKey          string `env:"RESEND_API_KEY"`
	NotifyFromEmail       string `env:"NOTIFY_FROM_EMAIL"`
	ReminderHoursBefore   string `env:"REMINDER_HOURS_BEFORE"   envDefault:"24"`

	// ── Optional ────────────────────────────────────────────────────────────
	SentryDSN string `env:"SENTRY_DSN"`
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
}

// Load reads the .env file (if present), parses environment variables into
// a Config struct, and validates critical fields.
func Load() (*Config, error) {
	// Load .env — ignore error if file doesn't exist (env vars may be set directly).
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// In production, secrets must be strong.
	if cfg.AppEnv == "production" {
		if len(cfg.SessionSecret) < 32 {
			return nil, fmt.Errorf("SESSION_SECRET must be at least 32 bytes in production")
		}
		if len(cfg.OTPPepper) < 32 {
			return nil, fmt.Errorf("OTP_PEPPER must be at least 32 bytes in production")
		}
	}

	return cfg, nil
}

// ParsedCORSOrigins splits the comma-separated CORS_ALLOWED_ORIGINS
// into a trimmed slice of strings.
func (c *Config) ParsedCORSOrigins() []string {
	parts := strings.Split(c.CORSAllowedOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}

// LogLevelSlog converts the LOG_LEVEL string to a slog.Level.
func (c *Config) LogLevelSlog() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// RedactedDatabaseURL returns the DATABASE_URL with the password replaced
// by "***" for safe logging.
func (c *Config) RedactedDatabaseURL() string {
	// Simple approach: find :password@ pattern
	url := c.DatabaseURL
	atIdx := strings.LastIndex(url, "@")
	if atIdx == -1 {
		return url
	}
	colonIdx := strings.Index(url, "://")
	if colonIdx == -1 {
		return "***"
	}
	prefix := url[:colonIdx+3]
	rest := url[colonIdx+3:]

	// rest is "user:pass@host..."
	atInRest := strings.Index(rest, "@")
	if atInRest == -1 {
		return url
	}
	userPass := rest[:atInRest]
	after := rest[atInRest:]

	colonInUP := strings.Index(userPass, ":")
	if colonInUP == -1 {
		return url
	}
	user := userPass[:colonInUP]
	return prefix + user + ":***" + after
}
