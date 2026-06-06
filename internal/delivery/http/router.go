package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/trace/trace/internal/delivery/http/handlers"
	"github.com/trace/trace/internal/delivery/http/middleware"
	"github.com/trace/trace/pkg/auth"
)

// API route table:
//
//   POST   /api/v1/auth/token                 — issue JWT (not in spec, needed for auth)
//   POST   /api/v1/interview/start            — start onboarding interview session
//   POST   /api/v1/interview/turn             — send one interview turn
//   POST   /api/v1/interview/confirm          — confirm or correct the profile summary
//   POST   /api/v1/log/start                  — start a post-workout log session
//   POST   /api/v1/log/turn                   — send one log turn
//   POST   /api/v1/log/confirm                — confirm extracted workout log
//   GET    /api/v1/workout/today              — fetch today's generated workout
//   GET    /api/v1/workout/stream             — SSE stream for workout generation progress
//   GET    /api/v1/graph                      — fetch full causal graph
//   GET    /api/v1/graph/paths               — query causal paths
//   GET    /api/v1/graph/progress             — fetch Screen 4 progress data

// NewRouter wires all handlers and middleware into the chi router.
func NewRouter(
	jwtSvc *auth.JWTService,
	rateLimitFn func(userID string) (bool, error),
	authHandler *handlers.AuthHandler,
	interview *handlers.InterviewHandler,
	workoutLog *handlers.WorkoutLogHandler,
	workout *handlers.WorkoutHandler,
	graph *handlers.GraphHandler,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(corsMiddleware)

	// Public routes
	r.Post("/api/v1/auth/token", authHandler.Token)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtSvc))
		r.Use(middleware.RateLimit(rateLimitFn))

		// Interview
		r.Post("/api/v1/interview/start", interview.Start)
		r.Post("/api/v1/interview/turn", interview.Turn)
		r.Post("/api/v1/interview/confirm", interview.Confirm)

		// Log
		r.Post("/api/v1/log/start", workoutLog.Start)
		r.Post("/api/v1/log/turn", workoutLog.Turn)
		r.Post("/api/v1/log/confirm", workoutLog.Confirm)

		// Workout
		r.Get("/api/v1/workout/today", workout.GetToday)
		r.Get("/api/v1/workout/stream", workout.Stream)

		// Graph
		r.Get("/api/v1/graph", graph.Get)
		r.Get("/api/v1/graph/paths", graph.QueryPaths)
		r.Get("/api/v1/graph/progress", graph.Progress)
	})

	// Health check (no auth)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
