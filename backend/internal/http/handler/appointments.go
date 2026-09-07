package handler

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/config"
	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/domain/availability"
	"github.com/tryaksh/clinic/backend/internal/http/middleware"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

type Appointments struct {
	q   *db.Queries
	cfg *config.Config
}

func NewAppointments(pool *pgxpool.Pool, cfg *config.Config) *Appointments {
	return &Appointments{
		q:   db.New(pool),
		cfg: cfg,
	}
}

func (h *Appointments) GetSlots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	doctorIDStr := r.URL.Query().Get("doctor_id")
	clinicIDStr := r.URL.Query().Get("clinic_id")
	dateStr := r.URL.Query().Get("date")

	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		response.BadRequest(w, "invalid doctor_id")
		return
	}
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		response.BadRequest(w, "invalid clinic_id")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.BadRequest(w, "invalid date format (YYYY-MM-DD)")
		return
	}

	// 1. Fetch availability blocks
	availBlocks, err := h.q.GetAvailabilityBlocks(ctx, doctorID, clinicID, date)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Convert to domain blocks, combining with the requested date
	var blocks []availability.Block
	for _, ab := range availBlocks {
		// ab.StartTime is likely 0000-01-01 or 2000-01-01. Combine its hours/mins with 'date'
		y, m, d := date.Date()
		start := time.Date(y, m, d, ab.StartTime.Hour(), ab.StartTime.Minute(), 0, 0, date.Location())
		end := time.Date(y, m, d, ab.EndTime.Hour(), ab.EndTime.Minute(), 0, 0, date.Location())

		var bs, be *time.Time
		if ab.BreakStart != nil && ab.BreakEnd != nil {
			t1 := time.Date(y, m, d, ab.BreakStart.Hour(), ab.BreakStart.Minute(), 0, 0, date.Location())
			t2 := time.Date(y, m, d, ab.BreakEnd.Hour(), ab.BreakEnd.Minute(), 0, 0, date.Location())
			bs = &t1
			be = &t2
		}

		blocks = append(blocks, availability.Block{
			Start:      start,
			End:        end,
			BreakStart: bs,
			BreakEnd:   be,
		})
	}

	// 2. Fetch booked times
	bookedTimesRaw, err := h.q.GetBookedTimes(ctx, doctorID, clinicID, date)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// Combine booked times with the date as well
	var bookedTimes []time.Time
	y, m, d := date.Date()
	for _, bt := range bookedTimesRaw {
		bookedTimes = append(bookedTimes, time.Date(y, m, d, bt.Hour(), bt.Minute(), 0, 0, date.Location()))
	}

	// 3. Merge blocks
	loc, err := time.LoadLocation(h.cfg.Timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	merged := availability.MergeBlocks(blocks, bookedTimes, now, h.cfg.MinLeadTimeMinutes, h.cfg.SlotDurationMinutes)

	type FrontendSlot struct {
		Start  string `json:"start"`
		End    string `json:"end"`
		Status string `json:"status"`
	}

	var flatSlots []FrontendSlot
	for _, b := range merged {
		for _, s := range b.Slots {
			// Convert status
			status := "AVAILABLE"
			if s.Status == availability.StatusBooked {
				status = "BOOKED"
			} else if s.Status == availability.StatusPast {
				status = "PAST"
			}

			flatSlots = append(flatSlots, FrontendSlot{
				Start:  s.Time.Format("15:04"),
				End:    s.Time.Add(time.Duration(h.cfg.SlotDurationMinutes) * time.Minute).Format("15:04"),
				Status: status,
			})
		}
	}

	if flatSlots == nil {
		flatSlots = []FrontendSlot{}
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"slots": flatSlots,
	})
}

// referenceCharset omits I, O, 0 and 1 so a reference read over the phone
// cannot be transcribed ambiguously.
const referenceCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// generateReference returns a human-quotable booking reference, e.g. "AB-CDEFGH".
func generateReference() (string, error) {
	b := make([]byte, 8)
	max := big.NewInt(int64(len(referenceCharset)))
	for i := range b {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("generating booking reference: %w", err)
		}
		b[i] = referenceCharset[num.Int64()]
	}
	return string(b[:2]) + "-" + string(b[2:]), nil
}

