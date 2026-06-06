package workout

// SessionType classifies the nature of a generated workout session.
type SessionType string

const (
	SessionTypeStrength       SessionType = "strength"
	SessionTypeHypertrophy    SessionType = "hypertrophy"
	SessionTypeConditioning   SessionType = "conditioning"
	SessionTypeActiveRecovery SessionType = "active_recovery"
	SessionTypeRest           SessionType = "rest"
)

// PlannedExercise is a single exercise in a generated workout.
// CausalReason is mandatory — must cite the graph path that justified this exercise.
// Substitution is populated by InjuryAgent if the primary exercise was flagged.
type PlannedExercise struct {
	Exercise     string           `json:"exercise"`
	Sets         int              `json:"sets"`
	Reps         string           `json:"reps"`
	WeightKg     float32          `json:"weight_kg"`
	RestSeconds  int              `json:"rest_seconds"`
	Tempo        string           `json:"tempo"`         // "3-1-2" (eccentric-pause-concentric)
	CausalReason string           `json:"causal_reason"` // why this exercise for this user today
	InjuryNote   string           `json:"injury_note"`   // nullable — set by InjuryAgent
	Substitution *PlannedExercise `json:"substitution"`  // nullable — safe alternative
}

// Plan is the complete generated workout for a single day.
// GraphVersion records which causal graph state generated this plan — enables diff-based explanations.
type Plan struct {
	PlanID       string            `json:"plan_id"`
	UserID       string            `json:"user_id"`
	ForDate      string            `json:"for_date"`
	GraphVersion int               `json:"graph_version"` // which graph version generated this
	SessionType  SessionType       `json:"session_type"`
	GoalToday    string            `json:"goal_today"`
	EstDuration  int               `json:"est_duration_min"`
	ExpectedRPE  string            `json:"expected_rpe"` // "6–7"
	WhyToday     string            `json:"why_today"`    // causal reasoning for session type choice
	WarmUp       []PlannedExercise `json:"warm_up"`
	Exercises    []PlannedExercise `json:"exercises"`
	CoolDown     string            `json:"cool_down"`
	CoTTraceID   string            `json:"cot_trace_id"` // chain-of-thought reference
	CreatedAt    int64             `json:"created_at"`
}
