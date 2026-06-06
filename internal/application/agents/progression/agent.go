package progression

import (
	"context"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
)

// Agent implements agents.ProgressionAgent.
// Uses the per-exercise weight/rep history from the last 7 sessions to compute
// the exact overload target for each exercise in the proposed list.
// Purely computational — no LLM.
type Agent struct {
	mcp ports.MCPClient
}

func New(mcp ports.MCPClient) *Agent {
	return &Agent{mcp: mcp}
}

// Evaluate looks up progression history for each exercise and applies overload logic.
// The output gives the PlannerAgent exact Sets × Reps × WeightKg for every exercise.
//
// If a user has missed 5+ days (spec requirement): reduce load by 15% for reintroduction.
func (a *Agent) Evaluate(ctx context.Context, input agents.ExpertInput, exercises []string) (agents.ProgressionReport, error) {
	panic("not implemented")
}
