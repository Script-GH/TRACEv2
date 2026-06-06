package growth

import (
	"context"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
)

// Agent implements agents.GrowthAgent.
// Reads POSITIVE causal paths from the graph to select exercises.
// No LLM, no veto power.
type Agent struct {
	mcp ports.MCPClient
}

func New(mcp ports.MCPClient) *Agent {
	return &Agent{mcp: mcp}
}

// Evaluate queries the causal graph for positive patterns and selects exercises
// that reinforce them. Passes proposed exercise names to InjuryAgent for vetting.
//
// For exercises with no graph history (new movements), uses population-level defaults
// at confidence 0.5, flagged as "learning — will personalise over 3 sessions".
func (a *Agent) Evaluate(ctx context.Context, input agents.ExpertInput) (agents.GrowthReport, error) {
	panic("not implemented")
}
