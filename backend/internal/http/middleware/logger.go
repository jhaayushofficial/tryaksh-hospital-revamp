package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriterObserver struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *responseWriterObserver) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *responseWriterObserver) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func Logger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			reqID := GetReqID(r.Context())
			ctx := r.Context()

			ww := &responseWriterObserver{ResponseWriter: w, status: http.StatusOK}
			
			next.ServeHTTP(ww, r.WithContext(ctx))

			logger.InfoContext(ctx, "request completed",
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
