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

		appointmentsHandler := handler.NewAppointments(pool)
		r.Get("/slots", appointmentsHandler.GetSlots)

		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.RequirePatientAuth(pool))
			r.Get("/appointments", appointmentsHandler.ListMyAppointments)
			r.Post("/appointments", appointmentsHandler.Book)
			r.Get("/appointments/{reference}", appointmentsHandler.GetByReference)
			r.Delete("/appointments/{reference}", appointmentsHandler.Cancel)
		})

		r.Get("/locations", clinicsHandler.ListActive)
		r.Get("/locations/{slug}", clinicsHandler.GetBySlug)

		authHandler := handler.NewAuth(pool)
		r.Post("/auth/request-code", authHandler.RequestCode)
		r.Post("/auth/verify-code", authHandler.VerifyCode)

		adminAuthHandler := handler.NewAdminAuth(cfg)
		r.Post("/admin/login", adminAuthHandler.Login)
		r.Post("/admin/logout", adminAuthHandler.Logout)

		r.Route("/admin", func(r chi.Router) {
			r.Use(customMiddleware.RequireAdminAuth(cfg))

			r.Get("/doctors", doctorsHandler.ListAll)
			r.Post("/doctors", doctorsHandler.Create)
			r.Put("/doctors/{id}", doctorsHandler.Update)

			r.Get("/clinics", clinicsHandler.ListAll)
			r.Post("/clinics", clinicsHandler.Create)
			r.Put("/clinics/{id}", clinicsHandler.Update)

			r.Get("/doctor-clinics/{doctor_id}/{clinic_id}", doctorClinicsHandler.Get)
			r.Post("/doctor-clinics", doctorClinicsHandler.Link)
			r.Put("/doctor-clinics/{doctor_id}/{clinic_id}", doctorClinicsHandler.UpdateHours)
			r.Delete("/doctor-clinics/{doctor_id}/{clinic_id}", doctorClinicsHandler.Unlink)

			r.Get("/availability", availabilityHandler.ListRange)
			r.Put("/availability", availabilityHandler.Upsert)
			r.Delete("/availability", availabilityHandler.DeleteDay)
			r.Delete("/availability/{id}", availabilityHandler.DeleteBlock)
			r.Post("/availability/bulk", availabilityHandler.BulkOpen)

			adminAppsHandler := handler.NewAdminAppointments(pool)
			r.Get("/appointments", adminAppsHandler.List)
			r.Post("/appointments", adminAppsHandler.ForceBook)
			r.Put("/appointments/{id}/status", adminAppsHandler.UpdateStatus)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.NotFound(w, "Endpoint not found")
	})

	return r
}
