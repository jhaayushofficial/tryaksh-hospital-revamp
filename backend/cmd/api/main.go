// Tryaksh Clinic Appointment System — API server entry point.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/config"
	httpInternal "github.com/tryaksh/clinic/backend/internal/http"
	"github.com/tryaksh/clinic/backend/internal/platform/cleanup"
	"github.com/tryaksh/clinic/backend/internal/platform/notify"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevelSlog(),
	}))
	slog.SetDefault(logger)

	logger.Info("tryaksh api starting",
		"env", cfg.AppEnv,
		"port", cfg.Port,
		"base_url", cfg.BaseURL,
		"database", cfg.RedactedDatabaseURL(),
		"timezone", cfg.Timezone,
		"slot_duration", cfg.SlotDurationMinutes,
		"booking_window_days", cfg.BookingWindowDays,
		"otp_channel", cfg.OTPChannel,
		"clinic", cfg.ClinicName,
	)

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("database ping failed", "error", err)
		os.Exit(1)
	}

	// Start background cleanup routines
	cleanup.StartOTPCleanup(pool, 1*time.Hour)

	// Build the notifier based on the configured OTP channel.
	var notifier notify.Notifier
	switch strings.ToLower(cfg.OTPChannel) {
	case "sms":
		notifier, err = notify.NewMSG91Notifier(cfg.MSG91AuthKey, cfg.MSG91SenderID, cfg.MSG91DLTTemplateID)
		if err != nil {
			logger.Error("failed to initialise MSG91 SMS notifier", "error", err)
			os.Exit(1)
		}
		logger.Info("OTP channel: SMS via MSG91")
	case "console":
		notifier = notify.NewConsoleNotifier()
		logger.Info("OTP channel: console (mock — codes printed to stdout)")
	default:
		// Fallback to console for any unrecognised value (e.g. "whatsapp" not yet implemented).
		logger.Warn("OTP channel not implemented, falling back to console", "channel", cfg.OTPChannel)
		notifier = notify.NewConsoleNotifier()
	}

	router := httpInternal.NewRouter(cfg, pool, logger, notifier)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		logger.Info("starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctxShutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}

	logger.Info("server exited cleanly")
}
