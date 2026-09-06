package handler

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tryaksh/clinic/backend/internal/http/response"
)

func Healthz(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func Readyz(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(context.Background()); err != nil {
			response.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "down"})
			return
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}
