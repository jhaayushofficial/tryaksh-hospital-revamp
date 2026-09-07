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

type Clinics struct {
	q *db.Queries
}

func NewClinics(pool *pgxpool.Pool) *Clinics {
	return &Clinics{
		q: db.New(pool),
	}
}

// Public endpoints

func (h *Clinics) ListActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clinics, err := h.q.ListActiveClinics(ctx)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, clinics)
}

func (h *Clinics) GetBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	clinic, err := h.q.GetClinicBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response.NotFound(w, "clinic not found")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, clinic)
}

// Admin endpoints

func (h *Clinics) ListAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clinics, err := h.q.ListAllClinics(ctx)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, clinics)
}

func (h *Clinics) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req db.CreateClinicParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}

	clinic, err := h.q.CreateClinic(ctx, req)
	if err != nil {
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusCreated, clinic)
}

func (h *Clinics) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid clinic id")
		return
	}

	var req db.UpdateClinicParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid json payload")
		return
	}
	req.ID = id

	clinic, err := h.q.UpdateClinic(ctx, req)
	if err != nil {
		if err.Error() == "no rows in result set" {
			response.NotFound(w, "clinic not found")
			return
		}
		response.InternalServerError(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, clinic)
}
