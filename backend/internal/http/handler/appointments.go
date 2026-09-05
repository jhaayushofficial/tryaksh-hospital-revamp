package handler

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/domain/availability"
	"github.com/tryaksh/clinic/backend/internal/http/middleware"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

type Appointments struct {
	q *db.Queries
}

func NewAppointments(pool *pgxpool.Pool) *Appointments {
	return &Appointments{
		q: db.New(pool),
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

	// Convert to domain blocks
	var blocks []availability.Block
	for _, ab := range availBlocks {
		blocks = append(blocks, availability.Block{
			Start:      ab.StartTime,
			End:        ab.EndTime,
			BreakStart: ab.BreakStart,
			BreakEnd:   ab.BreakEnd,
		})
	}

	// 2. Fetch booked times
	bookedTimes, err := h.q.GetBookedTimes(ctx, doctorID, clinicID, date)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// 3. Merge blocks
	// Hardcoded: 15 mins appointment duration, 60 mins lead time (no booking within 60 mins)
	now := time.Now() 
	merged := availability.MergeBlocks(blocks, bookedTimes, now, 60, 15)

	response.JSON(w, http.StatusOK, merged)
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
		Date           string    `json:"date"`
		Time           string    `json:"time"`
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

	endTime := startTime.Add(15 * time.Minute)

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
		// Basic error handling for unique violation
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"uniq_active_slot\" (SQLSTATE 23505)" {
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

	app, err := h.q.GetAppointmentByReference(ctx, ref)
	if err != nil {
		if err.Error() == "no rows in result set" {
			response.NotFound(w, "appointment not found")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, app)
}

func (h *Appointments) Cancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ref := chi.URLParam(r, "reference")

	app, err := h.q.GetAppointmentByReference(ctx, ref)
	if err != nil {
		if err.Error() == "no rows in result set" {
			response.NotFound(w, "appointment not found")
			return
		}
		response.InternalServerError(w, r, err)
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
