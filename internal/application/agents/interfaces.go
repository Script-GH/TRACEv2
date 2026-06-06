package agents

import (
	"context"

	"github.com/trace/trace/internal/domain/user"
	"github.com/trace/trace/internal/domain/workout"
)

// InterviewAgent conducts the onboarding conversation and extracts the user profile.
// Session-scoped: one session per user onboarding.
type InterviewAgent interface {
	// StartSession initialises a new interview session for a new user.
	StartSession(ctx context.Context, userID string) (sessionID string, err error)

	// ProcessTurn receives one user message and returns the agent's reply plus
	// any newly extracted fields. Internally manages conversation history.
	ProcessTurn(ctx context.Context, sessionID string, userMessage string) (InterviewTurnOutput, error)

	// GetExtractedProfile returns the current state of the profile being assembled.
	// Fields with confidence < 0.8 are still present but flagged.
	GetExtractedProfile(ctx context.Context, sessionID string) (*user.Profile, error)

	// IsComplete returns true when all required fields have confidence ≥ 0.8.
	IsComplete(ctx context.Context, sessionID string) (bool, error)

	// Finalise saves the confirmed profile via MCPClient and marks the interview done.
	// Called after the user confirms the INITIAL_SUMMARY.
	Finalise(ctx context.Context, sessionID string) error
}

// LogAgent conducts the post-workout logging conversation and extracts a WorkoutLog.
// Session-scoped: one session per workout day.
type LogAgent interface {
	// StartSession opens a logging session for a given user and date.
	StartSession(ctx context.Context, userID string, sessionDate string) (sessionID string, err error)

	// ProcessTurn receives one user message and returns the agent's reply.
	ProcessTurn(ctx context.Context, sessionID string, userMessage string) (LogTurnOutput, error)

	// GetExtractedLog returns the WorkoutLog being assembled from the conversation.
	GetExtractedLog(ctx context.Context, sessionID string) (*workout.Log, error)

	// Finalise saves the confirmed WorkoutLog via MCPClient.
	// Called after the user confirms the extracted summary.
	// Returns the saved log ID.
	Finalise(ctx context.Context, sessionID string) (logID string, err error)
}

// TwinAgent maintains the personal causal graph.
// It is the ONLY agent that writes to the graph (via MCPClient.UpdateCausalGraph).
type TwinAgent interface {
	// BuildInitial constructs the first causal graph from a completed user profile.
	// Called once immediately after the onboarding interview finishes.
	BuildInitial(ctx context.Context, userID string) (graphVersion int, err error)

	// Update receives a new workout log and updates causal edge confidence scores.
	// Creates new edges for newly observed patterns.
	// Returns the new graph version number.
	Update(ctx context.Context, userID string, logID string) (newVersion int, err error)
}

// MoEGateAgent assigns expert agent weights for a given user's daily context.
// It is a Gemini call — takes the full context, outputs a WeightVector via CompleteJSON.
type MoEGateAgent interface {
	// AssignWeights reads the current context and returns the weight vector.
	// This is the single Gemini call that determines the MoE routing for the day.
	AssignWeights(ctx context.Context, userID string, sessionID string) (WeightVector, error)
}

// RecoveryAgent evaluates the user's current recovery state.
// Veto power: can force a rest day (ForceRestDay = true in report).
type RecoveryAgent interface {
	Evaluate(ctx context.Context, input ExpertInput) (RecoveryReport, error)
}

// InjuryAgent enforces injury constraints against a proposed exercise list.
// Veto power: hard veto on contraindicated exercises.
type InjuryAgent interface {
	// Evaluate takes the proposed exercise list from GrowthAgent and returns
	// a categorised list: vetoed, flagged, and approved.
	Evaluate(ctx context.Context, input ExpertInput, proposedExercises []string) (InjuryReport, error)
}

// GrowthAgent selects exercises based on positive causal patterns.
// No veto power. Proposes; InjuryAgent disposes.
type GrowthAgent interface {
	Evaluate(ctx context.Context, input ExpertInput) (GrowthReport, error)
}

// MetabolicAgent handles energy balance and conditioning recommendations.
type MetabolicAgent interface {
	Evaluate(ctx context.Context, input ExpertInput) (MetabolicReport, error)
}

// ProgressionAgent computes exact weight and rep targets using progression history.
type ProgressionAgent interface {
	Evaluate(ctx context.Context, input ExpertInput, exercises []string) (ProgressionReport, error)
}

// PlannerAgent synthesises all expert reports into the final WorkoutPlan.
// Applies MoE weights when expert recommendations conflict.
type PlannerAgent interface {
	// Synthesise takes all expert outputs plus the weight vector and produces
	// the complete workout. Calls MCPClient.StoreWorkoutPlan on success.
	Synthesise(ctx context.Context, input PlannerInput) (*workout.Plan, error)
}

// PlannerInput bundles all expert reports for synthesis.
type PlannerInput struct {
	UserID       string
	SessionID    string
	GraphVersion int
	Weights      WeightVector
	Recovery     RecoveryReport
	Injury       InjuryReport
	Growth       GrowthReport
	Metabolic    MetabolicReport
	Progression  ProgressionReport
}

// CoachAgent adds plain-language causal reasoning to every exercise in the plan.
// Uses Gemini to generate the CausalReason field on each PlannedExercise.
type CoachAgent interface {
	// Annotate takes the structured plan and fills in CausalReason per exercise.
	// If Gemini is unavailable, returns the plan with empty CausalReason fields
	// and a non-fatal error — the caller shows "reasoning unavailable" in the UI.
	Annotate(ctx context.Context, userID string, plan *workout.Plan) (*workout.Plan, error)
}
