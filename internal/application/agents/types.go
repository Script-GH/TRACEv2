package agents

import (
	"github.com/trace/trace/internal/domain/workout"
)

// WeightVector is the MoEGateAgent's output.
// Each field is a 0–1 weight indicating how much influence that expert
// should have over the final workout synthesis.
// Weights must sum to 1.0.
type WeightVector struct {
	Recovery    float32 `json:"recovery"`
	Injury      float32 `json:"injury"`
	Growth      float32 `json:"growth"`
	Metabolic   float32 `json:"metabolic"`
	Progression float32 `json:"progression"`
}

// ExpertInput is the common input all expert agents receive.
// Populated by the orchestrator from MCP data before agents are dispatched.
type ExpertInput struct {
	UserID       string
	SessionID    string
	GraphVersion int
	Weight       float32 // assigned by MoEGateAgent for this expert
}

// RecoveryReport is RecoveryAgent's output.
type RecoveryReport struct {
	RecoveryScore float32 `json:"recovery_score"` // 0–1
	IntensityCap  float32 `json:"intensity_cap"`  // 0–1; 0 = rest day forced
	ForceRestDay  bool    `json:"force_rest_day"`
	Reasoning     string  `json:"reasoning"`
	Confidence    float32 `json:"confidence"`
}

// InjuryReport is InjuryAgent's output.
type InjuryReport struct {
	VetoedExercises  []string                       `json:"vetoed_exercises"`
	Substitutions    map[string]workout.PlannedExercise `json:"substitutions"`   // exercise name → safe alternative
	FlaggedExercises []FlaggedExercise              `json:"flagged_exercises"`
	ApprovedList     []string                       `json:"approved_list"`
	Confidence       float32                        `json:"confidence"`
}

// FlaggedExercise is an exercise InjuryAgent flagged (not vetoed — reduced volume + watch note).
type FlaggedExercise struct {
	Exercise      string  `json:"exercise"`
	VolumeReducePct float32 `json:"volume_reduce_pct"` // e.g. 0.40 = reduce by 40%
	InjuryNote    string  `json:"injury_note"`
	Confidence    float32 `json:"confidence"`
}

// GrowthReport is GrowthAgent's output — exercise selection based on positive causal patterns.
type GrowthReport struct {
	ProposedExercises []workout.PlannedExercise `json:"proposed_exercises"`
	SessionType       workout.SessionType       `json:"session_type"`
	Rationale         string                    `json:"rationale"`
	Confidence        float32                   `json:"confidence"`
}

// MetabolicReport is MetabolicAgent's output — conditioning and energy balance.
type MetabolicReport struct {
	CardioRecommendations []workout.PlannedExercise `json:"cardio_recommendations"`
	CalorieDeficitTarget  float32                   `json:"calorie_deficit_target"`  // kcal
	Rationale             string                    `json:"rationale"`
	Confidence            float32                   `json:"confidence"`
}

// ProgressionReport is ProgressionAgent's output — exact overload targets per exercise.
type ProgressionReport struct {
	Targets    []ProgressionTarget `json:"targets"`
	Confidence float32             `json:"confidence"`
}

// ProgressionTarget gives the exact weight and rep prescription for one exercise.
type ProgressionTarget struct {
	Exercise  string  `json:"exercise"`
	Sets      int     `json:"sets"`
	Reps      string  `json:"reps"`
	WeightKg  float32 `json:"weight_kg"`
	Rationale string  `json:"rationale"` // e.g. "last 3 sessions at 57.5kg RPE 7 → overload to 60kg"
}

// InterviewField represents a single extracted field with confidence score.
type InterviewField struct {
	FieldName  string  `json:"field_name"`
	Value      any     `json:"value"`
	Confidence float32 `json:"confidence"` // < 0.8 triggers a follow-up question
}

// InterviewTurnOutput is what InterviewAgent returns after processing each user turn.
type InterviewTurnOutput struct {
	AgentReply      string           `json:"agent_reply"`
	ExtractedFields []InterviewField `json:"extracted_fields"`
	IsComplete      bool             `json:"is_complete"` // true when all fields ≥ 0.8 confidence
}

// LogTurnOutput is what LogAgent returns after processing each user turn.
type LogTurnOutput struct {
	AgentReply    string `json:"agent_reply"`
	IsComplete    bool   `json:"is_complete"`
	ConfirmedLog  bool   `json:"confirmed_log"` // true after user confirms extracted summary
}
