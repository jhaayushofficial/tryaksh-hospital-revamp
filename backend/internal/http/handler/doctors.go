package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/db"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

type Doctors struct {
	q *db.Queries
}

func NewDoctors(pool *pgxpool.Pool) *Doctors {
	return &Doctors{
		q: db.New(pool),
	}
}

// Public endpoints

func (h *Doctors) ListActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	doctors, err := h.q.ListActiveDoctors(ctx)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	// For the public API, we might want to also load their clinics.
	// But let's keep it simple first as requested.
	response.JSON(w, http.StatusOK, doctors)
}

func (h *Doctors) GetBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	doctor, err := h.q.GetDoctorBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(w, "doctor not found")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, doctor)
}

// Admin endpoints

func (h *Doctors) ListAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	doctors, err := h.q.ListAllDoctors(ctx)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, doctors)
}

func (h *Doctors) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req db.CreateDoctorParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	doctor, err := h.q.CreateDoctor(ctx, req)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusCreated, doctor)
}

func (h *Doctors) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid doctor id")
		return
	}

	var req db.UpdateDoctorParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}
	req.ID = id

	doctor, err := h.q.UpdateDoctor(ctx, req)
	if err != nil {
		if err.Error() == "no rows in result set" {
			response.NotFound(w, "doctor not found")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, doctor)
}
