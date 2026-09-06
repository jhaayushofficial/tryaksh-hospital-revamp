package handler

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
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

func generateReference() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return string(b[:2]) + "-" + string(b[2:])
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

	authPhone, ok := ctx.Value(middleware.PatientPhoneKey).(string)
	if !ok || authPhone == "" {
		response.Unauthorized(w, "patient phone not found in authenticated context")
		return
	}
	// Force the phone to be the one from the authenticated token
	req.PatientPhone = authPhone

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

	ref := generateReference()

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
		// Check for unique constraint violation (double-booking)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			response.Conflict(w, "This slot is already booked")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusCreated, app)
}

func (h *Appointments) GetByReference(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ref := chi.URLParam(r, "reference")
	patientPhone := ctx.Value(middleware.PatientPhoneKey).(string)

	app, err := h.q.GetAppointmentByReference(ctx, ref)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(w, "appointment not found")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	if app.PatientPhone == nil || *app.PatientPhone != patientPhone {
		response.Unauthorized(w, "you don't have permission to access this appointment")
		return
	}

	response.JSON(w, http.StatusOK, app)
}

func (h *Appointments) Cancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ref := chi.URLParam(r, "reference")
	patientPhone := ctx.Value(middleware.PatientPhoneKey).(string)

	app, err := h.q.GetAppointmentByReference(ctx, ref)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(w, "appointment not found")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	if app.PatientPhone == nil || *app.PatientPhone != patientPhone {
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
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

func (h *Appointments) ListMyAppointments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authPhone, ok := ctx.Value(middleware.PatientPhoneKey).(string)
	if !ok || authPhone == "" {
		response.Unauthorized(w, "patient phone not found in authenticated context")
		return
	}

	apps, err := h.q.ListPatientAppointments(ctx, authPhone)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	if apps == nil {
		apps = []db.Appointment{}
	}

	response.JSON(w, http.StatusOK, apps)
}
