package metabolic

import (
	"context"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
)

// Agent implements agents.MetabolicAgent.
// Focuses on energy balance and fat loss alignment.
// No LLM dependency.
type Agent struct {
	mcp ports.MCPClient
}

func New(mcp ports.MCPClient) *Agent {
	return &Agent{mcp: mcp}
}

// Evaluate reads the user's goal profile and nutrition scores from recent logs,
// then recommends conditioning work and calorie deficit targets aligned with
// the primary goal (e.g. visible_abs, weight_loss).
func (a *Agent) Evaluate(ctx context.Context, input agents.ExpertInput) (agents.MetabolicReport, error) {
	panic("not implemented")
}
