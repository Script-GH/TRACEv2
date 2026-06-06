package recovery

import (
	"context"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
)

// Agent implements agents.RecoveryAgent.
// No LLM dependency — computes recovery score from logged data.
// Inputs: last 7 logs (sleep, stress, RPE), user's recovery profile baselines.
type Agent struct {
	mcp ports.MCPClient
}

func New(mcp ports.MCPClient) *Agent {
	return &Agent{mcp: mcp}
}

// Evaluate computes the current recovery score (0–1) and intensity cap.
//
// Allostatic load model:
//   - Chronic stress counts against recovery budget
//   - Sub-baseline sleep compounds across consecutive nights
//   - High RPE sessions require proportional recovery time
//   - ForceRestDay = true when recovery score falls below threshold
func (a *Agent) Evaluate(ctx context.Context, input agents.ExpertInput) (agents.RecoveryReport, error) {
	panic("not implemented")
}
