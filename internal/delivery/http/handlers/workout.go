package handlers

import (
	"net/http"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/delivery/http/middleware"
	"github.com/trace/trace/internal/delivery/sse"
	"github.com/trace/trace/internal/domain"
)

// WorkoutHandler handles workout retrieval and SSE streaming.
type WorkoutHandler struct {
	mcp       ports.MCPClient
	sseBroker *sse.Broker
}

func NewWorkoutHandler(mcp ports.MCPClient, sseBroker *sse.Broker) *WorkoutHandler {
	return &WorkoutHandler{mcp: mcp, sseBroker: sseBroker}
}

// GET /api/v1/workout/today
// Returns the generated workout for today, or 404 if not yet generated.
func (h *WorkoutHandler) GetToday(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.mcp.GetUserProfile(r.Context(), userID)
	if err == domain.ErrUserNotFound {
		writeError(w, http.StatusNotFound, "complete the interview first")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !profile.InterviewDone {
		writeError(w, http.StatusNotFound, "complete the interview first")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"status":  "pending",
		"message": "Workout generation is in progress. Subscribe to the stream endpoint for real-time updates.",
		"user_id": userID,
		"goal":    profile.Goals.Primary,
	})
}

// GET /api/v1/workout/stream?session_id=xxx
// Opens an SSE connection for real-time workout generation updates.
func (h *WorkoutHandler) Stream(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}
	h.sseBroker.ServeHTTP(w, r, sessionID)
}

// WorkoutLogHandler handles the post-workout logging flow.
type WorkoutLogHandler struct {
	logAgent agents.LogAgent
}

func NewWorkoutLogHandler(logAgent agents.LogAgent) *WorkoutLogHandler {
	return &WorkoutLogHandler{logAgent: logAgent}
}

// POST /api/v1/log/start
func (h *WorkoutLogHandler) Start(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "coming soon"})
}

// POST /api/v1/log/turn
func (h *WorkoutLogHandler) Turn(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "coming soon"})
}

// POST /api/v1/log/confirm
func (h *WorkoutLogHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "coming soon"})
}
