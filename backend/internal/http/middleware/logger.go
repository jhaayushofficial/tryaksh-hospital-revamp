package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

// healthEndpoints are polled continuously by the platform. They are logged at
// debug level so that steady-state probe traffic does not bury real requests.
var healthEndpoints = map[string]bool{
	"/healthz": true,
	"/readyz":  true,
}

// responseWriterObserver records the status code and body size actually sent,
// which net/http does not otherwise expose to middleware.
type responseWriterObserver struct {
	http.ResponseWriter
	status      int
	bytes       int
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
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// Logger emits one structured record per completed request. The level follows
// the outcome — 5xx is an error, 4xx a warning — so alerting can key off level
// alone. The request ID is added by the context-aware handler.
func Logger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := r.Context()

			ww := &responseWriterObserver{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)

			logger.LogAttrs(ctx, levelFor(ww.status, r.URL.Path), "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.Int("status", ww.status),
				slog.Int("bytes", ww.bytes),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.String("remote_ip", clientIP(r)),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}

func levelFor(status int, path string) slog.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return slog.LevelError
	case status >= http.StatusBadRequest:
		return slog.LevelWarn
	case healthEndpoints[path]:
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}

// clientIP strips the port from RemoteAddr. X-Forwarded-For is deliberately
// ignored: it is client-controlled and this service is not yet fronted by a
// proxy that overwrites it.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
