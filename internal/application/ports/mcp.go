package ports

import (
	"context"

	"github.com/trace/trace/internal/domain/graph"
	"github.com/trace/trace/internal/domain/user"
	"github.com/trace/trace/internal/domain/workout"
)

// MCPClient is the ONLY way agents access data.
// Direct repository access from agent code is a build failure.
// This interface maps 1:1 with the 13 MCP tools defined in the spec.
//
// The infrastructure/mcp package implements this interface.
// Tests inject a mock implementation.
type MCPClient interface {
	// Graph tools
	GetCausalGraph(ctx context.Context, userID string, domainFilter string) (*graph.CausalGraph, error)
	UpdateCausalGraph(ctx context.Context, userID string, g *graph.CausalGraph) (newVersion int, err error)
	QueryCausalPaths(ctx context.Context, userID string, query CausalPathQuery) ([]graph.CausalPath, error)

	// Log tools
	GetWorkoutLogHistory(ctx context.Context, userID string, n int) ([]workout.Log, error)
	GetProgressionHistory(ctx context.Context, userID string, exercise string) ([]workout.ProgressionEntry, error)

	// Profile tools
	GetUserProfile(ctx context.Context, userID string) (*user.Profile, error)
	SaveUserProfile(ctx context.Context, profile *user.Profile) error
	GetInjuryConstraints(ctx context.Context, userID string) ([]graph.InjuryConstraint, error)

	// Exercise tools
	GetExerciseLibrary(ctx context.Context, filter ExerciseFilter) ([]workout.Exercise, error)

	// Plan tools
	StoreWorkoutPlan(ctx context.Context, plan *workout.Plan) (planID string, err error)

	// Event tools
	LogVetoEvent(ctx context.Context, event VetoEvent) (vetoID string, err error)
	FlagConflict(ctx context.Context, conflict ConflictRecord) (conflictID string, err error)

	// CoT tools
	StoreCoTTrace(ctx context.Context, trace CoTTrace) (traceID string, err error)
	GetCoTTrace(ctx context.Context, traceID string) (*CoTTrace, error)
}

// CausalPathQuery parameterises a graph path search.
type CausalPathQuery struct {
	FromNodeLabel  string        // optional: start node pattern
	ToNodeLabel    string        // optional: target node pattern (e.g. "knee_pain")
	Polarity       graph.Polarity // optional filter
	MinConfidence  float32
}

// ExerciseFilter parameterises the exercise library query.
type ExerciseFilter struct {
	Equipment      []string
	MuscleGroups   []string
	ExcludeContra  []string // body parts to exclude (from injury constraints)
}

// VetoEvent records an InjuryAgent or RecoveryAgent veto for audit and graph feedback.
type VetoEvent struct {
	UserID          string  `json:"user_id"`
	AgentName       string  `json:"agent_name"`
	ExerciseVetoed  string  `json:"exercise_vetoed"`
	Reason          string  `json:"reason"`
	Confidence      float32 `json:"confidence"`
	GraphVersion    int     `json:"graph_version"`
	SessionDate     string  `json:"session_date"`
}

// ConflictRecord logs irreconcilable agent outputs for PlannerAgent.
type ConflictRecord struct {
	UserID       string   `json:"user_id"`
	AgentNames   []string `json:"agent_names"`
	Description  string   `json:"description"`
	GraphVersion int      `json:"graph_version"`
}

// CoTTrace stores the chain-of-thought reasoning for a workout generation.
type CoTTrace struct {
	TraceID      string         `json:"trace_id"`
	UserID       string         `json:"user_id"`
	GraphVersion int            `json:"graph_version"`
	Steps        []CoTStep      `json:"steps"`
	CreatedAt    int64          `json:"created_at"`
}

// CoTStep is one reasoning step in a chain-of-thought trace.
type CoTStep struct {
	AgentName   string `json:"agent_name"`
	StepLabel   string `json:"step_label"`
	Reasoning   string `json:"reasoning"`
	Confidence  float32 `json:"confidence"`
}
