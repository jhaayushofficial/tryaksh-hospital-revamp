package notify

import "context"

type Notifier interface {
	SendBookingConfirmation(ctx context.Context, phone, patientName, date, time string) error
	SendOTP(ctx context.Context, phone, code string) error
}
