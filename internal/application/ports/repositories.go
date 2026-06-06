package ports

import (
	"context"

	"github.com/trace/trace/internal/domain/graph"
	"github.com/trace/trace/internal/domain/user"
	"github.com/trace/trace/internal/domain/workout"
)

// These interfaces are used ONLY by the MCP server (infrastructure/mcp).
// Agent code never imports these directly — agents use MCPClient.

// CausalGraphRepository persists and retrieves causal graphs.
type CausalGraphRepository interface {
	Get(ctx context.Context, userID string) (*graph.CausalGraph, error)
	Save(ctx context.Context, g *graph.CausalGraph) error // increments version
	GetVersion(ctx context.Context, userID string) (int, error)
}

// UserProfileRepository persists and retrieves user profiles.
// All reads/writes are transparently encrypted/decrypted at the infrastructure layer.
type UserProfileRepository interface {
	Get(ctx context.Context, userID string) (*user.Profile, error)
	Save(ctx context.Context, profile *user.Profile) error
	Exists(ctx context.Context, userID string) (bool, error)
}

// WorkoutLogRepository persists and retrieves workout logs.
type WorkoutLogRepository interface {
	Save(ctx context.Context, log *workout.Log) error
	GetHistory(ctx context.Context, userID string, n int) ([]workout.Log, error)
	GetByID(ctx context.Context, logID string) (*workout.Log, error)
}

// WorkoutPlanRepository persists generated workout plans.
// Versioned — never overwrites an existing plan for the same date.
type WorkoutPlanRepository interface {
	Save(ctx context.Context, plan *workout.Plan) error
	GetForDate(ctx context.Context, userID string, date string) (*workout.Plan, error)
	GetLatest(ctx context.Context, userID string) (*workout.Plan, error)
}

// ExerciseLibraryRepository is a read-only catalogue of exercises.
type ExerciseLibraryRepository interface {
	GetFiltered(ctx context.Context, filter ExerciseFilter) ([]workout.Exercise, error)
	GetByID(ctx context.Context, id string) (*workout.Exercise, error)
	GetSubstitutes(ctx context.Context, exerciseID string, filter ExerciseFilter) ([]workout.Exercise, error)
}

// ProgressionHistoryRepository stores per-exercise weight/rep history.
type ProgressionHistoryRepository interface {
	GetHistory(ctx context.Context, userID string, exercise string) ([]workout.ProgressionEntry, error)
	Save(ctx context.Context, userID string, entry workout.ProgressionEntry) error
}

// VetoEventRepository persists veto records for audit and graph feedback loops.
type VetoEventRepository interface {
	Save(ctx context.Context, event VetoEvent) (string, error)
	GetByUser(ctx context.Context, userID string, limit int) ([]VetoEvent, error)
}

// CoTTraceRepository stores chain-of-thought traces with a 30-day TTL.
type CoTTraceRepository interface {
	Save(ctx context.Context, trace CoTTrace) error
	Get(ctx context.Context, traceID string) (*CoTTrace, error)
}

// ConflictRepository persists irreconcilable agent conflict records.
type ConflictRepository interface {
	Save(ctx context.Context, conflict ConflictRecord) (string, error)
}
