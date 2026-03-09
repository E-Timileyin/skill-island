package routes

import (
	"github.com/E-Timileyin/skill-island/services/api/internal/api/handlers"
	"github.com/E-Timileyin/skill-island/services/api/internal/auth"
	"github.com/E-Timileyin/skill-island/services/api/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// SetupRouter initializes the application routing logic.
func SetupRouter(h *handlers.Handler, wsHandler *handlers.WSHandler, cfg config.Config) *chi.Mux {
	r := chi.NewRouter()
	// CORS must be first
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Get("/health", h.Health)

	// Auth routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/logout", h.Logout)
		r.Post("/refresh", h.Refresh)

		// Authenticated auth routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))
			r.Get("/me", h.Me)
		})
	})

	// Authenticated API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))
		r.Post("/sessions/init", h.InitSession)
		r.Get("/sessions/manifest", h.GetManifest)
		r.Post("/sessions", h.SubmitSession)
		r.Post("/sessions/coop", h.SubmitCoopSession)
		r.Get("/analytics/overview", h.AnalyticsOverview)
	})

	// WebSocket route
	r.Get("/ws/game", wsHandler.ServeWS)

	// Profile routes
	r.Route("/api/profiles", func(r chi.Router) {
		r.Use(auth.Middleware(cfg.JWTSecret))
		r.Post("/", h.CreateProfile)
		r.Get("/me", h.GetProfile)
		r.Patch("/me", h.UpdateProfile)
	})

	return r
}
