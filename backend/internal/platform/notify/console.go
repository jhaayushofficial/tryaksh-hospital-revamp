package notify

import (
	"context"
	"log"
)

type ConsoleNotifier struct{}

func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

func (n *ConsoleNotifier) SendBookingConfirmation(ctx context.Context, phone, patientName, date, time string) error {
	log.Printf("[MOCK WHATSAPP] To: %s | Hi %s, your appointment is confirmed for %s at %s. Thanks, Tryaksh Hospital.", phone, patientName, date, time)
	return nil
}

func (n *ConsoleNotifier) SendOTP(ctx context.Context, phone, code string) error {
	log.Printf("[MOCK WHATSAPP] To: %s | Your Tryaksh Hospital verification code is: %s", phone, code)
	return nil
}
