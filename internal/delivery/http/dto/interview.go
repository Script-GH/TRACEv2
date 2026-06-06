package dto

// Interview DTOs — request/response shapes for the interview API.
// Distinct from domain types — DTOs are transport concerns, domain types are business concerns.

type StartInterviewRequest struct {
	UserID string `json:"user_id"`
}

type StartInterviewResponse struct {
	SessionID    string `json:"session_id"`
	OpeningMessage string `json:"opening_message"` // first question from the agent
}

type InterviewTurnRequest struct {
	SessionID   string `json:"session_id"`
	UserMessage string `json:"user_message"`
}

type InterviewTurnResponse struct {
	AgentReply      string           `json:"agent_reply"`
	ExtractedFields []FieldSummary   `json:"extracted_fields,omitempty"` // shown in progress bar
	IsComplete      bool             `json:"is_complete"`
	ProgressPct     int              `json:"progress_pct"` // 0–100 for progress bar
}

// FieldSummary is a redacted field status for the UI progress indicator.
type FieldSummary struct {
	FieldName  string  `json:"field_name"`
	Confidence float32 `json:"confidence"`
	Filled     bool    `json:"filled"`
}

type ConfirmInterviewRequest struct {
	SessionID string `json:"session_id"`
	Confirmed bool   `json:"confirmed"`
	// Corrections maps field names to corrected values if confirmed = false
	Corrections map[string]any `json:"corrections,omitempty"`
}

type ConfirmInterviewResponse struct {
	ProfileSummary string `json:"profile_summary"` // plain-English summary
	GraphBuildStarted bool `json:"graph_build_started"`
}
