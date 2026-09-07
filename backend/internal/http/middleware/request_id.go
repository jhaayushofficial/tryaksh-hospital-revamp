package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"

	"github.com/tryaksh/clinic/backend/internal/platform/logging"
)

// contextKey namespaces the values this package stores on a request context.
type contextKey string

// requestIDHeader is both read (to continue a trace started upstream) and
// written (so a client can quote the ID when reporting a failure).
const requestIDHeader = "X-Request-Id"

// safeRequestID guards the trust boundary: an inbound ID ends up in log
// records, so it is only reused when it is short and alphanumeric.
var safeRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// RequestID assigns every request a correlation ID and puts it on the context
// so that all logging for the request carries it automatically.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(requestIDHeader)
		if !safeRequestID.MatchString(reqID) {
			reqID = newRequestID()
		}

		w.Header().Set(requestIDHeader, reqID)
		next.ServeHTTP(w, r.WithContext(logging.WithRequestID(r.Context(), reqID)))
	})
}

func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is unrecoverable for anything security related,
		// but a correlation ID is not: degrade to an empty ID rather than
		// killing the request.
		return ""
	}
	return hex.EncodeToString(b)
}
