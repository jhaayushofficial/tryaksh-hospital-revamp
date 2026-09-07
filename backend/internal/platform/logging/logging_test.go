package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func decode(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("log line is not valid json: %v (%q)", err, buf.String())
	}
	return record
}

func TestRequestIDIsAddedToEveryRecord(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, slog.LevelInfo, FormatJSON)

	ctx := WithRequestID(context.Background(), "abc123")
	logger.InfoContext(ctx, "booked")

	if got := decode(t, &buf)["request_id"]; got != "abc123" {
		t.Fatalf("request_id = %v, want abc123", got)
	}
}

func TestContextAttrsAreAddedToEveryRecord(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, slog.LevelInfo, FormatJSON)

	ctx := With(context.Background(), slog.String("channel", "sms"))
	ctx = With(ctx, slog.String("reference", "AB-CDEFGH"))
	logger.InfoContext(ctx, "otp requested")

	record := decode(t, &buf)
	if record["channel"] != "sms" {
		t.Errorf("channel = %v, want sms", record["channel"])
	}
	if record["reference"] != "AB-CDEFGH" {
		t.Errorf("reference = %v, want AB-CDEFGH", record["reference"])
	}
}

// A child scope must not leak its attributes back into the parent context,
// which would attach one request's data to another's logs.
func TestWithDoesNotMutateParentContext(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, slog.LevelInfo, FormatJSON)

	parent := With(context.Background(), slog.String("shared", "yes"))
	_ = With(parent, slog.String("child_only", "leaked"))

	logger.InfoContext(parent, "parent scope")

	if _, present := decode(t, &buf)["child_only"]; present {
		t.Fatal("child attribute leaked into the parent context")
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, slog.LevelWarn, FormatJSON)

	logger.InfoContext(context.Background(), "should be dropped")

	if buf.Len() != 0 {
		t.Fatalf("info record emitted at warn level: %q", buf.String())
	}
}

func TestMaskPhone(t *testing.T) {
	cases := map[string]string{
		"+919229333922": "*********3922",
		"9229333922":    "******3922",
		"3922":          "****",
		"12":            "****",
		"":              "****",
	}
	for in, want := range cases {
		if got := MaskPhone(in); got != want {
			t.Errorf("MaskPhone(%q) = %q, want %q", in, got, want)
		}
	}
}
