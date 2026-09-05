package handler

import (
	"net/http"

	"github.com/tryaksh/clinic/backend/internal/config"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

func Config(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"clinic_name":               cfg.ClinicName,
			"clinic_tagline":            cfg.ClinicTagline,
			"consultation_fee_display":  cfg.ConsultationFeeDisplay,
			"slot_duration_minutes":     cfg.SlotDurationMinutes,
			"booking_window_days":       cfg.BookingWindowDays,
			"cancellation_window_hours": cfg.CancellationWindowHours,
		}
		response.JSON(w, http.StatusOK, data)
	}
}
