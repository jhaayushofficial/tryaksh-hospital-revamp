package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// MSG91Notifier sends OTP codes via the MSG91 SMS API.
// Requires a DLT-registered template and an MSG91 auth key.
type MSG91Notifier struct {
	authKey       string
	senderID      string
	dltTemplateID string
	httpClient    *http.Client
}

// NewMSG91Notifier creates a new MSG91Notifier.
// All three parameters are required and must be non-empty.
func NewMSG91Notifier(authKey, senderID, dltTemplateID string) (*MSG91Notifier, error) {
	if authKey == "" {
		return nil, fmt.Errorf("MSG91_AUTH_KEY is required for SMS channel")
	}
	if senderID == "" {
		return nil, fmt.Errorf("MSG91_SENDER_ID is required for SMS channel")
	}
	if dltTemplateID == "" {
		return nil, fmt.Errorf("MSG91_DLT_TEMPLATE_ID is required for SMS channel")
	}

	return &MSG91Notifier{
		authKey:       authKey,
		senderID:      senderID,
		dltTemplateID: dltTemplateID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// msg91SendRequest is the JSON payload for the MSG91 Send SMS API v2.
type msg91SendRequest struct {
	Flow    string            `json:"flow_id"`
	Sender  string            `json:"sender"`
	Mobiles string            `json:"mobiles"`
	OTP     string            `json:"otp,omitempty"`
	Vars    map[string]string `json:"VAR1,omitempty"`
}

// msg91SendOTPRequest is the JSON payload for MSG91's OTP API.
type msg91SendOTPRequest struct {
	TemplateID string `json:"template_id"`
	Mobile     string `json:"mobile"`
	AuthKey    string `json:"authkey"`
	OTP        string `json:"otp"`
}

// SendOTP sends the OTP code via MSG91 SMS to the given phone number.
func (n *MSG91Notifier) SendOTP(ctx context.Context, phone, code string) error {
	payload := msg91SendOTPRequest{
		TemplateID: n.dltTemplateID,
		Mobile:     phone,
		AuthKey:    n.authKey,
		OTP:        code,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("msg91: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://control.msg91.com/api/v5/otp",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("msg91: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authkey", n.authKey)

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("msg91: send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		slog.Error("msg91: OTP send failed",
			"status", resp.StatusCode,
			"body", string(respBody),
			"phone", phone,
		)
		return fmt.Errorf("msg91: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	slog.Info("msg91: OTP sent successfully", "phone", phone)
	return nil
}

// SendBookingConfirmation sends a booking confirmation SMS via MSG91.
// This is a placeholder — implement when you have a DLT template for confirmations.
func (n *MSG91Notifier) SendBookingConfirmation(ctx context.Context, phone, patientName, date, timeStr string) error {
	slog.Warn("msg91: SendBookingConfirmation not yet implemented — needs a separate DLT template",
		"phone", phone,
		"patient", patientName,
	)
	return nil
}
