package injury

import (
	"context"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
)

// confidenceVetoThreshold — if a negative causal edge exceeds this, InjuryAgent vetoes.
const confidenceVetoThreshold = 0.70

// confidenceFlagThreshold — if a negative causal edge exceeds this AND today's pain
// includes the relevant body part, the exercise is flagged (volume -40%).
const confidenceFlagThreshold = 0.50

// Agent implements agents.InjuryAgent.
// No LLM dependency — all logic is graph queries + threshold checks.
// This makes it fast (<5s) and fully deterministic for a given graph state.
type Agent struct {
	mcp ports.MCPClient
}

func New(mcp ports.MCPClient) *Agent {
	return &Agent{mcp: mcp}
}

// Evaluate checks each proposed exercise against:
//  1. Causal graph negative edges (confidence > vetoThreshold + injury status = active/chronic → VETO)
//  2. Causal graph negative edges (confidence > flagThreshold + today's pain signal → FLAG)
//  3. Everything else → APPROVE
//
// Returns the categorised InjuryReport. Vetoed exercises get substitutions from the exercise library.
func (a *Agent) Evaluate(ctx context.Context, input agents.ExpertInput, proposedExercises []string) (agents.InjuryReport, error) {
	panic("not implemented")
}
