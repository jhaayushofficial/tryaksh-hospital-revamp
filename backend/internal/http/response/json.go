package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			slog.Error("failed to encode json response", "error", err)
		}
	}
}

// InternalServerError sends a generic 500 error.
func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "internal server error", "error", err)
	JSON(w, http.StatusInternalServerError, JSONError{
		Error: ErrorResponse{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "An unexpected error occurred.",
		},
	})
}

// BadRequest sends a 400 error.
func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, JSONError{
		Error: ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: message,
		},
	})
}

// NotFound sends a 404 error.
func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, JSONError{
		Error: ErrorResponse{
			Code:    "NOT_FOUND",
			Message: message,
		},
	})
}

// Conflict sends a 409 error.
func Conflict(w http.ResponseWriter, message string) {
	JSON(w, http.StatusConflict, JSONError{
		Error: ErrorResponse{
			Code:    "CONFLICT",
			Message: message,
		},
	})
}

// Unauthorized sends a 401 error.
func Unauthorized(w http.ResponseWriter, message string) {
	JSON(w, http.StatusUnauthorized, JSONError{
		Error: ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}
