package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/config"
	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

// validAppointmentStatuses mirrors chk_appointment_status in the schema, so an
// unknown status is a 400 rather than a constraint violation surfacing as 500.
var validAppointmentStatuses = map[string]bool{
	"BOOKED":    true,
	"COMPLETED": true,
	"CANCELLED": true,
	"NO_SHOW":   true,
}

type AdminAppointments struct {
	q   *db.Queries
	cfg *config.Config
}

func NewAdminAppointments(pool *pgxpool.Pool, cfg *config.Config) *AdminAppointments {
	return &AdminAppointments{
		q:   db.New(pool),
		cfg: cfg,
	}
}

func (h *AdminAppointments) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := db.ListAppointmentsParams{}

	if doctorID := r.URL.Query().Get("doctor_id"); doctorID != "" {
		if parsed, err := uuid.Parse(doctorID); err == nil {
			params.DoctorID = &parsed
		}
	}

	if clinicID := r.URL.Query().Get("clinic_id"); clinicID != "" {
		if parsed, err := uuid.Parse(clinicID); err == nil {
			params.ClinicID = &parsed
		}
	}

	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			params.AppointmentDate = &parsed
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		params.Status = &status
	}

	appointments, err := h.q.ListAppointments(ctx, params)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	if appointments == nil {
		appointments = []db.Appointment{}
	}

	response.JSON(w, http.StatusOK, appointments)
}

func (h *AdminAppointments) ForceBook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		DoctorID       uuid.UUID `json:"doctor_id"`
		ClinicID       uuid.UUID `json:"clinic_id"`
		Date           string    `json:"date"`
		Time           string    `json:"time"`
		PatientName    *string   `json:"patient_name"`
		PatientPhone   *string   `json:"patient_phone"`
		PatientEmail   *string   `json:"patient_email"`
		PatientNote    *string   `json:"patient_note"`
		IsBlock        bool      `json:"is_block"` // true if receptionist is just blocking the slot
		IdempotencyKey *string   `json:"idempotency_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
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
		PatientName:     req.PatientName,
		PatientPhone:    req.PatientPhone,
		PatientEmail:    req.PatientEmail,
		PatientNote:     req.PatientNote,
		IsBlock:         req.IsBlock,
		Status:          "BOOKED",
		IdempotencyKey:  req.IdempotencyKey,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			response.Conflict(w, "This slot is already booked")
			return
		}
		response.InternalServerError(w, r, fmt.Errorf("force-booking appointment: %w", err))
		return
	}

	slog.InfoContext(ctx, "appointment force-booked by admin",
		slog.String("reference", app.Reference),
		slog.String("doctor_id", req.DoctorID.String()),
		slog.String("clinic_id", req.ClinicID.String()),
		slog.String("date", req.Date),
		slog.String("start_time", req.Time),
		slog.Bool("is_block", req.IsBlock),
	)

	response.JSONCtx(ctx, w, http.StatusCreated, app)
}

func (h *AdminAppointments) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid appointment id")
		return
	}

	var req struct {
		Status          string  `json:"status"`
		CancelledBy     *string `json:"cancelled_by"`
		CancelledReason *string `json:"cancelled_reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	if req.Status == "" {
		response.BadRequest(w, "status is required")
		return
	}

	if !validAppointmentStatuses[req.Status] {
		response.BadRequest(w, "invalid status (must be BOOKED, COMPLETED, CANCELLED or NO_SHOW)")
		return
	}

	app, err := h.q.UpdateAppointmentStatus(ctx, id, req.Status, req.CancelledBy, req.CancelledReason)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(w, "appointment not found")
			return
		}
		response.InternalServerError(w, r, fmt.Errorf("updating appointment status: %w", err))
		return
	}

	slog.InfoContext(ctx, "appointment status updated by admin",
		slog.String("appointment_id", id.String()),
		slog.String("reference", app.Reference),
		slog.String("status", req.Status),
	)

	response.JSONCtx(ctx, w, http.StatusOK, app)
}
