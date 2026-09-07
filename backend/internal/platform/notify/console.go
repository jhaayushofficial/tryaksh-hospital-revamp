package notify

import (
	"context"
	"log/slog"
)

// ConsoleNotifier is the local/dev channel: instead of sending anything it
// writes the message to the log so the flow can be exercised without a
// provider account.
type ConsoleNotifier struct{}

func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

func (n *ConsoleNotifier) SendBookingConfirmation(ctx context.Context, phone, patientName, date, time string) error {
	slog.InfoContext(ctx, "mock notification: booking confirmation",
		slog.String("channel", "console"),
		slog.String("phone", phone),
		slog.String("patient_name", patientName),
		slog.String("date", date),
		slog.String("time", time),
	)
	return nil
}

// SendOTP writes the verification code in clear text. That is the point of this
// channel, and the reason it must never be selected outside local development.
func (n *ConsoleNotifier) SendOTP(ctx context.Context, phone, code string) error {
	slog.WarnContext(ctx, "mock notification: OTP printed in clear text",
		slog.String("channel", "console"),
		slog.String("phone", phone),
		slog.String("code", code),
	)
	return nil
}
