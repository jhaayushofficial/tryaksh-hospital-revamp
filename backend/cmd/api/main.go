// Tryaksh Clinic Appointment System — API server entry point.
package main

import (
	"context"
	"errors"
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
	"github.com/tryaksh/clinic/backend/internal/platform/logging"
	"github.com/tryaksh/clinic/backend/internal/platform/notify"
)

const (
	shutdownTimeout = 10 * time.Second
	otpSweepEvery   = 1 * time.Hour
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.Any("error", err))
		os.Exit(1)
	}
}

// run holds the whole lifecycle so that every path returns an error instead of
// calling os.Exit, which would skip the deferred cleanup below.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger := logging.New(os.Stdout, cfg.LogLevelSlog(), cfg.LogFormat)
	slog.SetDefault(logger)

	logger.Info("tryaksh api starting",
		slog.String("env", cfg.AppEnv),
		slog.Int("port", cfg.Port),
		slog.String("base_url", cfg.BaseURL),
		slog.String("database", cfg.RedactedDatabaseURL()),
		slog.String("timezone", cfg.Timezone),
		slog.Int("slot_duration", cfg.SlotDurationMinutes),
		slog.Int("booking_window_days", cfg.BookingWindowDays),
		slog.String("otp_channel", cfg.OTPChannel),
		slog.String("log_level", cfg.LogLevel),
	)

	// Cancelled on the first shutdown signal, which stops background jobs.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}
	logger.Info("database connected")

	cleanup.StartOTPCleanup(ctx, pool, otpSweepEvery)

	notifier, err := buildNotifier(cfg, logger)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: httpInternal.NewRouter(cfg, pool, logger, notifier),
		// Bound the time a stalled client can hold a connection open.
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", slog.Int("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received", slog.Duration("grace_period", shutdownTimeout))
	}

	// Shut down with a fresh context: ctx is already cancelled by the signal,
	// and Shutdown needs a live deadline to drain in-flight requests.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutting down http server: %w", err)
	}

	logger.Info("server exited cleanly")
	return nil
}

// buildNotifier picks the OTP delivery channel. An unrecognised channel is a
// misconfiguration, not a reason to fail to boot — it degrades to console.
func buildNotifier(cfg *config.Config, logger *slog.Logger) (notify.Notifier, error) {
	switch strings.ToLower(cfg.OTPChannel) {
	case "sms":
		notifier, err := notify.NewMSG91Notifier(cfg.MSG91AuthKey, cfg.MSG91SenderID, cfg.MSG91DLTTemplateID)
		if err != nil {
			return nil, fmt.Errorf("initialising MSG91 SMS notifier: %w", err)
		}
		logger.Info("otp channel selected", slog.String("channel", "sms"), slog.String("provider", "msg91"))
		return notifier, nil

	case "console":
		logger.Info("otp channel selected", slog.String("channel", "console"))
		return notify.NewConsoleNotifier(), nil

	default:
		logger.Warn("otp channel not implemented, falling back to console",
			slog.String("channel", cfg.OTPChannel),
		)
		return notify.NewConsoleNotifier(), nil
	}
}
