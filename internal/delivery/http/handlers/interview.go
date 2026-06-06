package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/delivery/http/dto"
	"github.com/trace/trace/internal/delivery/http/middleware"
)

// InterviewHandler handles the onboarding interview HTTP endpoints.
// It does NOT use the state machine — the interview is self-contained,
// with completeness determined by the agent's field confidence scores.
type InterviewHandler struct {
	interview agents.InterviewAgent
	twin      agents.TwinAgent
}

func NewInterviewHandler(interview agents.InterviewAgent, twin agents.TwinAgent) *InterviewHandler {
	return &InterviewHandler{interview: interview, twin: twin}
}

// POST /api/v1/interview/start
// Creates a new interview session. Returns the session ID and the opening question.
func (h *InterviewHandler) Start(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID, err := h.interview.StartSession(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, dto.StartInterviewResponse{
		SessionID:      sessionID,
		OpeningMessage: "Hi! I'm TRACE. I'm going to build a personal causal model of how your body responds to training. Let's start — what's your main fitness goal right now?",
	})
}

// POST /api/v1/interview/turn
// Processes one conversation turn and returns the agent's reply with field progress.
func (h *InterviewHandler) Turn(w http.ResponseWriter, r *http.Request) {
	var req dto.InterviewTurnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SessionID == "" || req.UserMessage == "" {
		writeError(w, http.StatusBadRequest, "session_id and user_message required")
		return
	}

	result, err := h.interview.ProcessTurn(r.Context(), req.SessionID, req.UserMessage)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := dto.InterviewTurnResponse{
		AgentReply:  result.AgentReply,
		IsComplete:  result.IsComplete,
		ProgressPct: computeProgressPct(result.ExtractedFields),
	}
	for _, f := range result.ExtractedFields {
		resp.ExtractedFields = append(resp.ExtractedFields, dto.FieldSummary{
			FieldName:  f.FieldName,
			Confidence: f.Confidence,
			Filled:     f.Confidence >= 0.8,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// POST /api/v1/interview/confirm
// User confirms the extracted profile summary.
// On confirm=true: saves the profile and triggers initial graph build.
func (h *InterviewHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.ConfirmInterviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !req.Confirmed {
		// User wants to correct something — just acknowledge for now.
		writeJSON(w, http.StatusOK, dto.ConfirmInterviewResponse{
			ProfileSummary:    "Got it — please continue the interview to clarify.",
			GraphBuildStarted: false,
		})
		return
	}

	// Save the profile.
	if err := h.interview.Finalise(r.Context(), req.SessionID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save profile: "+err.Error())
		return
	}

	// Build the initial causal graph synchronously for the demo.
	graphVersion, err := h.twin.BuildInitial(r.Context(), userID)
	if err != nil {
		// Graph build failure is non-fatal — return success with a note.
		writeJSON(w, http.StatusOK, dto.ConfirmInterviewResponse{
			ProfileSummary:    "Profile saved. Graph build encountered an issue: " + err.Error(),
			GraphBuildStarted: false,
		})
		return
	}

	writeJSON(w, http.StatusOK, dto.ConfirmInterviewResponse{
		ProfileSummary:    buildProfileSummary(graphVersion),
		GraphBuildStarted: true,
	})
}

func buildProfileSummary(graphVersion int) string {
	return strings.Join([]string{
		"Your profile has been saved and your personal causal graph (v" + itoa(graphVersion) + ") has been built.",
		"TRACE will now use this graph to generate workouts and explain every decision.",
		"Head to the Graph tab to explore your causal model.",
	}, " ")
}

func computeProgressPct(fields []agents.InterviewField) int {
	if len(fields) == 0 {
		return 0
	}
	var total float32
	for _, f := range fields {
		if f.Confidence > 1 {
			total += 1
		} else {
			total += f.Confidence
		}
	}
	pct := int(total / float32(len(fields)) * 100)
	if pct > 100 {
		return 100
	}
	return pct
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
