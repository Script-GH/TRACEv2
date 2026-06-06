package planner

import (
	"context"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/domain/workout"
)

// Agent implements agents.PlannerAgent.
// Takes all expert reports and synthesises them into a WorkoutPlan.
// Applies MoE weights when resolving conflicts between expert recommendations.
// Persists the plan via MCPClient.StoreWorkoutPlan.
type Agent struct {
	mcp ports.MCPClient
}

func New(mcp ports.MCPClient) *Agent {
	return &Agent{mcp: mcp}
}

// Synthesise assembles the final WorkoutPlan from all expert outputs.
//
// Conflict resolution order:
//  1. RecoveryAgent.ForceRestDay → overrides everything → session_type = "rest"
//  2. InjuryAgent.VetoedExercises → removed from GrowthAgent proposals
//  3. InjuryAgent.FlaggedExercises → volume reduced, InjuryNote populated
//  4. ProgressionReport targets applied to approved exercises
//  5. MetabolicAgent conditioning appended if goal is weight_loss / visible_abs
//  6. Remaining conflicts resolved by MoE weight vector (higher weight = priority)
//
// Calls MCPClient.FlagConflict if any agent outputs remain irreconcilable.
func (a *Agent) Synthesise(ctx context.Context, input agents.PlannerInput) (*workout.Plan, error) {
	panic("not implemented")
}