func (h *Appointments) Book(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		DoctorID       uuid.UUID `json:"doctor_id"`
		ClinicID       uuid.UUID `json:"clinic_id"`
		Date           string    `json:"appointment_date"`
		Time           string    `json:"start_time"`
		PatientName    string    `json:"patient_name"`
		PatientPhone   string    `json:"patient_phone"`
		PatientEmail   *string   `json:"patient_email"`
		PatientNote    *string   `json:"patient_note"`
		IdempotencyKey *string   `json:"idempotency_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	authPhone, ok := middleware.PatientPhone(ctx)
	if !ok {
		response.Unauthorized(w, "patient phone not found in authenticated context")
		return
	}
	// Force the phone to be the one from the authenticated token
	req.PatientPhone = authPhone

	if strings.TrimSpace(req.PatientName) == "" {
		response.BadRequest(w, "patient_name is required")
		return
	}

	appDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		response.BadRequest(w, "invalid date format (YYYY-MM-DD)")
		return
	}

	startTime, err := time.Parse("15:04", req.Time)
	if err != nil {
		response.BadRequest(w, "invalid time format (HH:MM)")
		return
	}

	endTime := startTime.Add(time.Duration(h.cfg.SlotDurationMinutes) * time.Minute)

	// Note: We are relying on the database partial unique index (uniq_active_slot)
	// to prevent double-booking. If it fails due to constraint violation,
	// it will return a specific error code which we should ideally catch and translate.

	ref, err := generateReference()
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	app, err := h.q.CreateAppointment(ctx, db.CreateAppointmentParams{
		Reference:       ref,
		DoctorID:        req.DoctorID,
		ClinicID:        req.ClinicID,
		AppointmentDate: appDate,
		StartTime:       startTime,
		EndTime:         endTime,
		PatientName:     &req.PatientName,
		PatientPhone:    &req.PatientPhone,
		PatientEmail:    req.PatientEmail,
		PatientNote:     req.PatientNote,
		IsBlock:         false,
		Status:          "BOOKED",
		IdempotencyKey:  req.IdempotencyKey,
	})

	if err != nil {
		// uniq_active_slot rejects a second live booking for the same slot;
		// that is a client-visible conflict, not a server fault.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			slog.WarnContext(ctx, "booking rejected: slot already taken",
				slog.String("doctor_id", req.DoctorID.String()),
				slog.String("clinic_id", req.ClinicID.String()),
				slog.String("date", req.Date),
				slog.String("start_time", req.Time),
			)
			response.Conflict(w, "This slot is already booked")
			return
		}
		response.InternalServerError(w, r, fmt.Errorf("creating appointment: %w", err))
		return
	}

	slog.InfoContext(ctx, "appointment booked",
		slog.String("reference", app.Reference),
		slog.String("doctor_id", req.DoctorID.String()),
		slog.String("clinic_id", req.ClinicID.String()),
		slog.String("date", req.Date),
		slog.String("start_time", req.Time),
	)

	response.JSONCtx(ctx, w, http.StatusCreated, app)
}

func (h *Appointments) GetByReference(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ref := chi.URLParam(r, "reference")

	patientPhone, ok := middleware.PatientPhone(ctx)
	if !ok {
		response.Unauthorized(w, "patient phone not found in authenticated context")
		return
	}

	app, err := h.q.GetAppointmentByReference(ctx, ref)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(w, "appointment not found")
			return
		}
		response.InternalServerError(w, r, fmt.Errorf("loading appointment %q: %w", ref, err))
		return
	}

	if app.PatientPhone == nil || *app.PatientPhone != patientPhone {
		slog.WarnContext(ctx, "appointment access denied", slog.String("reference", ref))
		response.Unauthorized(w, "you don't have permission to access this appointment")
		return
	}

	response.JSONCtx(ctx, w, http.StatusOK, app)
}

func (h *Appointments) Cancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ref := chi.URLParam(r, "reference")

	patientPhone, ok := middleware.PatientPhone(ctx)
	if !ok {
		response.Unauthorized(w, "patient phone not found in authenticated context")
		return
	}

	app, err := h.q.GetAppointmentByReference(ctx, ref)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(w, "appointment not found")
			return
		}
		response.InternalServerError(w, r, fmt.Errorf("loading appointment %q: %w", ref, err))
		return
	}

	if app.PatientPhone == nil || *app.PatientPhone != patientPhone {
		slog.WarnContext(ctx, "appointment cancellation denied", slog.String("reference", ref))
		response.Unauthorized(w, "you don't have permission to modify this appointment")
		return
	}

	if app.Status == "CANCELLED" {
		response.BadRequest(w, "appointment is already cancelled")
		return
	}

	// For patient-facing cancellations, we assume 'PATIENT'.
	// Admin cancellations would use a different endpoint or auth context.
	cancelledBy := "PATIENT"
	var reason *string

	var req struct {
		Reason string `json:"reason"`
	}
	if json.NewDecoder(r.Body).Decode(&req) == nil && req.Reason != "" {
		reason = &req.Reason
	}

	updated, err := h.q.UpdateAppointmentStatus(ctx, app.ID, "CANCELLED", &cancelledBy, reason)
	if err != nil {
		response.InternalServerError(w, r, fmt.Errorf("cancelling appointment %q: %w", ref, err))
		return
	}

	slog.InfoContext(ctx, "appointment cancelled",
		slog.String("reference", ref),
		slog.String("cancelled_by", cancelledBy),
		slog.Bool("reason_given", reason != nil),
	)

	response.JSONCtx(ctx, w, http.StatusOK, updated)
}

func (h *Appointments) ListMyAppointments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authPhone, ok := middleware.PatientPhone(ctx)
	if !ok {
		response.Unauthorized(w, "patient phone not found in authenticated context")
		return
	}

	apps, err := h.q.ListPatientAppointments(ctx, authPhone)
	if err != nil {
		response.InternalServerError(w, r, fmt.Errorf("listing patient appointments: %w", err))
		return
	}

	if apps == nil {
		apps = []db.Appointment{}
	}

	response.JSONCtx(ctx, w, http.StatusOK, apps)
}
