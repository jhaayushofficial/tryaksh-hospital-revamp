package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"github.com/tryaksh/clinic/backend/internal/config"
	"github.com/tryaksh/clinic/backend/internal/http/handler"
	customMiddleware "github.com/tryaksh/clinic/backend/internal/http/middleware"
	"github.com/tryaksh/clinic/backend/internal/http/response"
	"log/slog"
)

func NewRouter(cfg *config.Config, pool *pgxpool.Pool, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(customMiddleware.RequestID)
	r.Use(customMiddleware.Logger(logger))
	r.Use(customMiddleware.Recovery)
	r.Use(customMiddleware.CORS(cfg.ParsedCORSOrigins()))

	r.Get("/healthz", handler.Healthz)
	r.Get("/readyz", handler.Readyz(pool))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/config", handler.Config(cfg))
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.NotFound(w, "Endpoint not found")
	})

	return r
}
