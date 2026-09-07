// Package response centralises JSON encoding and the error envelope so every
// endpoint fails in the same shape.
package response

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tryaksh/clinic/backend/internal/platform/logging"
)

// Error codes returned in the envelope. Clients switch on these, not on prose.
const (
	CodeBadRequest    = "BAD_REQUEST"
	CodeUnauthorized  = "UNAUTHORIZED"
	CodeNotFound      = "NOT_FOUND"
	CodeConflict      = "CONFLICT"
	CodeInternalError = "INTERNAL_SERVER_ERROR"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type JSONError struct {
	Error     ErrorResponse `json:"error"`
	RequestID string        `json:"request_id,omitempty"`
}

// JSON sends a JSON response with the given status code.
//
// Encoding failures are logged without request correlation; use JSONCtx from
// any call site that has a request in hand.
func JSON(w http.ResponseWriter, status int, data any) {
	JSONCtx(context.Background(), w, status, data)
}

// JSONCtx is JSON, with the request context attached so that an encoding
// failure is logged against the request ID that caused it.
func JSONCtx(ctx context.Context, w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// The status line is already on the wire, so the response cannot be
		// converted into an error — record it and move on.
		slog.ErrorContext(ctx, "failed to encode json response",
			slog.Int("status", status),
			slog.Any("error", err),
		)
	}
}

// WriteInternalServerError sends the 500 envelope without logging. Callers that
// have already logged the cause (the panic recoverer) use this; everything else
// should use InternalServerError.
func WriteInternalServerError(w http.ResponseWriter, r *http.Request) {
	JSONCtx(r.Context(), w, http.StatusInternalServerError, JSONError{
		Error: ErrorResponse{
			Code:    CodeInternalError,
			Message: "An unexpected error occurred.",
		},
		RequestID: logging.RequestID(r.Context()),
	})
}

// InternalServerError logs the underlying cause and sends a generic 500. The
// cause is never echoed to the client; the request ID in the body is the only
// handle a user needs to quote for it to be found in the logs.
func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "internal server error",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Any("error", err),
	)
	WriteInternalServerError(w, r)
}

// BadRequest sends a 400 error.
func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, JSONError{
		Error: ErrorResponse{Code: CodeBadRequest, Message: message},
	})
}

// NotFound sends a 404 error.
func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, JSONError{
		Error: ErrorResponse{Code: CodeNotFound, Message: message},
	})
}

// Conflict sends a 409 error.
func Conflict(w http.ResponseWriter, message string) {
	JSON(w, http.StatusConflict, JSONError{
		Error: ErrorResponse{Code: CodeConflict, Message: message},
	})
}

// Unauthorized sends a 401 error.
func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, JSONError{
		Error: ErrorResponse{Code: CodeUnauthorized, Message: message},
	})
}
