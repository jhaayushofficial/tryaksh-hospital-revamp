package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/domain/availability"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

type Availability struct {
	q *db.Queries
}

func NewAvailability(pool *pgxpool.Pool) *Availability {
	return &Availability{
		q: db.New(pool),
	}
}

func (h *Availability) ListRange(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	doctorIDStr := r.URL.Query().Get("doctor_id")
	clinicIDStr := r.URL.Query().Get("clinic_id")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

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

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		response.BadRequest(w, "invalid from date format")
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		response.BadRequest(w, "invalid to date format")
		return
	}

	blocks, err := h.q.ListAvailabilityRange(ctx, doctorID, clinicID, from, to)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, blocks)
}

func (h *Availability) Upsert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		ID            *uuid.UUID `json:"id,omitempty"`
		DoctorID      uuid.UUID  `json:"doctor_id"`
		ClinicID      uuid.UUID  `json:"clinic_id"`
		AvailableDate string     `json:"available_date"`
		StartTime     string     `json:"start_time"`
		EndTime       string     `json:"end_time"`
		BreakStart    *string    `json:"break_start,omitempty"`
		BreakEnd      *string    `json:"break_end,omitempty"`
		IsOpen        bool       `json:"is_open"`
		Note          *string    `json:"note,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	availDate, err := time.Parse("2006-01-02", req.AvailableDate)
	if err != nil {
		response.BadRequest(w, "invalid date format (YYYY-MM-DD)")
		return
	}

	parseTime := func(tStr string) (time.Time, error) {
		return time.Parse("15:04", tStr)
	}

	start, err := parseTime(req.StartTime)
	if err != nil {
		response.BadRequest(w, "invalid start_time format (HH:MM)")
		return
	}
	end, err := parseTime(req.EndTime)
	if err != nil {
		response.BadRequest(w, "invalid end_time format (HH:MM)")
		return
	}
	if start.After(end) || start.Equal(end) {
		response.BadRequest(w, "start_time must be before end_time")
		return
	}

	var breakStart, breakEnd *time.Time
	if req.BreakStart != nil && req.BreakEnd != nil {
		bs, err := parseTime(*req.BreakStart)
		if err != nil {
			response.BadRequest(w, "invalid break_start format (HH:MM)")
			return
		}
		be, err := parseTime(*req.BreakEnd)
		if err != nil {
			response.BadRequest(w, "invalid break_end format (HH:MM)")
			return
		}
		if bs.After(be) || bs.Equal(be) {
			response.BadRequest(w, "break_start must be before break_end")
			return
		}
		if bs.Before(start) || be.After(end) {
			response.BadRequest(w, "break must be within start and end time")
			return
		}
		breakStart = &bs
		breakEnd = &be
	} else if req.BreakStart != nil || req.BreakEnd != nil {
		response.BadRequest(w, "both break_start and break_end must be provided, or neither")
		return
	}

	// Fetch existing blocks to validate overlap
	existing, err := h.q.GetAvailabilityBlocks(ctx, req.DoctorID, req.ClinicID, availDate)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	newBlock := db.Availability{
		DoctorID:      req.DoctorID,
		ClinicID:      req.ClinicID,
		AvailableDate: availDate,
		StartTime:     start,
		EndTime:       end,
		BreakStart:    breakStart,
		BreakEnd:      breakEnd,
		IsOpen:        req.IsOpen,
		Note:          req.Note,
	}

	if req.ID != nil {
		newBlock.ID = *req.ID
	}

	if err := availability.CheckOverlap(existing, newBlock); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if req.ID != nil {
		// Update
		upd, err := h.q.UpdateAvailability(ctx, db.UpdateAvailabilityParams{
			ID:         *req.ID,
			StartTime:  &start,
			EndTime:    &end,
			BreakStart: breakStart,
			BreakEnd:   breakEnd,
			IsOpen:     &req.IsOpen,
			Note:       req.Note,
		})
		if err != nil {
			response.InternalServerError(w, r, err)
			return
		}
		response.JSON(w, http.StatusOK, upd)
	} else {
		// Create
		cre, err := h.q.CreateAvailability(ctx, db.CreateAvailabilityParams{
			DoctorID:      req.DoctorID,
			ClinicID:      req.ClinicID,
			AvailableDate: availDate,
			StartTime:     start,
			EndTime:       end,
			BreakStart:    breakStart,
			BreakEnd:      breakEnd,
			IsOpen:        req.IsOpen,
			Note:          req.Note,
		})
		if err != nil {
			response.InternalServerError(w, r, err)
			return
		}
		response.JSON(w, http.StatusCreated, cre)
	}
}

func (h *Availability) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.BadRequest(w, "missing id param")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}

	if err := h.q.DeleteAvailability(ctx, id); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Availability) DeleteDay(w http.ResponseWriter, r *http.Request) {
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
		response.BadRequest(w, "invalid date")
		return
	}

	if err := h.q.DeleteAvailabilityByDate(ctx, doctorID, clinicID, date); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type WeekSchedule map[string][]struct {
	Start      string `json:"start"`
	End        string `json:"end"`
	BreakStart string `json:"break_start,omitempty"`
	BreakEnd   string `json:"break_end,omitempty"`
}

func (h *Availability) BulkOpen(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		DoctorID uuid.UUID `json:"doctor_id"`
		ClinicID uuid.UUID `json:"clinic_id"`
		From     string    `json:"from"`
		To       string    `json:"to"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	from, err := time.Parse("2006-01-02", req.From)
	if err != nil || from.IsZero() {
		response.BadRequest(w, "invalid from date")
		return
	}
	to, err := time.Parse("2006-01-02", req.To)
	if err != nil || to.Before(from) {
		response.BadRequest(w, "invalid to date")
		return
	}

	dc, err := h.q.GetDoctorClinic(ctx, req.DoctorID, req.ClinicID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			response.BadRequest(w, "doctor is not linked to this clinic")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	if len(dc.DefaultHours) == 0 {
		response.BadRequest(w, "no default hours set for this doctor-clinic link")
		return
	}

	var schedule WeekSchedule
	if err := json.Unmarshal(dc.DefaultHours, &schedule); err != nil {
		response.InternalServerError(w, r, err) // Corrupt DB data
		return
	}

	createdCount := 0

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		weekdayStr := strings.ToLower(d.Weekday().String())
		blocks, ok := schedule[weekdayStr]
		if !ok || len(blocks) == 0 {
			continue // No hours for this day
		}

		// Skip if any blocks already exist on this date for safety
		existing, err := h.q.GetAvailabilityBlocks(ctx, req.DoctorID, req.ClinicID, d)
		if err != nil {
			response.InternalServerError(w, r, err)
			return
		}
		if len(existing) > 0 {
			continue // Day already configured
		}

		for _, b := range blocks {
			start, _ := time.Parse("15:04", b.Start)
			end, _ := time.Parse("15:04", b.End)
			
			var bs, be *time.Time
			if b.BreakStart != "" && b.BreakEnd != "" {
				bsParsed, _ := time.Parse("15:04", b.BreakStart)
				beParsed, _ := time.Parse("15:04", b.BreakEnd)
				bs = &bsParsed
				be = &beParsed
			}

			_, err := h.q.CreateAvailability(ctx, db.CreateAvailabilityParams{
				DoctorID:      req.DoctorID,
				ClinicID:      req.ClinicID,
				AvailableDate: d,
				StartTime:     start,
				EndTime:       end,
				BreakStart:    bs,
				BreakEnd:      be,
				IsOpen:        true,
			})
			if err != nil {
				// We can just log and ignore, or fail entirely. Simple approach: fail
				response.InternalServerError(w, r, err)
				return
			}
			createdCount++
		}
	}

	response.JSON(w, http.StatusCreated, map[string]int{"created_blocks": createdCount})
}
