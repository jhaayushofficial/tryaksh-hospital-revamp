package cleanup

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func StartOTPCleanup(pool *pgxpool.Pool, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			<-ticker.C
			ctx := context.Background()

			// Delete OTPs older than 24 hours
			query := `DELETE FROM auth_otp WHERE created_at < NOW() - INTERVAL '24 hours'`
			tag, err := pool.Exec(ctx, query)
			if err != nil {
				log.Printf("[Cleanup] Failed to delete expired OTPs: %v", err)
			} else {
				if tag.RowsAffected() > 0 {
					log.Printf("[Cleanup] Deleted %d expired OTPs", tag.RowsAffected())
				}
			}
		}
	}()
}
