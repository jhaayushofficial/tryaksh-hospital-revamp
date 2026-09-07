// Package logging wires the application's structured logger.
//
// Everything logs through log/slog. The handler built here is context-aware:
// attributes attached to a request context (the request ID, and anything a
// handler adds via With) are automatically emitted on every record produced
// during that request, so a log line never has to be manually threaded with
// correlation data.
package logging

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

type ctxKey int

const (
	requestIDKey ctxKey = iota
	attrsKey
)

// Format selects the output encoding of the logger.
const (
	FormatJSON = "json"
	FormatText = "text"
)

// New returns a logger that emits context attributes on every record.
// format is "json" (default, machine readable) or "text" (readable locally).
func New(w io.Writer, level slog.Level, format string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}

	var base slog.Handler
	if strings.EqualFold(format, FormatText) {
		base = slog.NewTextHandler(w, opts)
	} else {
		base = slog.NewJSONHandler(w, opts)
	}

	return slog.New(&contextHandler{Handler: base})
}

// contextHandler decorates every record with the attributes carried on the
// context it was logged with.
type contextHandler struct {
	slog.Handler
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := RequestID(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	if attrs, ok := ctx.Value(attrsKey).([]slog.Attr); ok {
		r.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithGroup(name)}
}

// WithRequestID returns a context whose log records carry the given request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the request ID on the context, or "" when there is none.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// With returns a context whose log records carry the given attributes in
// addition to any already present. Use it to pin request-scoped facts (the
// authenticated phone, the appointment reference) onto every later log line.
func With(ctx context.Context, attrs ...slog.Attr) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	existing, _ := ctx.Value(attrsKey).([]slog.Attr)

	// Copy rather than append in place: the parent context's slice must not be
	// mutated by a child request scope sharing its backing array.
	merged := make([]slog.Attr, 0, len(existing)+len(attrs))
	merged = append(merged, existing...)
	merged = append(merged, attrs...)

	return context.WithValue(ctx, attrsKey, merged)
}
