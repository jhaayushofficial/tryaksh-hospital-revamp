package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

type DoctorClinics struct {
	q *db.Queries
}

func NewDoctorClinics(pool *pgxpool.Pool) *DoctorClinics {
	return &DoctorClinics{
		q: db.New(pool),
	}
}

func (h *DoctorClinics) Link(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req db.LinkDoctorClinicParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	if err := h.q.LinkDoctorClinic(ctx, req); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"status": "linked"})
}

func (h *DoctorClinics) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	doctorIDStr := chi.URLParam(r, "doctor_id")
	clinicIDStr := chi.URLParam(r, "clinic_id")

	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		response.BadRequest(w, "invalid doctor id")
		return
	}
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		response.BadRequest(w, "invalid clinic id")
		return
	}

	dc, err := h.q.GetDoctorClinic(ctx, doctorID, clinicID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			// Not an error, just means they aren't linked yet. Return empty/404.
			response.NotFound(w, "doctor-clinic link not found")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, dc)
}

func (h *DoctorClinics) UpdateHours(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	doctorIDStr := chi.URLParam(r, "doctor_id")
	clinicIDStr := chi.URLParam(r, "clinic_id")

	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		response.BadRequest(w, "invalid doctor id")
		return
	}
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		response.BadRequest(w, "invalid clinic id")
		return
	}

	var req struct {
		DefaultHours json.RawMessage `json:"default_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	dc, err := h.q.UpdateDoctorClinicHours(ctx, db.UpdateDoctorClinicHoursParams{
		DoctorID:     doctorID,
		ClinicID:     clinicID,
		DefaultHours: req.DefaultHours,
	})
	if err != nil {
		if err.Error() == "no rows in result set" {
			response.NotFound(w, "doctor-clinic link not found")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, dc)
}

func (h *DoctorClinics) Unlink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	doctorIDStr := chi.URLParam(r, "doctor_id")
	clinicIDStr := chi.URLParam(r, "clinic_id")

	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		response.BadRequest(w, "invalid doctor id")
		return
	}
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		response.BadRequest(w, "invalid clinic id")
		return
	}

	if err := h.q.UnlinkDoctorClinic(ctx, doctorID, clinicID); err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
