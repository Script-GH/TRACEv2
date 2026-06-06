package dto

import "github.com/trace/trace/internal/domain/workout"

// Workout and log DTOs.

type StartLogRequest struct {
	UserID      string `json:"user_id"`
	SessionDate string `json:"session_date"` // "2024-01-15"
}

type StartLogResponse struct {
	SessionID      string `json:"session_id"`
	OpeningMessage string `json:"opening_message"`
}

type LogTurnRequest struct {
	SessionID   string `json:"session_id"`
	UserMessage string `json:"user_message"`
}

type LogTurnResponse struct {
	AgentReply   string `json:"agent_reply"`
	IsComplete   bool   `json:"is_complete"`
	ExtractedLog *workout.Log `json:"extracted_log,omitempty"` // shown in confirmation step
}

type ConfirmLogRequest struct {
	SessionID string `json:"session_id"`
	Confirmed bool   `json:"confirmed"`
}

type ConfirmLogResponse struct {
	LogID                string `json:"log_id"`
	WorkoutGenStarted    bool   `json:"workout_gen_started"`
	EstimatedReadySecs   int    `json:"estimated_ready_secs"` // hint for UI progress indicator
}

type GetWorkoutResponse struct {
	Plan         *workout.Plan `json:"plan"`
	GraphVersion int           `json:"graph_version"`
}
