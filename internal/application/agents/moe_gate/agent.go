package moegate

import (
	"context"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
)

// Agent implements agents.MoEGateAgent.
// It calls Gemini (CompleteJSON) with the full daily context and receives a WeightVector.
// This is the most LLM-dependent agent — the weight vector is computed via structured generation.
type Agent struct {
	mcp ports.MCPClient
	llm ports.LLMClient
}

func New(mcp ports.MCPClient, llm ports.LLMClient) *Agent {
	return &Agent{mcp: mcp, llm: llm}
}

// AssignWeights assembles the daily context (graph summary, last 7 logs, goal profile,
// recovery signals), sends it to Gemini as a structured prompt, and returns the
// parsed WeightVector. Weights are normalised to sum to 1.0 after parsing.
func (a *Agent) AssignWeights(ctx context.Context, userID string, sessionID string) (agents.WeightVector, error) {
	panic("not implemented")
}
