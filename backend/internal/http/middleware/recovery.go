package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/tryaksh/clinic/backend/internal/http/response"
)

// Recovery turns a panic in a handler into a logged 500 instead of a dropped
// connection. The stack trace is captured here because it is unavailable once
// the deferred function returns.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			// The client hung up mid-write; nothing is wrong with the server
			// and no response can be sent.
			if errors.Is(toError(rec), http.ErrAbortHandler) {
				slog.WarnContext(r.Context(), "handler aborted",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
				)
				return
			}

			slog.ErrorContext(r.Context(), "panic recovered",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Any("panic", rec),
				slog.String("stack", string(debug.Stack())),
			)

			response.WriteInternalServerError(w, r)
		}()

		next.ServeHTTP(w, r)
	})
}

func toError(rec any) error {
	if err, ok := rec.(error); ok {
		return err
	}
	return nil
}
