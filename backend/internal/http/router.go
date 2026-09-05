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

	doctorsHandler := handler.NewDoctors(pool)
	clinicsHandler := handler.NewClinics(pool)
	doctorClinicsHandler := handler.NewDoctorClinics(pool)
	availabilityHandler := handler.NewAvailability(pool)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/config", handler.Config(cfg))
		
		r.Get("/doctors", doctorsHandler.ListActive)
		r.Get("/doctors/{slug}", doctorsHandler.GetBySlug)
		r.Get("/locations", clinicsHandler.ListActive)
		r.Get("/locations/{slug}", clinicsHandler.GetBySlug)

		r.Route("/admin", func(r chi.Router) {
			r.Get("/doctors", doctorsHandler.ListAll)
			r.Post("/doctors", doctorsHandler.Create)
			r.Put("/doctors/{id}", doctorsHandler.Update)

			r.Get("/clinics", clinicsHandler.ListAll)
			r.Post("/clinics", clinicsHandler.Create)
			r.Put("/clinics/{id}", clinicsHandler.Update)

			r.Post("/doctor-clinics", doctorClinicsHandler.Link)
			r.Put("/doctor-clinics/{doctor_id}/{clinic_id}", doctorClinicsHandler.UpdateHours)
			r.Delete("/doctor-clinics/{doctor_id}/{clinic_id}", doctorClinicsHandler.Unlink)

			r.Get("/availability", availabilityHandler.ListRange)
			r.Put("/availability", availabilityHandler.Upsert)
			r.Delete("/availability", availabilityHandler.DeleteDay)
			r.Delete("/availability/{id}", availabilityHandler.DeleteBlock)
			r.Post("/availability/bulk", availabilityHandler.BulkOpen)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.NotFound(w, "Endpoint not found")
	})

	return r
}
