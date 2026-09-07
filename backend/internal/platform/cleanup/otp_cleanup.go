// Package cleanup holds the background jobs that keep transient tables small.
package cleanup

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// otpRetention is how long a verification row is kept after it is created.
// Codes expire far sooner (OTP_TTL_SECONDS); this only sweeps the rows.
const otpRetention = "24 hours"

const deleteExpiredOTPs = `DELETE FROM phone_verifications WHERE created_at < NOW() - INTERVAL '` + otpRetention + `'`

// StartOTPCleanup deletes expired phone verification rows on a ticker until
// ctx is cancelled, which happens when the server shuts down.
func StartOTPCleanup(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		slog.InfoContext(ctx, "otp cleanup started",
			slog.Duration("interval", interval),
			slog.String("retention", otpRetention),
		)

		for {
			select {
			case <-ctx.Done():
				slog.InfoContext(ctx, "otp cleanup stopped")
				return
			case <-ticker.C:
				sweepOTPs(ctx, pool)
			}
		}
	}()
}

func sweepOTPs(ctx context.Context, pool *pgxpool.Pool) {
	start := time.Now()

	tag, err := pool.Exec(ctx, deleteExpiredOTPs)
	if err != nil {
		slog.ErrorContext(ctx, "otp cleanup failed", slog.Any("error", err))
		return
	}

	slog.DebugContext(ctx, "otp cleanup completed",
		slog.Int64("deleted", tag.RowsAffected()),
		slog.Int64("duration_ms", time.Since(start).Milliseconds()),
	)
}
